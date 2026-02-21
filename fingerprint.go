package estelle

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
)

// Hash represents a SHA-1 hash sum.
type Hash [sha1.Size]byte

// fingerprint holds file metadata used to uniquely identify a source image state.
type fingerprint struct {
	Path      string
	Size      int64
	MtimeSec  int64
	MtimeNsec int64
}

// fingerprintFromFile generates a fingerprint for the file at the given path.
func fingerprintFromFile(path string) (fingerprint, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fingerprint{}, err
	}
	fi, err := os.Stat(absPath)
	if err != nil {
		return fingerprint{}, err
	}
	return fingerprint{
		Path:      absPath,
		Size:      fi.Size(),
		MtimeSec:  fi.ModTime().Unix(),
		MtimeNsec: int64(fi.ModTime().Nanosecond()),
	}, nil
}

// Hash returns the SHA-1 hash of the fingerprint to be used as a cache key.
func (fp *fingerprint) Hash() Hash {
	// Serialize fingerprint by joining fields with null bytes, which are not allowed in file paths.
	str := fmt.Sprintf("%s\x00%x\x00%x\x00%x", fp.Path, fp.Size, fp.MtimeSec, fp.MtimeNsec)
	return sha1.Sum([]byte(str))
}

// String returns the hexadecimal string representation of the Hash.
func (h Hash) String() string {
	return fmt.Sprintf("%x", [sha1.Size]byte(h))
}
