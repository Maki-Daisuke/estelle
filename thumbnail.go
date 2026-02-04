package estelle

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
)

type ThumbInfo struct {
	id     string // ID of this thumbnail (Fingerprint-Size-Mode.Format)
	source string // Absolute path to source file
	hash   Hash   // Hash of the source file
	size   Size   // Size of this thumbnail
	mode   Mode   // Mode of this thumbnail
	format Format // File format (extension) of this thumbnail
}

func ThumbInfoFromFile(path string, size Size, mode Mode, format Format) (ThumbInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ThumbInfo{}, err
	}
	hash, err := HashFromFile(absPath)
	if err != nil {
		return ThumbInfo{}, err
	}
	return ThumbInfo{
		id:     fmt.Sprintf("%s-%s-%s.%s", hash, size, mode, format),
		source: absPath,
		hash:   hash,
		size:   size,
		mode:   mode,
		format: format,
	}, nil
}

func (ti ThumbInfo) String() string {
	return ti.id
}

func (ti ThumbInfo) Make(out io.WriteCloser) error {
	params := ti.prepareMagickArgs()
	cmd := exec.Command("convert", params...)
	cmd.Stdout = out
	defer out.Close()
	stderr := bytes.NewBuffer([]byte{})
	cmd.Stderr = stderr
	err := cmd.Run() // block until the command completes.
	if err != nil {
		return fmt.Errorf("%s", stderr.String())
	}
	return nil
}

func (ti ThumbInfo) prepareMagickArgs() []string {
	args := []string{ti.source}
	switch ti.mode {
	case ModeFill:
		args = append(args,
			"-resize", ti.size.String(),
			"-background", "white",
			"-gravity", "center",
			"-extent", ti.size.String(),
		)
	case ModeFit:
		args = append(args,
			"-resize", ti.size.String()+"^",
			"-gravity", "center",
			"-extent", ti.size.String(),
		)
	case ModeShrink:
		args = append(args,
			"-resize", ti.size.String(),
		)
	default:
		panic(fmt.Sprintf("unknown resize mode (%d)", ti.mode))
	}
	args = append(args, ti.format.String()+":-") // explicitly specify image format
	return args
}
