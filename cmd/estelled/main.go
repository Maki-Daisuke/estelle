package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	. "github.com/Maki-Daisuke/estelle"

	"github.com/caarlos0/env/v11"
)

var config struct {
	Addr           string  `env:"ESTELLE_ADDR" envDefault:":1186"`
	AllowedDirs    string  `env:"ESTELLE_ALLOWED_DIRS"`
	CacheDir       string  `env:"ESTELLE_CACHE_DIR"`
	Limit          string  `env:"ESTELLE_CACHE_LIMIT" envDefault:"1GB"`
	GCHighRatio    float64 `env:"ESTELLE_GC_HIGH_RATIO" envDefault:"0.90"`
	GCLowRatio     float64 `env:"ESTELLE_GC_LOW_RATIO" envDefault:"0.75"`
	WorkerPoolSize int     `env:"ESTELLE_WORKERS"`
	TaskBufferSize int     `env:"ESTELLE_QUEUE_SIZE" envDefault:"1024"`
}

var estelle *Estelle
var allowedDirs []string

type ForbiddenError struct {
	msg string
}

func (e ForbiddenError) Error() string { return e.msg }

func main() {
	if err := env.Parse(&config); err != nil {
		slog.Error("Failed to parse env", "error", err)
		os.Exit(1)
	}

	if config.CacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("Failed to get user home dir", "error", err)
			os.Exit(1)
		}
		config.CacheDir = filepath.Join(home, ".cache", "estelled")
	}

	if config.AllowedDirs == "" {
		slog.Error("ESTELLE_ALLOWED_DIRS is required")
		os.Exit(1)
	}
	allowedDirs = filepath.SplitList(config.AllowedDirs)
	for i, dir := range allowedDirs {
		abs, err := filepath.Abs(dir)
		if err != nil {
			slog.Error("Failed to get absolute path", "dir", dir, "error", err)
			os.Exit(1)
		}
		allowedDirs[i] = abs + "/"
	}

	limitBytes, err := parseBytes(config.Limit)
	if err != nil {
		slog.Error("Invalid limit format", "ESTELLE_CACHE_LIMIT", config.Limit, "error", err)
		os.Exit(1)
	}

	if config.WorkerPoolSize == 0 {
		config.WorkerPoolSize = runtime.NumCPU() / 2
		if config.WorkerPoolSize < 1 {
			config.WorkerPoolSize = 1
		}
	}

	// Setup signal handler to properly shutdown the goroutine behind Estelle
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	estelle, err = New(config.CacheDir,
		WithCacheLimit(limitBytes),
		WithGCRatio(config.GCHighRatio, config.GCLowRatio),
		WithWorkers(config.WorkerPoolSize),
		WithBufferSize(config.TaskBufferSize),
		WithPanicHandler(func(v interface{}) {
			slog.Error("Worker Panic", "panic", v, "stack", string(debug.Stack()))
		}),
	)
	if err != nil {
		slog.Error("Failed to initialize estelle", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /get", handleGet)
	mux.HandleFunc("POST /get", handleGet)
	mux.HandleFunc("GET /queue", handleQueue)
	mux.HandleFunc("POST /queue", handleQueue)

	handler := withRecovery(withLogger(mux))

	network := "tcp"
	addr := config.Addr
	if strings.HasPrefix(addr, "unix://") {
		network = "unix"
		addr = strings.TrimPrefix(addr, "unix://")
	}

	l, err := net.Listen(network, addr)
	if err != nil {
		slog.Error("Failed to listen", "addr", config.Addr, "error", err)
		os.Exit(1)
	}
	defer l.Close()

	server := &http.Server{
		Handler: handler,
	}

	go func() {
		slog.Info("listening", "network", network, "addr", addr)
		if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	if err := estelle.Shutdown(shutdownCtx); err != nil {
		slog.Error("Estelle shutdown failed", "error", err)
	}

	if network == "unix" {
		if err := os.Remove(addr); err != nil {
			slog.Error("Failed to remove socket file", "error", err)
		} else {
			slog.Info("Socket file removed")
		}
	}
}

func parseBytes(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" || s == "0" {
		return 0, nil
	}
	var unit int64 = 1
	if strings.HasSuffix(s, "KB") || strings.HasSuffix(s, "K") {
		unit = 1024
		s = strings.TrimRight(s, "KB")
		s = strings.TrimRight(s, "K")
	} else if strings.HasSuffix(s, "MB") || strings.HasSuffix(s, "M") {
		unit = 1024 * 1024
		s = strings.TrimRight(s, "MB")
		s = strings.TrimRight(s, "M")
	} else if strings.HasSuffix(s, "GB") || strings.HasSuffix(s, "G") {
		unit = 1024 * 1024 * 1024
		s = strings.TrimRight(s, "GB")
		s = strings.TrimRight(s, "G")
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return val * unit, nil
}

func handleGet(res http.ResponseWriter, req *http.Request) {
	ti, err := thumbInfoFromReq(req)
	if err != nil {
		if os.IsNotExist(err) {
			res.WriteHeader(404)
			res.Write([]byte("Not found"))
			return
		}
		if errors.As(err, &ForbiddenError{}) {
			res.WriteHeader(http.StatusForbidden)
			res.Write([]byte("Access denied"))
			return
		}
		panic(err)
	}
	c := estelle.Enqueue(ti)
	err = <-c
	if err != nil {
		panic(err)
	}
	res.WriteHeader(200)
	res.Write([]byte(ti.Path()))
}

func handleQueue(res http.ResponseWriter, req *http.Request) {
	ti, err := thumbInfoFromReq(req)
	if err != nil {
		if os.IsNotExist(err) {
			res.WriteHeader(404)
			res.Write([]byte("Not found"))
			return
		}
		if errors.As(err, &ForbiddenError{}) {
			res.WriteHeader(http.StatusForbidden)
			res.Write([]byte("Access denied"))
			return
		}
		panic(err)
	}
	c := estelle.Enqueue(ti)
	// Does not wait for the thumbnail to be created.
	select {
	case err, ok := <-c:
		if !ok && err != nil {
			panic(err)
		}
		res.WriteHeader(200)
		res.Write([]byte(ti.Path()))
	default:
		res.WriteHeader(202) // Accepted
	}
}

func thumbInfoFromReq(req *http.Request) (ThumbInfo, error) {
	if !(len(req.URL.Query()["source"]) > 0) {
		return ThumbInfo{}, fmt.Errorf(`"source" is required`)
	}
	source := req.URL.Query()["source"][0]
	if source == "" || source[0] != '/' {
		source = "/" + source
	}
	source = filepath.Clean(source)

	allowed := false
	for _, dir := range allowedDirs {
		if strings.HasPrefix(source, dir) {
			allowed = true
			break
		}
	}
	if !allowed {
		return ThumbInfo{}, ForbiddenError{msg: fmt.Sprintf("Access denied: %s is not in allowed directories", source)}
	}

	size := parseQuerySize(req.URL.Query()["size"])
	mode := parseQueryMode(req.URL.Query()["mode"])
	format := parseQueryFormat(req.URL.Query()["format"])
	return estelle.NewThumbInfo(source, size, mode, format)
}

func parseQuerySize(query []string) Size {
	if len(query) > 0 {
		size, err := SizeFromString(query[0])
		if err == nil {
			return size
		}
	}
	return SizeFromUint(85, 85)
}

func parseQueryMode(query []string) Mode {
	if len(query) > 0 {
		m := ModeFromString(query[0])
		if m != ModeUnknown {
			return m
		}
	}
	return ModeCrop
}

func parseQueryFormat(query []string) Format {
	if len(query) > 0 {
		f := FormatFromString(query[0])
		if f != FMT_UNKNOWN {
			return f
		}
	}
	return FMT_JPG
}
