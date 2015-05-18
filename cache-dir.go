package estelle

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type CacheDir struct {
	dir string
}

func NewCacheDir(path string) (*CacheDir, error) {
	abs_path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	abs_path = filepath.Clean(abs_path)
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
	return &CacheDir{dir: abs_path}, nil
}

func (cdir *CacheDir) CreateFile(ti *ThumbInfo) (io.WriteCloser, error) {
	path := cdir.Locate(ti)
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

func (cdir *CacheDir) Locate(ti *ThumbInfo) string {
	return fmt.Sprintf("%s/%s/%s", cdir.dir, ti.Id[:2], ti.Id[2:])
}

func (cdir *CacheDir) Exists(ti *ThumbInfo) bool {
	_, err := os.Stat(cdir.Locate(ti))
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}
