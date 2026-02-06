package estelle

import (
	"context"
	"os"
	"sync"

	"github.com/Maki-Daisuke/go-filiq"
	"golang.org/x/sync/singleflight"
)

type Estelle struct {
	dir    ThumbInfoFactory
	runner *filiq.Runner
	sf     *singleflight.Group
	gc     *GarbageCollector
}

func New(path string, cacheLimit int64, gcHighRatio, gcLowRatio float64) (*Estelle, error) {
	dir, err := NewThumbInfoFactory(path)
	if err != nil {
		return nil, err
	}
	return &Estelle{
		dir:    dir,
		runner: filiq.New(filiq.WithLIFO(), filiq.WithWorkers(2), filiq.WithBufferSize(1024)),
		sf:     new(singleflight.Group),
		gc:     NewGarbageCollector(dir.BaseDir(), cacheLimit, gcHighRatio, gcLowRatio),
	}, nil
}

func (estl *Estelle) Shutdown(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		estl.runner.Shutdown(ctx)
	}()
	go func() {
		defer wg.Done()
		estl.gc.Shutdown(ctx)
	}()
	wg.Wait()
	return ctx.Err()
}

func (estl *Estelle) NewThumbInfo(path string, size Size, mode Mode, format Format) (ThumbInfo, error) {
	return estl.dir.FromFile(path, size, mode, format)
}

func (estl *Estelle) Enqueue(ti ThumbInfo) <-chan error {
	c := make(chan error)
	if ti.Exists() {
		close(c)
		return c
	}
	estl.runner.Put(func() {
		defer close(c)
		_, err, _ := estl.sf.Do(ti.String(), func() (any, error) {
			if ti.Exists() {
				return nil, nil
			}
			if err := ti.Make(); err != nil {
				return nil, err
			}
			st, err := os.Stat(ti.Path())
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
