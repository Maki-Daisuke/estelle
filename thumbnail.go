package estelle

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
)

type ThumbInfo struct {
	id     string // ID of this thumbnail
	source string // Absolute path to source file
	hash   Hash   // Hash of the source file
	size   Size   // Size of this thumbnail
	mode   Mode   // Mode of this thumbnail
	format Format // File format (extension) of this thumbnail
}

func NewThumbInfoFromFile(path string, size Size, mode Mode, format Format) (*ThumbInfo, error) {
	hash, err := NewHashFromFile(path)
	if err != nil {
		return nil, err
	}
	return &ThumbInfo{
		id:     fmt.Sprintf("%s-%s-%s.%s", hash, size, mode, format),
		source: path,
		hash:   hash,
		size:   size,
		mode:   mode,
		format: format,
	}, nil
}

var regexpId = regexp.MustCompile("([^-]+)-([^-]+)-([^.]+)\\.([^.]+)")

func NewThumbInfoFromId(id string) (*ThumbInfo, error) {
	m := regexpId.FindStringSubmatch(id)
	if m == nil {
		return nil, fmt.Errorf("invalid ID string: %s", id)
	}
	hash, err := NewHashFromString(m[1])
	if err != nil {
		return nil, err
	}
	size, err := SizeFromString(m[2])
	if err != nil {
		return nil, err
	}
	mode, err := ModeFromString(m[3])
	if err != nil {
		return nil, err
	}
	format, err := FormatFromString(m[4])
	if err != nil {
		return nil, err
	}
	return &ThumbInfo{
		id:     m[0],
		source: "",
		hash:   hash,
		size:   size,
		mode:   mode,
		format: format,
	}, nil
}

func (ti *ThumbInfo) String() string {
	return ti.id
}

func (ti *ThumbInfo) Id() string {
	return ti.id
}

func (ti *ThumbInfo) Source() string {
	return ti.source
}

func (ti *ThumbInfo) Hash() Hash {
	return ti.hash
}

func (ti *ThumbInfo) Size() Size {
	return ti.size
}

func (ti *ThumbInfo) Mode() Mode {
	return ti.mode
}

func (ti *ThumbInfo) Format() Format {
	return ti.format
}

func (ti *ThumbInfo) ETag() string {
	return `"` + ti.id + `"`
}

func (ti *ThumbInfo) CanMake() bool {
	return ti.source != ""
}

func (ti *ThumbInfo) Make(out io.WriteCloser) error {
	if !ti.CanMake() {
		return NewNoSourceError(ti)
	}
	params := ti.prepareMagickArgs()
	cmd := exec.Command("convert", params...)
	cmd.Stdout = out
	defer out.Close()
	stderr := bytes.NewBuffer([]byte{})
	cmd.Stderr = stderr
	err := cmd.Run() // block until the command completes.
	if err != nil {
		return fmt.Errorf(stderr.String())
	}
	return nil
}

func (ti *ThumbInfo) prepareMagickArgs() []string {
	args := []string{ti.Source()}
	switch ti.Mode() {
	case ModeFill:
		args = append(args,
			"-resize", ti.Size().String(),
			"-background", "white",
			"-gravity", "center",
			"-extent", ti.Size().String(),
		)
	case ModeFit:
		args = append(args,
			"-resize", ti.Size().String()+"^",
			"-gravity", "center",
			"-extent", ti.Size().String(),
		)
	case ModeShrink:
		args = append(args,
			"-resize", ti.Size().String(),
		)
	default:
		panic(fmt.Sprintf("unknown resize mode (%d)", ti.Mode()))
	}
	args = append(args, ti.Format().String()+":-") // explicitly specify image format
	return args
}

type InvalidIdError struct {
	error
}

func NewInvalidIdError(e error) InvalidIdError {
	return InvalidIdError{e}
}

type NoSourceError struct {
	*ThumbInfo
}

func NewNoSourceError(ti *ThumbInfo) NoSourceError {
	return NoSourceError{ti}
}

func (e NoSourceError) Error() string {
	return fmt.Sprintf("this ThumbInfo does not have source file: %s", e.Id())
}
