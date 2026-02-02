package estelle

import (
	"testing"
)

func TestHash(t *testing.T) {
	const fileName = "tests/IMG_20141207_201549.jpg"

	// First calculation
	id1, err := NewHashFromFile(fileName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// Second calculation (should be identical)
	id2, err := NewHashFromFile(fileName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if id1 != id2 {
		t.Errorf("Hash inconsistency. First: %s, Second: %s", id1, id2)
	}

	// Verify length (SHA256 hex string is 64 chars)
	if len(id1.String()) != 64 {
		t.Errorf("Invalid hash length: %d (expected 64)", len(id1.String()))
	}
}
