package estelle

import (
	"os"
	"sync"

	"github.com/Maki-Daisuke/qlose"
)

type promise struct {
	e error
	c chan struct{}
}

func newPromise() *promise {
	return &promise{
		c: make(chan struct{}),
	}
}

func (p *promise) Signal(e error) {
	p.e = e
	close(p.c)
}

func (p *promise) Wait() error {
	<-p.c
	return p.e
}

type ThumbnailQueue struct {
	locator func(*ThumbInfo) string
	lock    sync.Locker
	queue   *qlose.Qlose
	inQueue map[string]*promise
}

func NewThumbnailQueue(locator func(*ThumbInfo) string) *ThumbnailQueue {
	return &ThumbnailQueue{
		locator: locator,
		lock:    new(sync.Mutex),
		queue:   qlose.New(1, 128),
		inQueue: make(map[string]*promise),
	}
}

func (tq *ThumbnailQueue) IsInQueue(ti *ThumbInfo) bool {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	_, found := tq.inQueue[ti.Id]
	return found
}

func (tq *ThumbnailQueue) Enqueue(prio uint, ti *ThumbInfo) *promise {
	tq.lock.Lock()
	defer tq.lock.Unlock()
	if p, found := tq.inQueue[ti.Id]; found {
		return p
	} else {
		p := newPromise()
		tq.queue.Enqueue(prio, func() interface{} {
			out := tq.locator(ti)
			var err error
			if _, err := os.Stat(out); err == nil { // file does not exist
				err = ti.SaveAs(out)
			} else {
				err = nil
			}
			tq.lock.Lock()
			defer tq.lock.Unlock()
			delete(tq.inQueue, ti.Id)
			p.Signal(err)
			return err
		})
		return p
	}
}
