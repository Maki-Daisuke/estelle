package estelle

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Maki-Daisuke/go-filiq"
	"golang.org/x/sync/singleflight"
)

type Estelle struct {
	cacheDir *CacheDir
	runner   *filiq.Runner
	sf       *singleflight.Group
	gc       *GarbageCollector
}

func New(ctx context.Context, path string, cacheLimit int64, gcHighRatio, gcLowRatio float64) (*Estelle, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	cdir, err := NewCacheDir(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	return &Estelle{
		cacheDir: cdir,
		runner:   filiq.New(filiq.WithLIFO(), filiq.WithWorkers(2), filiq.WithBufferSize(1024)),
		sf:       new(singleflight.Group),
		gc:       NewGarbageCollector(ctx, cdir.path, cacheLimit, gcHighRatio, gcLowRatio),
	}, nil
}

func (estl *Estelle) ThumbPath(ti ThumbInfo) string {
	return estl.cacheDir.ThumbPath(ti)
}

func (estl *Estelle) Exists(ti ThumbInfo) bool {
	return estl.cacheDir.Locate(ti) != ""
}

func (estl *Estelle) Enqueue(ti ThumbInfo) <-chan error {
	c := make(chan error)
	if estl.Exists(ti) {
		close(c)
		return c
	}
	estl.runner.Put(func() {
		defer close(c)
		_, err, _ := estl.sf.Do(ti.String(), func() (any, error) {
			if estl.Exists(ti) {
				return nil, nil
			}
			w, err := estl.cacheDir.CreateFile(ti)
			if err != nil {
				return nil, err
			}
			defer w.Close()
			if err := ti.Make(w); err != nil {
				return nil, err
			}
			st, err := os.Stat(estl.cacheDir.ThumbPath(ti))
			if err != nil {
				return nil, err
			}
			estl.gc.Track(st.Size())
			return nil, nil
		})
		if err != nil {
			c <- err
		}
	})
	return c
}
