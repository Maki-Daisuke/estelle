package estelle

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestEnqueueAfterShutdown(t *testing.T) {
	// Create a temporary directory for cache
	cacheDir := t.TempDir()

	// Initialize Estelle
	estl, err := New(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create Estelle: %v", err)
	}

	// Create a dummy image file
	dummyImage := "test.jpg"
	if err := os.WriteFile(dummyImage, []byte("dummy image content"), 0644); err != nil {
		t.Fatalf("Failed to create dummy image: %v", err)
	}
	defer os.Remove(dummyImage)

	// Create a dummy ThumbInfo
	ti, err := estl.NewThumbInfo(dummyImage, Size{Width: 100, Height: 100}, ModeCrop, FMT_JPG)
	if err != nil {
		t.Fatalf("Failed to create ThumbInfo: %v", err)
	}

	// Shutdown Estelle
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := estl.Shutdown(ctx); err != nil {
		t.Fatalf("Failed to shutdown Estelle: %v", err)
	}

	// Try to Enqueue after Shutdown
	// This should return an error channel with ErrEstelleClosed
	errChan := estl.Enqueue(ti)
	select {
	case err := <-errChan:
		if err != ErrEstelleClosed {
			t.Errorf("Expected ErrEstelleClosed, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout waiting for error")
	}
}
