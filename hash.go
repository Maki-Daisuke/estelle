package estelle

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"regexp"
)

type Hash [sha1.Size]byte

func NewHashFromReader(reader io.Reader) (Hash, error) {
	hash := sha1.New()
	_, err := io.Copy(hash, reader)
	if err != nil {
		return Hash{}, err
	}
	var sum [sha1.Size]byte
	for i, x := range hash.Sum(nil) {
		sum[i] = x
	}
	id := Hash(sum)
	return id, nil
}

func NewHashFromFile(path string) (Hash, error) {
	file, err := os.Open(path)
	if err != nil {
		return Hash{}, err
	}
	return NewHashFromReader(file)
}

func (id Hash) String() string {
	return fmt.Sprintf("%x", [sha1.Size]byte(id))
}

var reHash *regexp.Regexp = regexp.MustCompile(fmt.Sprintf("^[0-9a-fA-F]{%d}", sha1.Size))

func NewHashFromString(s string) (Hash, error) {
	m := reHash.FindStringSubmatch(s)
	if m == nil {
		return Hash{}, fmt.Errorf("not a hash string")
	}

	var buf [sha1.Size]byte
	for i := 0; i < sha1.Size; i++ {
		_, e := fmt.Sscanf(s, "%02x", &buf[i])
		if e != nil {
			return Hash{}, e
		}
		s = s[2:]
	}
	return Hash(buf), nil
}
