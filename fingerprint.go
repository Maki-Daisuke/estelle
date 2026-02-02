package estelle

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
)

type Hash [sha256.Size]byte

type Fingerprint struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	MtimeSec  int64  `json:"mtime_sec"`
	MtimeNsec int64  `json:"mtime_nsec"`
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
	// Serialize fingerprint to JSON to ensure deterministic hashing of the struct fields
	data, _ := json.Marshal(fp)
	return sha256.Sum256(data)
}

func (h Hash) String() string {
	return fmt.Sprintf("%x", [sha256.Size]byte(h))
}

func HashFromString(s string) (Hash, error) {
	if len(s) != sha256.Size*2 {
		return Hash{}, fmt.Errorf("invalid hash length: %d", len(s))
	}
	var h Hash
	for i := 0; i < sha256.Size; i++ {
		_, err := fmt.Sscanf(s[i*2:i*2+2], "%02x", &h[i])
		if err != nil {
			return Hash{}, err
		}
	}
	return h, nil
}
