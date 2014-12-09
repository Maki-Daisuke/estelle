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
	defer func() {
		os.RemoveAll("tests/cache")
	}()
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	cache_dir := filepath.Clean(wd + "/tests/cache")
	if cache.dir != cache_dir {
		t.Errorf("extected=%q, but actual=%q", cache.dir, cache_dir)
	}
	thumb_dir := filepath.Clean(cache_dir + "/thumbs")
	if cache.thumbDir != thumb_dir {
		t.Errorf("extected=%q, but actual=%q", cache.thumbDir, thumb_dir)
	}
}

func TestCacheDirSaveAs(t *testing.T) {
	const fileName = "tests/IMG_20141207_201549.jpg"
	const correctHash = "2b4656041c1922391a04c3e08e6ed362ebf902ca"
	cacheDir, err := NewCacheDir("tests/cache")
	if err != nil {
		t.Error(err)
	}
	defer func() { os.RemoveAll("tests/cache") }()
	thumbInfo, err := NewThumbInfoFromFile(fileName, 480, 480, ModeFill, "jpg")
	if err != nil {
		t.Error(err)
	}

	if cacheDir.Exists(thumbInfo) {
		t.Errorf("thumbnail should not exist yet")
	}

	path := cacheDir.Locate(thumbInfo)
	expected, _ := filepath.Abs("./tests/cache/thumbs/2b/4656041c1922391a04c3e08e6ed362ebf902ca-480x480-fill.jpg")
	if expected != path {
		t.Errorf(`unnexpected result: %s`, path)
	}

	err = thumbInfo.SaveAs(path)
	if err != nil {
		t.Error(err)
	}

	if !cacheDir.Exists(thumbInfo) {
		t.Errorf("thumbnail should exist now")
	}
}
