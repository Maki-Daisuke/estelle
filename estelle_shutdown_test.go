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

	res, err := estl.Enqueue(ti)
	if err == nil && res != nil {
		<-res.Done()
		err = res.Err()
	}
	
	if err != ErrEstelleClosed {
		t.Errorf("Expected ErrEstelleClosed, got %v", err)
	}
}
