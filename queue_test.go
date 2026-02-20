package estelle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestQueueDeduplication(t *testing.T) {
	// Setup temporary directory structure
	tmpDir, err := os.MkdirTemp("", "estelle-test-queue")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	cacheDir := filepath.Join(tmpDir, "cache")
	if err := os.Mkdir(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a dummy source file
	srcFile := filepath.Join(srcDir, "test.jpg")
	if err := os.WriteFile(srcFile, []byte("dummy image content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize Estelle
	// Use 1 worker to make race conditions easier to trigger if we want, or just standard.
	// We'll use a larger buffer to avoid blocking on simple tests.
	estl, err := New(cacheDir, WithWorkers(1), WithBufferSize(10))
	if err != nil {
		t.Fatal(err)
	}
	defer estl.Shutdown(context.Background())

	t.Run("Broadcast Result", func(t *testing.T) {
		// Re-create src file to ensure different hash if needed? 
		// Actually ti includes hash.
		// Let's make a new dummy file to get a fresh task.
		srcFile2 := filepath.Join(srcDir, "test2.jpg")
		if err := os.WriteFile(srcFile2, []byte("dummy image content 2"), 0644); err != nil {
			t.Fatal(err)
		}
		ti2, err := estl.dir.FromFile(srcFile2, Size{Width: 100, Height: 100}, ModeCrop, FMT_JPG)
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		n := 5
		errors := make([]error, n)

		// Start multiple Enqueue
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				res, err := estl.Enqueue(ti2)
				if err != nil {
					errors[idx] = err
				} else if res != nil {
					<-res.Done()
					errors[idx] = res.Err()
				}
			}(i)
		}

		wg.Wait()

		// All errors should be the same (likely "vipsthumbnail failed" or similar, or nil if it works)
		firstErr := errors[0]
		for i := 1; i < n; i++ {
			if errors[i] != firstErr { // Error equality might be checking pointer for wrapped errors
				// Using string comparison for "same error message" at least
				if fmt.Sprintf("%v", errors[i]) != fmt.Sprintf("%v", firstErr) {
					t.Errorf("Error mismatch at %d: got %v, want %v", i, errors[i], firstErr)
				}
			}
		}
	})
}

// TestStress creates high contention to ensure no races/panics
func TestQueueStress(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "estelle-stress")
	defer os.RemoveAll(tmpDir)
	srcDir := filepath.Join(tmpDir, "src")
	os.Mkdir(srcDir, 0755)
	cacheDir := filepath.Join(tmpDir, "cache")
	os.Mkdir(cacheDir, 0755)
	
	estl, _ := New(cacheDir, WithWorkers(4), WithBufferSize(100))
	defer estl.Shutdown(context.Background())

	// Single target file
	srcFile := filepath.Join(srcDir, "stress.jpg")
	os.WriteFile(srcFile, []byte("stress"), 0644)
	ti, _ := estl.dir.FromFile(srcFile, Size{Width: 100, Height: 100}, ModeCrop, FMT_JPG)

	var wg sync.WaitGroup
	n := 100
	
	start := make(chan struct{})
	
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // synchronize start
			res, _ := estl.Enqueue(ti)
			if res != nil {
				<-res.Done()
			}
		}()
	}
	
	close(start) // GO!
	wg.Wait()
}
