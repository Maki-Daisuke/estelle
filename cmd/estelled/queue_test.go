package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	estellev2 "github.com/Maki-Daisuke/estelle/v2"
)

func TestQueueOverflow(t *testing.T) {
	// Setup temporary cache directory
	tempCache, err := os.MkdirTemp("", "estelle-test-queue")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempCache)

	// Setup large source image to ensure processing takes some time
	tempSourceFile := filepath.Join(tempCache, "large_source.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 2000, 2000)) // 2000x2000 should take > 10ms
	// Fill with some pattern to avoid compression being too fast?
	// Just noise is fine.
	for x := 0; x < 2000; x += 100 {
		for y := 0; y < 2000; y += 100 {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 0, 255})
		}
	}

	f, err := os.Create(tempSourceFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Initialize Estelle with 1 worker and 1 buffer (using 1 to be safe against 0=infinite)
	var errInit error
	estelle, errInit = estellev2.New(tempCache,
		estellev2.WithWorkers(1),
		estellev2.WithBufferSize(1),
	)
	if errInit != nil {
		t.Fatal(errInit)
	}
	defer estelle.Shutdown(context.Background())

	// Set allowedDirs global for testing
	allowedDirs = []string{tempCache}

	// Setup Router
	router := http.NewServeMux()
	router.HandleFunc("GET /queue", handleQueue)
	router.HandleFunc("POST /queue", handleQueue)

	ts := httptest.NewServer(router)
	defer ts.Close()

	// 1. Send 10 requests concurrently.
	// 1 worker + 1 buffer = 2 tasks max.
	// We send 10. Should see some 503s.

	var wg sync.WaitGroup
	results := make(chan int, 20)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		
		// Create a unique dummy file for each request to bypass deduplication
		uniqueTemp := filepath.Join(tempCache, fmt.Sprintf("dummy_%d.jpg", i))
		os.WriteFile(uniqueTemp, []byte(fmt.Sprintf("dummy %d", i)), 0644)
		
		go func(i int, sourceFile string) {
			defer wg.Done()
			reqURL := fmt.Sprintf("%s/queue?source=%s&size=100x100", ts.URL, sourceFile)

			resp, err := http.Get(reqURL)
			if err != nil {
				t.Logf("Req %d error: %v", i, err)
				results <- 0
				return
			}
			resp.Body.Close()
			results <- resp.StatusCode
		}(i, uniqueTemp)
	}

	wg.Wait()
	close(results)

	successCount := 0
	overflowCount := 0

	for code := range results {
		if code == 200 || code == 202 {
			successCount++
		} else if code == 503 {
			overflowCount++
		} else {
			t.Errorf("Unexpected status code: %d", code)
		}
	}

	t.Logf("Success: %d, Overflow: %d", successCount, overflowCount)

	if overflowCount == 0 {
		t.Error("Expected at least some 503s due to queue overflow, got none")
	}
	if successCount == 0 {
		t.Error("Expected at least one success, got none")
	}
}
