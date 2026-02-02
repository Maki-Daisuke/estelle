package estelle

import (
	"fmt"
	"strings"
)

type Format int

const (
	FMT_UNKNOWN Format = iota
	FMT_JPG
	FMT_PNG
	FMT_WEBP
)

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

func FormatFromString(s string) (Format, error) {
	switch strings.ToUpper(s) {
	case "JPG", "JPEG":
		return FMT_JPG, nil
	case "PNG":
		return FMT_PNG, nil
	case "WEBP":
		return FMT_WEBP, nil
	}
	return FMT_UNKNOWN, fmt.Errorf("Unsupported image format: %v", s)
}
