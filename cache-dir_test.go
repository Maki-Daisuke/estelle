package estelle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheDir(t *testing.T) {
	cache, err := NewCacheDir("tests/cache")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("tests/cache")
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	cache_dir := filepath.Clean(wd + "/tests/cache")
	if cache.path != cache_dir {
		t.Errorf("extected=%q, but actual=%q", cache.path, cache_dir)
	}
}

func TestCacheDirSaveAs(t *testing.T) {
	const fileName = "tests/IMG_20141207_201549.jpg"
	cacheDir, err := NewCacheDir("tests/cache")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("tests/cache")
	thumbInfo, err := NewThumbInfoFromFile(fileName, SizeFromUint(480, 480), ModeFill, FMT_JPG)
	if err != nil {
		t.Error(err)
	}

	if cacheDir.Locate(thumbInfo) != "" {
		t.Errorf("thumbnail should not exist yet")
	}

	path := cacheDir.ThumbPath(thumbInfo)

	// Expected format: .../cache/xx/yy/full_hash-...
	id := thumbInfo.String()
	expectedRel := filepath.Join("tests", "cache", id[:2], id[2:4], thumbInfo.String())
	expected, _ := filepath.Abs(expectedRel)

	if expected != path {
		t.Errorf("Unexpected path.\nExpected: %s\nActual: %s", expected, path)
	}

	func() {
		writer, err := cacheDir.CreateFile(thumbInfo)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		defer writer.Close()
		err = thumbInfo.Make(writer)
		if err != nil {
			t.Fatalf("Failed to make thumbnail: %v", err)
		}
	}()

	if cacheDir.Locate(thumbInfo) == "" {
		t.Errorf("thumbnail should exist now")
	}
}
