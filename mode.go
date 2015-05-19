package estelle

import (
	"fmt"
	"strings"
)

type Mode uint

const (
	ModeFill   Mode = iota + 1
	ModeFit    Mode = iota + 1
	ModeShrink Mode = iota + 1
)

func ModeFromString(s string) (Mode, error) {
	switch strings.ToLower(s) {
	case "fill":
		return ModeFill, nil
	case "fit":
		return ModeFit, nil
	case "shrink":
		return ModeShrink, nil
	default:
		return ModeFill, fmt.Errorf("invalid mode string")
	}
}

func (m Mode) String() string {
	switch m {
	case ModeFill:
		return "fill"
	case ModeFit:
		return "fit"
	case ModeShrink:
		return "shrink"
	default:
		panic(fmt.Sprintf("unknown Mode value (%d)", m))
	}
}
