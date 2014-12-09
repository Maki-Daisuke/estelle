package estelle

import "fmt"

type Mode uint

const (
	ModeFill   Mode = iota + 1
	ModeFit    Mode = iota + 1
	ModeShrink Mode = iota + 1
)

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
