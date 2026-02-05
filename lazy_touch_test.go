package estelle

import (
	"os"
	"testing"
	"time"
)

func TestLazyTouch(t *testing.T) {
	// 1. Prepare
	const fileName = "tests/IMG_20141207_201549.jpg"
	baseDir := "tests/cache_lazy_touch"
	factory, err := NewThumbInfoFactory(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(baseDir)

	thumbInfo, err := factory.FromFile(fileName, SizeFromUint(100, 100), ModeCrop, FMT_JPG)
	if err != nil {
		t.Fatal(err)
	}

	// Create thumbnail
	err = thumbInfo.Make()
	if err != nil {
		t.Fatalf("Failed to make thumbnail: %v", err)
	}

	if !thumbInfo.Exists() {
		t.Fatal("Thumbnail should exist")
	}

	path := thumbInfo.Path()

	// wait 3 seconds
	time.Sleep(3 * time.Second)

	// 2. Set Atime/Mtime to 48 hours ago
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(path, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to chtimes: %v", err)
	}

	// Verify it is firmly in the past
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// Check if Atime is old enough (allowing some margin for filesystem resolution)
	if !GetAtime(fi).Before(oldTime.Add(1 * time.Hour)) {
		t.Fatalf("Failed to set old time. Current Atime: %v, Target: %v", GetAtime(fi), oldTime)
	}

	// 3. Call Exists() -> Should trigger Lazy Touch because file is old
	if !thumbInfo.Exists() {
		t.Fatal("Exists returned false")
	}

	// 4. Verify Atime is updated to Now
	fi, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	// It should be close to Now (e.g. within 5 seconds)
	if time.Since(GetAtime(fi)) > 5*time.Second {
		t.Errorf("Atime was not updated! Got: %v, Expected close to now. (Diff: %v)", GetAtime(fi), time.Since(GetAtime(fi)))
	}
}
