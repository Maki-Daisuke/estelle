package estelle

import "path/filepath"

type Estelle struct {
	cacheDir *CacheDir
	queue    *ThumbnailQueue
}

func New(path string) (*Estelle, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	cdir, err := NewCacheDir(filepath.Clean(path + "/thumbs"))
	if err != nil {
		return nil, err
	}
	queue := NewThumbnailQueue(cdir)
	return &Estelle{cacheDir: cdir, queue: queue}, nil
}

func (estl *Estelle) Exists(ti *ThumbInfo) bool {
	return estl.cacheDir.Exists(ti)
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
