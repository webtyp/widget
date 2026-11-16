//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// HitArea adds the missing room as padding and takes it back as margin, so the
// layout and the drawn glyph do not change.
func TestHitAreaPadsAndCompensates(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Part(partAdd, style.HitArea()).Stylesheet().String()
	for _, want := range []string{"padding: max(0px, calc((24px - 1em) / 2));", "margin: min(0px, calc((1em - 24px) / 2));"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in\n%s", want, s)
		}
	}
}
