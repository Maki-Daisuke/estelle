package estelle

import (
	"fmt"
	"testing"
)

func TestMode(t *testing.T) {
	m := ModeFill
	if fmt.Sprintf("%s", m) != "fill" {
		t.Errorf(`Expected "fill", but actual "%s"`, m)

	}
}
