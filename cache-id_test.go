//go:build linux

// 'cause, this requires 'touch' command.

package estelle

import (
	"os/exec"
	"testing"
	"time"
)

func TestHash(t *testing.T) {
	const fileName = "tests/IMG_20141207_201549.jpg"

	id1, err := fingerprintFromFile(fileName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	time.Sleep(10 * time.Millisecond)
	if err := exec.Command("touch", fileName).Run(); err != nil {
		t.Fatalf("%s", err)
	}

	id2, err := fingerprintFromFile(fileName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if id1 == id2 {
		t.Errorf("fingerprint is not changed after touch")
	}

	// Verify length (SHA1 hex string is 40 chars)
	if len(id1.Hash().String()) != 40 {
		t.Errorf("Invalid fingerprint length: %d (expected 40)", len(id1.Hash().String()))
	}
}
