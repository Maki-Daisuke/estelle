package estelle

import (
	"context"
	"path/filepath"
)

type Estelle struct {
	cacheDir *CacheDir
	queue    *ThumbnailQueue
	gc       *GarbageCollector
}

func New(ctx context.Context, path string, cacheLimit int64, gcHighRatio, gcLowRatio float64) (*Estelle, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	cdir, err := NewCacheDir(filepath.Clean(path + "/thumbs"))
	if err != nil {
		return nil, err
	}
	gc := NewGarbageCollector(ctx, cdir.path, cacheLimit, gcHighRatio, gcLowRatio)
	queue := NewThumbnailQueue(cdir, gc)
	return &Estelle{cacheDir: cdir, queue: queue, gc: gc}, nil
}

func (estl *Estelle) Exists(ti *ThumbInfo) bool {
	return estl.cacheDir.Locate(ti) != ""
}

func (estl *Estelle) Enqueue(priority uint, ti *ThumbInfo) *MaybeError {
	return estl.queue.Enqueue(priority, ti)
}

func (estl *Estelle) Get(priority uint, ti *ThumbInfo) (string, error) {
	if !estl.Exists(ti) {
		err := estl.queue.Enqueue(priority, ti).Wait()
		if err != nil {
			return "", err
		}
	}
	return estl.cacheDir.Locate(ti), nil
}

func (estl *Estelle) IsInQueue(ti *ThumbInfo) bool {
	return estl.queue.IsInQueue(ti)
}
