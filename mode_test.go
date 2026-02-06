package estelle

import (
	"testing"
)

func TestMode(t *testing.T) {
	tests := []struct {
		m    Mode
		want string
	}{
		{ModeCrop, "crop"},
		{ModeShrink, "shrink"},
		{ModeStretch, "stretch"},
	}

	for _, tt := range tests {
		if got := tt.m.String(); got != tt.want {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.m, got, tt.want)
		}
	}

	// Test default behavior fallback (though ModeFromString returns error, the default matches crop)
	m := ModeFromString("invalid")
	if m != ModeUnknown {
		t.Errorf("Expected ModeUnknown for invalid mode strings")
	}
}
