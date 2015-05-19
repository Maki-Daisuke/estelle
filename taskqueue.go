package estelle

import (
	"runtime"
	"sync"

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
	lock     sync.Locker
	queue    *qlose.Qlose
	inQueue  map[string]*MaybeError
}

func NewThumbnailQueue(cdir *CacheDir) *ThumbnailQueue {
	tq := &ThumbnailQueue{
		cacheDir: cdir,
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
	_, found := tq.inQueue[ti.Id()]
	return found
}

func (tq *ThumbnailQueue) Enqueue(prio uint, ti *ThumbInfo) *MaybeError {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	if me, found := tq.inQueue[ti.Id()]; found {
		return me
	} else {
		me = newMaybeError()
		tq.queue.Enqueue(prio, func() interface{} {
			var err error
			if tq.cacheDir.Exists(ti) {
				err = nil
			} else {
				out, err := tq.cacheDir.CreateFile(ti)
				if err == nil {
					err = ti.Make(out)
				}
			}
			tq.lock.Lock()
			defer tq.lock.Unlock()
			delete(tq.inQueue, ti.Id())
			me.signal(err)
			return err
		})
		return me
	}
}
