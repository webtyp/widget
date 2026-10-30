//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// SpaceNone summed inside calc() must be a <length>. A bare 0 is a <number>
// there, the whole calc() is invalid and the browser drops the declaration: a
// Docked(Viewport, EdgeTop, …, SpaceNone) + DockedClear(Space12) toast stack
// lost its inset-block-start and covered the app bar it was meant to clear.
func TestSpaceNoneInsideCalcIsALength(t *testing.T) {
	w := testWidget{name: "cx", kind: widget.Region}
	s := style.For(w).
		Part("stack",
			style.Docked(style.Viewport, style.EdgeTop, style.SideEnd, style.SpaceNone),
			style.DockedClear(style.Space12),
		).
		Part("host", style.FloatingChrome(style.EdgeBottom, style.IconLg, style.SpaceNone)).
		Part("list", style.Scroll(), style.ScrollGutter(style.SpaceNone)).
		Stylesheet().String()

	for _, bad := range []string{"calc(0 +", "+ 0)", "+ 2 * 0)"} {
		if strings.Contains(s, bad) {
			t.Errorf("SpaceNone emitted a unitless 0 inside calc() (%q), got:\n%s", bad, s)
		}
	}
	if !strings.Contains(s, "inset-block-start: calc(0px + var(--space-12") {
		t.Errorf("expected DockedClear to add to a 0px gap, got:\n%s", s)
	}
}
