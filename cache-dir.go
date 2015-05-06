package estelle

import (
	"fmt"
	"os"
	"path/filepath"
)

type CacheDir struct {
	dir      string
	thumbDir string
}

func NewCacheDir(path string) (*CacheDir, error) {
	abs_path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	cdir := &CacheDir{
		dir:      filepath.Clean(abs_path),
		thumbDir: filepath.Clean(abs_path + "/thumbs"),
	}
	stat, err := os.Stat(cdir.thumbDir)
	if err != nil {
		// Probably, it does not exist, then, try to mkdir.
		err = os.MkdirAll(cdir.thumbDir, 0755)
		if err != nil {
			return nil, err
		}
	} else {
		if !stat.Mode().IsDir() {
			return nil, fmt.Errorf(`Path %s exists, but is not a dirctory`, path)
		}
	}
	return cdir, nil
}

func (cdir *CacheDir) Get(ti *ThumbInfo) (string, error) {
	path := cdir.Locate(ti)
	if !cdir.Exists(ti) {
		err := ti.SaveAs(cdir.Locate(ti))
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

func (cdir *CacheDir) Locate(ti *ThumbInfo) string {
	return fmt.Sprintf("%s/%s/%s", cdir.thumbDir, ti.Id[:2], ti.Id[2:])
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
