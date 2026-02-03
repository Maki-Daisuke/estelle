package estelle

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
)

type Hash [sha1.Size]byte

type Fingerprint struct {
	Path      string
	Size      int64
	MtimeSec  int64
	MtimeNsec int64
}

func HashFromFile(path string) (Hash, error) {
	fp, err := FingerprintFromFile(path)
	if err != nil {
		return Hash{}, err
	}
	return fp.Hash(), nil
}

func FingerprintFromFile(path string) (Fingerprint, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Fingerprint{}, err
	}
	fi, err := os.Stat(absPath)
	if err != nil {
		return Fingerprint{}, err
	}
	return Fingerprint{
		Path:      absPath,
		Size:      fi.Size(),
		MtimeSec:  fi.ModTime().Unix(),
		MtimeNsec: int64(fi.ModTime().Nanosecond()),
	}, nil
}

func (fp *Fingerprint) Hash() Hash {
	// Serialize fingerprint by joining fields with null bytes, which are not allowed in file paths.
	str := fmt.Sprintf("%s\x00%x\x00%x\x00%x", fp.Path, fp.Size, fp.MtimeSec, fp.MtimeNsec)
	return sha1.Sum([]byte(str))
}

func (h Hash) String() string {
	return fmt.Sprintf("%x", [sha1.Size]byte(h))
}
