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
	if cache.dir != cache_dir {
		t.Errorf("extected=%q, but actual=%q", cache.dir, cache_dir)
	}
}

func TestCacheDirSaveAs(t *testing.T) {
	const fileName = "tests/IMG_20141207_201549.jpg"
	const correctHash = "2b4656041c1922391a04c3e08e6ed362ebf902ca"
	cacheDir, err := NewCacheDir("tests/cache")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("tests/cache")
	thumbInfo, err := NewThumbInfoFromFile(fileName, SizeFromUint(480, 480), ModeFill, FMT_JPG)
	if err != nil {
		t.Error(err)
	}

	if cacheDir.Exists(thumbInfo) {
		t.Errorf("thumbnail should not exist yet")
	}

	path := cacheDir.Locate(thumbInfo)
	expected, _ := filepath.Abs("./tests/cache/2b/4656041c1922391a04c3e08e6ed362ebf902ca-480x480-fill.jpg")
	if expected != path {
		t.Errorf(`unnexpected result: %s`, path)
	}

	writer, err := cacheDir.CreateFile(thumbInfo)
	if err != nil {
		t.Error(err)
	}
	err = thumbInfo.Make(writer)
	if err != nil {
		t.Error(err)
	}

	if !cacheDir.Exists(thumbInfo) {
		t.Errorf("thumbnail should exist now")
	}
}
