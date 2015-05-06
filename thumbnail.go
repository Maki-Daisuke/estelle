package estelle

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type ThumbInfo struct {
	Id     string
	path   string
	Hash   Hash
	Width  uint
	Height uint
	Mode   Mode
	Format string
}

func NewThumbInfoFromFile(path string, width, height uint, mode Mode, format string) (*ThumbInfo, error) {
	hash, err := NewHashFromFile(path)
	if err != nil {
		return nil, err
	}
	return &ThumbInfo{
		Id:     fmt.Sprintf("%s-%dx%d-%s.%s", hash, width, height, mode, format),
		path:   path,
		Hash:   hash,
		Width:  width,
		Height: height,
		Mode:   mode,
		Format: format,
	}, nil
}

func (ti *ThumbInfo) String() string {
	return ti.Id
}

func (this *ThumbInfo) SaveAs(savePath string) error {
	dir := filepath.Dir(savePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	params := this.prepareMagickArgs(savePath)
	cmd := exec.Command("convert", params...)
	cmd.Stdout = ioutil.Discard
	stderr := bytes.NewBuffer([]byte{})
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
}

func (this *ThumbInfo) prepareMagickArgs(out string) []string {
	args := []string{this.path}
	switch this.Mode {
	case ModeFill:
		geometry := fmt.Sprintf("%dx%d", this.Width, this.Height)
		args = append(args,
			"-resize", geometry,
			"-background", "white",
			"-gravity", "center",
			"-extent", geometry,
		)
	case ModeFit:
		resize := fmt.Sprintf("%dx%d^", this.Width, this.Height)
		extent := fmt.Sprintf("%dx%d", this.Width, this.Height)
		args = append(args,
			"-resize", resize,
			"-gravity", "center",
			"-extent", extent,
		)
	case ModeShrink:
		geometry := fmt.Sprintf("%dx%d", this.Width, this.Height)
		args = append(args,
			"-resize", geometry,
		)
	default:
		panic(fmt.Sprintf("unknown resize mode (%d)", this.Mode))
	}
	args = append(args, fmt.Sprintf("%s:%s", this.Format, out)) // explicitly specify image format
	return args
}
