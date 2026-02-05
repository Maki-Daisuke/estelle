package estelle

import (
	"context"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type GarbageCollector struct {
	dir       string
	limit     int64
	highLimit int64 // cache-limit * high-ratio
	lowLimit  int64 // cache-limit * low-ratio
	used      int64 // atomic
	gcSignal  chan struct{}
}

func NewGarbageCollector(ctx context.Context, dir string, limit int64, highRatio, lowRatio float64) *GarbageCollector {
	gc := &GarbageCollector{
		dir:       dir,
		limit:     limit,
		highLimit: int64(float64(limit) * highRatio),
		lowLimit:  int64(float64(limit) * lowRatio),
		gcSignal:  make(chan struct{}, 1),
	}
	// Asynchronous startup scan
	go gc.Run(ctx)
	return gc
}

func (gc *GarbageCollector) Track(delta int64) {
	atomic.AddInt64(&gc.used, delta)
	select {
	case gc.gcSignal <- struct{}{}:
	default:
		// GC is already running
	}
}

func (gc *GarbageCollector) Run(ctx context.Context) {
	gc.initialScan(ctx)
	for { // Wait for GC signal or context cancellation
		select {
		case <-ctx.Done():
			return
		case <-gc.gcSignal:
			if atomic.LoadInt64(&gc.used) > gc.highLimit {
				gc.RunGC(ctx)
			}
		}
	}
}

func (gc *GarbageCollector) initialScan(ctx context.Context) {
	var total int64
	filepath.WalkDir(gc.dir, func(path string, de fs.DirEntry, err error) error {
		select { // Check if the context is canceled
		case <-ctx.Done():
			return fs.SkipAll
		default:
		}
		if err == nil && de.Type().IsRegular() {
			info, _ := de.Info()
			total += info.Size()
		}
		return nil
	})
	gc.Track(total) // This kicks the GC if needed
}

func (gc *GarbageCollector) RunGC(ctx context.Context) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for atomic.LoadInt64(&gc.used) > gc.lowLimit {
		select {
		case <-ctx.Done():
			return
		default:
			removed := gc.evictOneBatch(rng)
			if removed == 0 {
				// Could not remove anything or empty cache?
				// Avoid infinite loop if cache is small but "used" is high (inconsistency?)
				// Or if all files are locked?
				// Sleep a bit to avoid CPU spin
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (gc *GarbageCollector) evictOneBatch(rng *rand.Rand) int64 {
	// Random Sampling LRU
	// 1. Pick a random subdirectory
	// 2. Scan files in it
	// 3. Delete the oldest accessed file

	// head2: 00-ff
	// next2: 00-ff

	// To be efficient, we try to pick existing directories.
	// Since we don't track directory list, we can just try random hex.
	// But scanning empty dirs is wasteful.
	// Alternatively, ReadDir of root, pick one, ReadDir of that, pick one.
	root, err := os.Open(gc.dir)
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
	h2Path := filepath.Join(gc.dir, h2)

	h2Dir, err := os.Open(h2Path)
	if err != nil {
		return 0
	}
	defer h2Dir.Close()

	next2s, err := h2Dir.Readdirnames(-1)
	if err != nil {
		log.Println("failed to read directory:", err)
		return 0
	}
	if len(next2s) == 0 {
		os.Remove(h2Path)
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

	entries, err := dir.ReadDir(-1)
	if err != nil {
		log.Println("failed to read directory:", err)
		return 0
	}
	if len(entries) == 0 {
		os.Remove(n2Path)
		return 0
	}

	oldest := entries[0]
	oldestInfo, _ := oldest.Info()
	oldestTime := GetAtime(oldestInfo)
	for _, de := range entries[1:] {
		if de.Type().IsRegular() {
			fi, _ := de.Info()
			t := GetAtime(fi)
			if t.Before(oldestTime) {
				oldest = de
				oldestTime = t
			}
		}
	}

	path := filepath.Join(n2Path, oldest.Name())
	fi, err := oldest.Info()
	if err != nil {
		log.Printf("failed to get file info of %s: %v", path, err)
		return 0
	}
	size := fi.Size()
	err = os.Remove(path)
	if err != nil {
		log.Printf("failed to remove %s: %v", path, err)
		return 0
	}
	atomic.AddInt64(&gc.used, -size)

	if len(entries) == 1 {
		err = os.Remove(n2Path)
		if err != nil {
			log.Printf("failed to remove %s: %v", n2Path, err)
		}
	}
	return size
}
