package estelle

import (
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type GarbageCollector struct {
	cacheDir  *CacheDir
	limit     int64
	highLimit int64 // cache-limit * high-ratio
	lowLimit  int64 // cache-limit * low-ratio
	used      int64 // atomic
	mu        sync.Mutex
	isGCing   bool
}

func NewGarbageCollector(cd *CacheDir, limit int64, highRatio, lowRatio float64) *GarbageCollector {
	gc := &GarbageCollector{
		cacheDir:  cd,
		limit:     limit,
		highLimit: int64(float64(limit) * highRatio),
		lowLimit:  int64(float64(limit) * lowRatio),
	}
	// Asynchronous startup scan
	go gc.ScanUsage()
	return gc
}

func (gc *GarbageCollector) Track(delta int64) {
	newUsed := atomic.AddInt64(&gc.used, delta)
	if delta > 0 && newUsed > gc.highLimit {
		gc.MaybeGC()
	}
}

func (gc *GarbageCollector) ScanUsage() {
	var total int64
	filepath.Walk(gc.cacheDir.dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	atomic.StoreInt64(&gc.used, total)
}

func (gc *GarbageCollector) MaybeGC() {
	gc.mu.Lock()
	if gc.isGCing {
		gc.mu.Unlock()
		return
	}
	gc.isGCing = true
	gc.mu.Unlock()

	go gc.RunGC()
}

func (gc *GarbageCollector) RunGC() {
	defer func() {
		gc.mu.Lock()
		gc.isGCing = false
		gc.mu.Unlock()
	}()

	// Random Sampling LRU
	// 1. Pick random subdirectories
	// 2. Scan files in them
	// 3. Delete oldest accessed until usage < lowLimit

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for atomic.LoadInt64(&gc.used) > gc.lowLimit {
		// Randomly select a shard: root/head2/next2
		// head2: 00-ff
		// next2: 00-ff
		
		// To be efficient, we might try to pick existing directories.
		// Since we don't track directory list, we can just try random hex.
		// But scanning empty dirs is wasteful.
		// Alternatively, ReadDir of root, pick random, ReadDir of that, pick random.
		
		removed := gc.evictOneBatch(rng)
		if removed == 0 {
			// Could not remove anything or empty cache?
			// Avoid infinite loop if cache is small but "used" is high (inconsistency?)
			// Or if all files are locked?
			// Sleep a bit to avoid CPU spin
			time.Sleep(100 * time.Millisecond)
			
			// If used is still high but we can't find files, maybe re-scan?
			if atomic.LoadInt64(&gc.used) > gc.lowLimit {
                 // Resync used count just in case
                 gc.ScanUsage()
                 // If still high and no eviction, maybe just break to avoid hang
                 if atomic.LoadInt64(&gc.used) <= gc.lowLimit {
                     break
                 }
                 // Force break to prevent deadlock if we really can't delete
                 break
			}
		}
	}
}

func (gc *GarbageCollector) evictOneBatch(rng *rand.Rand) int64 {
	// Simple strategy: List root dirs, pick one. List its subdirs, pick one. List files, pick oldest.
	root, err := os.Open(gc.cacheDir.dir)
	if err != nil {
		return 0
	}
	defer root.Close()
	
	head2s, err := root.Readdirnames(-1)
	if err != nil || len(head2s) == 0 {
		return 0
	}
	
	// Pick random head2
	h2 := head2s[rng.Intn(len(head2s))]
	h2Path := filepath.Join(gc.cacheDir.dir, h2)
	
	h2Dir, err := os.Open(h2Path)
	if err != nil {
		return 0
	}
	defer h2Dir.Close()
	
	next2s, err := h2Dir.Readdirnames(-1)
	if err != nil || len(next2s) == 0 {
		return 0
	}
	
	// Pick random next2
	n2 := next2s[rng.Intn(len(next2s))]
	n2Path := filepath.Join(h2Path, n2)
	
	// Read files
	dir, err := os.Open(n2Path)
	if err != nil {
		return 0
	}
	defer dir.Close()
	
	files, err := dir.Readdir(-1)
	if err != nil || len(files) == 0 {
		return 0
	}
	
	// Find oldest Atime (using ModTime as proxy if Atime not available, 
	// BUT design says Atime. 
	// Go os.FileInfo doesn't expose Atime directly in platform-independent way.
	// However, estelle.go and design implies checking Atime. 
	// Since we are moving to platform-independent (or Linux target), 
	// we will try to use AccessTime if possible, or fallback to ModTime?
	// The design says: "OSの relatime / noatime 設定に依存せずLRUを機能させるため、アプリケーション側でAtimeを管理する。"
	// "キャッシュヒット時... os.Chtimes を実行してAtimeを更新する" -> This updates ModTime AND AccessTime.
	// So ModTime might be updated too if we use Chtimes(now, now).
	// BUT if file content doesn't change, we shouldn't change ModTime ideally?
	// Chtimes changes both.
	// If we use Chtimes to update Atime, ModTime also updates usually unless we preserve it.
	// But `os.Chtimes` takes `atime` and `mtime`. We can set `mtime` to old value!
	
	// So, we rely on `Atime`.
	// For now, let's pick the file with oldest ModTime or Atime if available.
	// Since I want to compile on Windows (dev env), Atime access is tricky without syscall.
	// Design says: "Most implementations use ModTime if Atime not reliable" - wait, no.
	// Design says: "Approximated LRU... 最も Atime (最終アクセス時刻) が古いファイルを削除する"
	
	// I will just use ModTime for now for portability, and assume Chtimes updates ModTime (or I maintain ModTime and update Atime).
	// If I use `os.Chtimes(now, now)`, ModTime updates.
	// Then ModTime == Last Access Time (mostly).
	// This simplifies things. I will use ModTime.
	
	var oldest os.FileInfo
	for _, fi := range files {
		if !fi.IsDir() {
			if oldest == nil || fi.ModTime().Before(oldest.ModTime()) {
				oldest = fi
			}
		}
	}
	
	if oldest != nil {
		path := filepath.Join(n2Path, oldest.Name())
		size := oldest.Size()
		err := os.Remove(path)
		if err == nil {
			gc.Track(-size)
			return size
		}
	}
	return 0
}
