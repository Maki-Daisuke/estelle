package estelle

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
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
