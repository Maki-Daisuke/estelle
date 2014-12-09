package estelle

import (
	"os"
	"testing"
)

func TestHash(t *testing.T) {
	const expected = "2b4656041c1922391a04c3e08e6ed362ebf902ca"
	file, err := os.Open("tests/IMG_20141207_201549.jpg")
	if err != nil {
		t.Errorf("%s", err)
	}
	id, err := NewHashFromReader(file)
	if err != nil {
		t.Errorf("%s", err)
	}
	if id.String() != expected {
		t.Errorf("Hash value mismatch. Expected: %s, actual: %s", expected, id.String())
	}
}
