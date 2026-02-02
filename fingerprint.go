package estelle

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

type Hash [sha256.Size]byte

type Fingerprint struct {
	Path      string
	Size      int64
	MtimeSec  int64
	MtimeNsec int64
}

func NewHashFromFile(path string) (Hash, error) {
	fp, err := NewFingerprint(path)
	if err != nil {
		return Hash{}, err
	}
	return fp.Hash(), nil
}

func NewFingerprint(path string) (*Fingerprint, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fi, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	return &Fingerprint{
		Path:      absPath,
		Size:      fi.Size(),
		MtimeSec:  fi.ModTime().Unix(),
		MtimeNsec: int64(fi.ModTime().Nanosecond()),
	}, nil
}

func (fp *Fingerprint) Hash() Hash {
	// Serialize fingerprint by joining fields with null bytes, which are not allowed in file paths.
	str := fmt.Sprintf("%s\x00%x\x00%x\x00%x", fp.Path, fp.Size, fp.MtimeSec, fp.MtimeNsec)
	return sha256.Sum256([]byte(str))
}

func (h Hash) String() string {
	return fmt.Sprintf("%x", [sha256.Size]byte(h))
}
