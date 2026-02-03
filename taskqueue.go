package estelle

import (
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/Maki-Daisuke/qlose"
)

type MaybeError struct {
	e error
	c chan struct{}
}

func newMaybeError() *MaybeError {
	return &MaybeError{
		c: make(chan struct{}),
	}
}

func (me *MaybeError) signal(e error) {
	me.e = e
	close(me.c)
}

func (me *MaybeError) Wait() error {
	<-me.c
	return me.e
}

type ThumbnailQueue struct {
	cacheDir *CacheDir
	gc       *GarbageCollector
	lock     sync.Locker
	queue    *qlose.Qlose
	inQueue  map[string]*MaybeError
}

func NewThumbnailQueue(cdir *CacheDir, gc *GarbageCollector) *ThumbnailQueue {
	tq := &ThumbnailQueue{
		cacheDir: cdir,
		gc:       gc,
		lock:     new(sync.Mutex),
		queue:    qlose.New(1, 128),
		inQueue:  make(map[string]*MaybeError),
	}
	runtime.SetFinalizer(tq, finalizer)
	return tq
}

func finalizer(tq *ThumbnailQueue) {
	tq.queue.Stop()
}

func (tq *ThumbnailQueue) IsInQueue(ti *ThumbInfo) bool {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	_, found := tq.inQueue[ti.String()]
	return found
}

func (tq *ThumbnailQueue) Enqueue(prio uint, ti *ThumbInfo) *MaybeError {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	if me, found := tq.inQueue[ti.String()]; found {
		return me
	} else {
		me = newMaybeError()
		if !ti.CanMake() {
			me.signal(NewNoSourceError(ti))
		} else {
			tq.queue.Enqueue(prio, func() interface{} {
				var err error
				if tq.cacheDir.Locate(ti) != "" {
					err = nil
					// Lazy Touch logic could go here:
					// existingPath := tq.cacheDir.Locate(ti)
					// updateAtime(existingPath)
				} else {
					out, err := tq.cacheDir.CreateFile(ti)
					if err == nil {
						err = ti.Make(out)
						if err == nil {
							// Successfully created. Track size.
							// out is closed by Make.
							path := tq.cacheDir.Locate(ti)
							fi, statErr := os.Stat(path)
							if statErr == nil {
								tq.gc.Track(fi.Size())
								// Update mtime/atime to now (Lazy Touch equivalent for new files)
								now := time.Now()
								os.Chtimes(path, now, now)
							}
						} else {
							// Failed to make. Cleanup partial file.
							path := tq.cacheDir.Locate(ti)
							os.Remove(path)
						}
					}
				}
				tq.lock.Lock()
				defer tq.lock.Unlock()
				delete(tq.inQueue, ti.String())
				me.signal(err)
				return err
			})
		}
		return me
	}
}
