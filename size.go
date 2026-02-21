package estelle

import (
	"fmt"
	"regexp"
	"strconv"
)

// Size represents the dimensions (width and height) of an image.
type Size struct {
	Width, Height uint
}

// SizeFromUint creates a Size from given width and height.
func SizeFromUint(w, h uint) Size {
	return Size{w, h}
}

var regexpSize, _ = regexp.Compile("([0-9]+)(?:[^0-9]+([0-9]+))?")

// SizeFromString parses a string into a Size object.
// The string can be in formats like "100x200" or just "100" (which implies 100x100).
func SizeFromString(s string) (Size, error) {
	m := regexpSize.FindStringSubmatch(s)
	if m != nil {
		w, _ := strconv.ParseUint(m[1], 10, 32)
		if m[2] != "" {
			h, _ := strconv.ParseUint(m[2], 10, 32)
			return Size{uint(w), uint(h)}, nil
		} else {
			return Size{uint(w), uint(w)}, nil
		}
	} else {
		return Size{}, fmt.Errorf("SizeFromString: can't parse string %v", s)
	}
}

// String returns the string representation of the size (e.g., "100x200").
func (s Size) String() string {
	return fmt.Sprintf("%dx%d", s.Width, s.Height)
}
