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

type config struct {
	cacheLimit   int64
	gcHighRatio  float64
	gcLowRatio   float64
	workerNum    int
	bufferSize   int
	panicHandler func(interface{})
}

type Option func(*config)

func WithCacheLimit(limit int64) Option {
	return func(c *config) {
		c.cacheLimit = limit
	}
}

func WithGCRatio(high, low float64) Option {
	return func(c *config) {
		c.gcHighRatio = high
		c.gcLowRatio = low
	}
}

func WithWorkers(n int) Option {
	return func(c *config) {
		c.workerNum = n
	}
}

func WithBufferSize(size int) Option {
	return func(c *config) {
		c.bufferSize = size
	}
}

func WithPanicHandler(h func(interface{})) Option {
	return func(c *config) {
		c.panicHandler = h
	}
}

func New(path string, opts ...Option) (*Estelle, error) {
	dir, err := NewThumbInfoFactory(path)
	if err != nil {
		return nil, err
	}

	cfg := config{
		// Default values
		cacheLimit:  1024 * 1024 * 1024, // 1GB default
		gcHighRatio: 0.90,
		gcLowRatio:  0.75,
		workerNum:   1, // Safe default
		bufferSize:  1024,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	filiqOpts := []filiq.Option{
		filiq.WithLIFO(),
		filiq.WithWorkers(cfg.workerNum),
		filiq.WithBufferSize(cfg.bufferSize),
	}
	if cfg.panicHandler != nil {
		filiqOpts = append(filiqOpts, filiq.WithPanicHandler(cfg.panicHandler))
	}

	return &Estelle{
		dir:    dir,
		runner: filiq.New(filiqOpts...),
		sf:     new(singleflight.Group),
		gc:     NewGarbageCollector(dir.BaseDir(), cfg.cacheLimit, cfg.gcHighRatio, cfg.gcLowRatio),
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
