package estelle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestThumbInfoFactory(t *testing.T) {
	baseDir := "tests/cache"
	factory, err := NewThumbInfoFactory(baseDir)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(baseDir)

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	expectedDir := filepath.Clean(filepath.Join(wd, baseDir))
	if factory.BaseDir() != expectedDir {
		t.Errorf("expected=%q, but actual=%q", expectedDir, factory.BaseDir())
	}
}

func TestThumbInfoMake(t *testing.T) {
	const fileName = "tests/IMG_20141207_201549.jpg"
	baseDir := "tests/cache"
	factory, err := NewThumbInfoFactory(baseDir)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(baseDir)

	thumbInfo, err := factory.FromFile(fileName, SizeFromUint(480, 480), ModeCrop, FMT_JPG)
	if err != nil {
		t.Error(err)
	}

	if thumbInfo.Exists() {
		t.Errorf("thumbnail should not exist yet")
	}

	path := thumbInfo.Path()

	// Expected format: .../cache/xx/yy/full_hash-...
	id := thumbInfo.String()
	expectedRel := filepath.Join(baseDir, id[:2], id[2:4], id)
	expected, _ := filepath.Abs(expectedRel)

	if expected != path {
		t.Errorf("Unexpected path.\nExpected: %s\nActual: %s", expected, path)
	}

	err = thumbInfo.Make()
	if err != nil {
		t.Fatalf("Failed to make thumbnail: %v", err)
	}

	if !thumbInfo.Exists() {
		t.Errorf("thumbnail should exist now")
	}

	// Double check file existence
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file not found at %s", path)
	}
}
