package estelle

import (
	"fmt"
	"strings"
)

// Format represents the image file format.
type Format int

const (
	// FMT_UNKNOWN represents an unknown or unsupported image format.
	FMT_UNKNOWN Format = iota
	// FMT_JPG represents the JPEG image format.
	FMT_JPG
	// FMT_PNG represents the PNG image format.
	FMT_PNG
	// FMT_WEBP represents the WebP image format.
	FMT_WEBP
)

// String returns the string representation of the format.
func (f Format) String() string {
	switch f {
	case FMT_JPG:
		return "jpg"
	case FMT_PNG:
		return "png"
	case FMT_WEBP:
		return "webp"
	}
	panic(fmt.Sprintf("Unknow format type: %d", f))
}

// MimeType returns the MIME type corresponding to the format.
func (f Format) MimeType() string {
	switch f {
	case FMT_JPG:
		return "image/jpeg"
	case FMT_PNG:
		return "image/png"
	case FMT_WEBP:
		return "image/webp"
	}
	panic(fmt.Sprintf("Unknow format type: %d", f))
}

// FormatFromString parses a string and returns the corresponding Format.
func FormatFromString(s string) Format {
	switch strings.ToUpper(s) {
	case "JPG", "JPEG":
		return FMT_JPG
	case "PNG":
		return FMT_PNG
	case "WEBP":
		return FMT_WEBP
	}
	return FMT_UNKNOWN
}
