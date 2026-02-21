package estelle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Maki-Daisuke/go-filiq"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// ErrEstelleClosed is returned when the Estelle instance is already shut down.
var ErrEstelleClosed = fmt.Errorf("estelle is already shutdown")

// ErrEstelleQueueFull is returned when the internal task queue is full.
var ErrEstelleQueueFull = fmt.Errorf("estelle queue is full")

// Result represents the status of an enqueued task.
// It follows a context-like pattern allowing select-based waiting.
type Result struct {
	done chan struct{} // Closed when the task finishes
	err  error         // The resulting error, valid only after done is closed
}

// Done returns a channel that's closed when the task completes.
// This allows the Result to be used in select statements.
func (r *Result) Done() <-chan struct{} {
	return r.done
}

// Err returns the error resulting from the task.
// It is only safe to call after the Done channel is closed.
func (r *Result) Err() error {
	return r.err
}

// Estelle is the main thumbnail generation engine that manages the queue, worker pool, and garbage collection.
type Estelle struct {
	dir          ThumbInfoFactory
	runner       *filiq.Runner
	gc           *garbageCollector
	pendingTasks atomic.Pointer[cmap.ConcurrentMap[string, *Result]]
}

type config struct {
	cacheLimit   int64
	gcHighRatio  float64
	gcLowRatio   float64
	workerNum    int
	bufferSize   int
	panicHandler func(interface{})
}

// Option defines a functional option for configuring an Estelle instance.
type Option func(*config)

// WithCacheLimit sets the maximum cache size in bytes for the garbage collector.
func WithCacheLimit(limit int64) Option {
	return func(c *config) {
		c.cacheLimit = limit
	}
}

// WithGCRatio sets the high and low watermarks (ratios between 0.0 and 1.0) for cache garbage collection.
func WithGCRatio(high, low float64) Option {
	return func(c *config) {
		c.gcHighRatio = high
		c.gcLowRatio = low
	}
}

// WithWorkers sets the number of concurrent worker routines for thumbnail generation.
func WithWorkers(n int) Option {
	return func(c *config) {
		c.workerNum = n
	}
}

// WithBufferSize sets the size of the internal task queue. 0 means unbounded.
func WithBufferSize(size int) Option {
	return func(c *config) {
		c.bufferSize = size
	}
}

// WithPanicHandler sets a custom panic handler for the thumbnail generation workers.
func WithPanicHandler(h func(interface{})) Option {
	return func(c *config) {
		c.panicHandler = h
	}
}

// New creates a new Estelle instance.
// It initializes the underlying directory structure, worker pool, and garbage collection.
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

	estl := &Estelle{
		dir:    dir,
		runner: filiq.New(filiqOpts...),
		gc:     newGarbageCollector(dir.BaseDir(), cfg.cacheLimit, cfg.gcHighRatio, cfg.gcLowRatio),
	}
	cm := cmap.New[*Result]()
	estl.pendingTasks.Store(&cm)
	return estl, nil
}

// Shutdown gracefully stops the thumbnail generation workers and the garbage collector.
// It closes the channels of all pending tasks to unblock any waiting clients.
func (estl *Estelle) Shutdown(ctx context.Context) error {
	pending := estl.pendingTasks.Swap(nil)
	if pending == nil {
		return ErrEstelleClosed
	}
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
	// Close channels of all pending tasks so that any goroutine waiting on them will unblock.
	for _, k := range pending.Keys() {
		r, ok := pending.Pop(k)
		if !ok {
			continue
		}
		tryClose(r.done)
	}
	return ctx.Err()
}

// NewThumbInfo creates a ThumbInfo for a given source path, size, mode, and format.
func (estl *Estelle) NewThumbInfo(path string, size Size, mode Mode, format Format) (ThumbInfo, error) {
	return estl.dir.FromFile(path, size, mode, format)
}

// closedResult is a Result that is already closed.
// We reuse this for optimization for the case where the thumbnail already exists.
var closedResult = &Result{done: func() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}()}

// Enqueue submits a thumbnail generation task to the queue.
// It returns a newly created `*Result` object and a possible immediate error.
// If the thumbnail already exists, it returns `(res, nil)`.
// If the task is queued (or already pending), it returns `(res, nil)`.
// If the queue is full or already closed, it returns `(nil, error)`.
func (estl *Estelle) Enqueue(ti ThumbInfo) (*Result, error) {
	if ti.Exists() {
		return closedResult, nil
	}

	res := &Result{done: make(chan struct{})}
	key := ti.String()

	pending := estl.pendingTasks.Load()
	if pending == nil {
		return nil, ErrEstelleClosed
	}

	// Try to set new task as pending
	if pending.SetIfAbsent(key, res) {
		// We successfully registered a new task.
		task := estl.makeTask(ti, res)
		if err := estl.runner.Submit(task); err != nil {
			// Runner closed or Queue full
			pending.Remove(key) // cleanup
			if errors.Is(err, filiq.ErrQueueClosed) {
				res.err = ErrEstelleClosed
			} else if errors.Is(err, filiq.ErrQueueFull) {
				res.err = ErrEstelleQueueFull
			} else {
				res.err = err
			}
			tryClose(res.done) // Unblock any listeners (just in case)
			// The channel may be closed already by Shutdown.
			return nil, res.err
		}
		// Return the waitable Result struct
		return res, nil
	}

	// Task is already pending or running. Return the existing Result.
	if actual, ok := pending.Get(key); ok {
		return actual, nil
	}

	// Race condition edge case:
	// SetIfAbsent returned false (key existed), but Get returned false (key removed).
	// This means the task JUST finished. Let's retry Enqueue (will hit Exists() fast path).
	return estl.Enqueue(ti)
}

// makeTask creates a thunk that executes the thumbnail generation.
func (estl *Estelle) makeTask(ti ThumbInfo, res *Result) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				res.err = err.(error)
			}
			close(res.done)
			pending := estl.pendingTasks.Load()
			if pending != nil {
				pending.Remove(ti.String())
			}
		}()

		if ti.Exists() {
			return
		}
		if err := ti.make(); err != nil {
			res.err = err
		}
		st, err := os.Stat(ti.Path())
		if err != nil {
			res.err = err
		}
		estl.gc.Track(st.Size())
	}
}

func tryClose(ch chan struct{}) {
	defer func() {
		recover()
	}()
	close(ch) // calling close on a closed channel panics
}
