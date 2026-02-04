package estelle

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type CacheDir struct {
	path string
}

func NewCacheDir(path string) (*CacheDir, error) {
	abs_path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if stat, err := os.Stat(abs_path); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		err = os.MkdirAll(abs_path, 0755)
		if err != nil {
			return nil, err
		}
	} else if !stat.IsDir() {
		return nil, fmt.Errorf(`"%s" exists, but it is not a dirctory`, abs_path)
	}
	temp, err := os.CreateTemp(abs_path, "estelle-test-*")
	if err != nil {
		return nil, fmt.Errorf("cache directory (%s) is not writable: %s", abs_path, err)
	}
	temp.Close()
	os.Remove(temp.Name())
	return &CacheDir{path: abs_path}, nil
}

func (cdir *CacheDir) CreateFile(ti ThumbInfo) (io.WriteCloser, error) {
	path := cdir.ThumbPath(ti)
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (cdir *CacheDir) ThumbPath(ti ThumbInfo) string {
	id := ti.String()
	return filepath.Join(cdir.path, id[:2], id[2:4], id)
}

// Locate returns the absolute path to the cache file for the given ThumbInfo.
// It returns an empty string if the cache file does not exist.
func (cdir *CacheDir) Locate(ti ThumbInfo) string {
	path := cdir.ThumbPath(ti)
	_, err := os.Stat(path)
	if err != nil {
		// Returns empty string if cache does not exist, but logs other errors.
		if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("Error checking cache existence: %v", err)
		}
		return ""
	}
	return path
}
