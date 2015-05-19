package estelle

import (
	"fmt"
	"regexp"
	"strconv"
)

type Size struct {
	Width, Height uint
}

func SizeFromUint(w, h uint) Size {
	return Size{w, h}
}

var regexpSize, _ = regexp.Compile("([0-9]+)(?:[^0-9]+([0-9]+))?")

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

func (s Size) String() string {
	return fmt.Sprintf("%dx%d", s.Width, s.Height)
}
