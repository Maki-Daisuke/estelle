package estelle

import (
	"fmt"
	"strings"
)

// Mode represents the resizing mode for thumbnail generation.
type Mode uint

const (
	// ModeUnknown represents an unknown or unsupported resizing mode.
	ModeUnknown Mode = iota
	// ModeCrop crops the image to fit the exact size, using smart cropping by attention.
	ModeCrop
	// ModeShrink shrinks the image to fit within the given boundaries without changing aspect ratio.
	ModeShrink
	// ModeStretch stretches the image to exactly match the requested dimensions, ignoring aspect ratio.
	ModeStretch
)

// ModeFromString parses a string and returns the corresponding Mode.
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

// String returns the string representation of the mode.
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
