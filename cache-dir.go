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
	this := &CacheDir{
		dir:      filepath.Clean(abs_path),
		thumbDir: filepath.Clean(abs_path + "/thumbs"),
	}
	stat, err := os.Stat(this.thumbDir)
	if err != nil {
		// Probably, it does not exist, then, try to mkdir.
		err = os.MkdirAll(this.thumbDir, 0755)
		if err != nil {
			return nil, err
		}
	} else {
		if !stat.Mode().IsDir() {
			return nil, fmt.Errorf(`Path %s exists, but is not a dirctory`, path)
		}
	}
	return this, nil
}

func (this *CacheDir) Get(ti *ThumbInfo) (string, error) {
	path := this.Locate(ti)
	if !this.Exists(ti) {
		err := ti.SaveAs(this.Locate(ti))
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

func (this *CacheDir) Locate(ti *ThumbInfo) string {
	hash := ti.Hash.String()
	return fmt.Sprintf("%s/%s/%s-%dx%d-%s.%s", this.thumbDir, hash[:2], hash[2:], ti.Width, ti.Height, ti.Mode, ti.Format)
}

func (this *CacheDir) Exists(ti *ThumbInfo) bool {
	_, err := os.Stat(this.Locate(ti))
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}
