package estelle

import (
	"fmt"
	"strings"
)

type Mode uint

const (
	ModeUnknown Mode = iota
	ModeCrop
	ModeShrink
	ModeStretch
)

func ModeFromString(s string) Mode {
	switch strings.ToLower(s) {
	case "crop":
		return ModeCrop
	case "shrink":
		return ModeShrink
	case "stretch":
		return ModeStretch
	default:
		return ModeUnknown
	}
}

func (m Mode) String() string {
	switch m {
	case ModeCrop:
		return "crop"
	case ModeShrink:
		return "shrink"
	case ModeStretch:
		return "stretch"
	default:
		panic(fmt.Sprintf("unknown Mode value (%d)", m))
	}
}
