package estelle

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ThumbInfo holds the information about a thumbnail.
type ThumbInfo struct {
	id     string // ID of this thumbnail (Fingerprint-Size-Mode.Format)
	path   string // Absolute path to thumbnail file
	source string // Absolute path to source file
	size   Size   // Size of this thumbnail
	mode   Mode   // Mode of this thumbnail
	format Format // File format (extension) of this thumbnail
}

// Keeps base directory path to generate ThumbInfo.
// This is for optimization, to avoid calling filepath.Abs() repeatedly.
type ThumbInfoFactory string

// BaseDir returns the base directory path.
func (dir ThumbInfoFactory) BaseDir() string {
	return string(dir)
}

// NewThumbInfoFactory creates a new ThumbInfoFactory.
// It checks if the base directory exists and is writable.
func NewThumbInfoFactory(baseDir string) (ThumbInfoFactory, error) {
	absPath, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	if stat, err := os.Stat(absPath); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		err = os.MkdirAll(absPath, 0755)
		if err != nil {
			return "", err
		}
	} else if !stat.IsDir() {
		return "", fmt.Errorf(`"%s" exists, but it is not a dirctory`, absPath)
	}
	temp, err := os.CreateTemp(absPath, "estelle-test-*")
	if err != nil {
		return "", fmt.Errorf("cache directory (%s) is not writable: %s", absPath, err)
	}
	temp.Close()
	os.Remove(temp.Name())
	return ThumbInfoFactory(absPath), nil
}

// FromFile creates a new ThumbInfo from the given path.
// It calculates the fingerprint of the source file and creates the thumbnail information.
func (dir ThumbInfoFactory) FromFile(path string, size Size, mode Mode, format Format) (ThumbInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ThumbInfo{}, err
	}
	fp, err := FingerprintFromFile(absPath)
	if err != nil {
		return ThumbInfo{}, err
	}
	hash := fp.Hash().String()
	id := fmt.Sprintf("%s-%s-%s.%s", hash, size, mode, format)
	return ThumbInfo{
		id:     id,
		source: absPath,
		path:   filepath.Join(string(dir), hash[:2], hash[2:4], id),
		size:   size,
		mode:   mode,
		format: format,
	}, nil
}

// String returns the ID of the thumbnail.
func (ti ThumbInfo) String() string {
	return ti.id
}

// Path returns the absolute path of the thumbnail.
func (ti ThumbInfo) Path() string {
	return ti.path
}

// Exists returns true if the thumbnail file exists and is a regular file.
func (ti ThumbInfo) Exists() bool {
	st, err := os.Stat(ti.path)
	if err != nil {
		return false
	}
	if !st.Mode().IsRegular() {
		return false
	}
	// Lazy Touch: Update timestamp if it's older than 1 hour.
	// This ensures that frequently accessed files are not collected by GC.
	now := time.Now()
	// Use GetAtime to get access time (platform dependent).
	// On Linux, this will return Atime. On Windows, it will fallback to ModTime.
	if now.Sub(GetAtime(st)) > 24*time.Hour {
		os.Chtimes(ti.path, now, now)
	}
	return true
}

func (ti ThumbInfo) Make() error {
	// Make sure that sharding directories (cachedir/XX/XX/) exist.
	dir := filepath.Dir(ti.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Generate the thumbnail using a temporary filename and rename it to the target name after completion.
	// This prevents incomplete (corrupted) thumbnail files from being recognized as valid.
	// We simply prepend "incomplete_" to the filename. This is sufficient to avoid conflicts
	// because Estelle.Enqueue() uses singleflight to ensure only one generation process runs at a time.
	tmpName := filepath.Join(dir, "incomplete_"+filepath.Base(ti.path))

	params := ti.prepareVipsArgs(tmpName)
	cmd := exec.Command("vipsthumbnail", params...)

	// Capture stderr for debugging
	stderr := bytes.NewBuffer([]byte{})
	cmd.Stderr = stderr

	err := cmd.Run() // block until the command completes.
	if err != nil {
		return fmt.Errorf("vipsthumbnail failed: %s: %w", stderr.String(), err)
	}

	if err := os.Rename(tmpName, ti.path); err != nil {
		return err
	}
	return nil
}

func (ti ThumbInfo) prepareVipsArgs(outputPath string) []string {
	// vipsthumbnail [flags] sourcefile -o outputfile
	args := []string{ti.source}

	sizeStr := ti.size.String()
	// Size logic
	// vipsthumbnail source.img --size WxH
	// ModeCrop: --smartcrop=attention
	// ModeShrink: default
	// ModeStretch: Postfix "!" to size
	switch ti.mode {
	case ModeCrop:
		args = append(args, "--smartcrop", "attention")
	case ModeStretch:
		sizeStr += "!"
	}
	args = append(args, "--size", sizeStr)

	args = append(args, "-o", outputPath)

	return args
}
