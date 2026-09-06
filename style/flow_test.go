//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// FixedGrid's column count never reflows, unlike Grid()'s auto-fit/minmax —
// it always reads --cols, which the rule declares as a value (never a
// literal repeat(N, 1fr)) so a stylesheet builder can set a sane default on
// a zero-value receiver and the host overrides the count per instance.
func TestFixedGridEmitsColsVariable(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	s := style.For(w).Root(style.FixedGrid(3, style.SpaceNone)).Stylesheet().String()

	if !strings.Contains(s, "--cols: 3;") {
		t.Errorf("expected --cols: 3; to be declared, got:\n%s", s)
	}
	if !strings.Contains(s, "grid-template-columns: repeat(var(--cols), minmax(0, 1fr));") {
		t.Errorf("expected the fixed-column track function reading var(--cols), got:\n%s", s)
	}
	if !strings.Contains(s, "display: grid;") {
		t.Errorf("expected display: grid;, got:\n%s", s)
	}
}

// The minmax(0, ...) floor is what stops a fixed column from blowing out its
// row on a narrow viewport — a bare 1fr track has no minimum and lets long
// unbreakable content (a date, a number) push the column past its share.
func TestFixedGridTrackHasZeroFloor(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	s := style.For(w).Root(style.FixedGrid(7, style.Space1)).Stylesheet().String()

	if strings.Contains(s, "repeat(var(--cols), 1fr)") {
		t.Errorf("FixedGrid's track must be minmax(0, 1fr), not a bare 1fr, got:\n%s", s)
	}
}

// A flow type used only inside On()/OnlyOn() has nothing to group with on
// the main emission path and is emitted from the device-scoped switch
// instead — a primitive missing from that switch is silently dropped
// (emit.go's own warning comment). This is the regression test for that
// failure mode, for FixedGrid specifically.
func TestFixedGridSurvivesDeviceScope(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	s := style.For(w).
		Part("strip", style.Stack(style.Space2)).
		On(css.Mobile, "strip", style.FixedGrid(1, style.SpaceNone)).
		Stylesheet().String()

	if !strings.Contains(s, "@media (max-width: 639.98px)") {
		t.Fatalf("expected a device-scoped block, got:\n%s", s)
	}
	mobileBlock := s[strings.Index(s, "@media (max-width: 639.98px)"):]
	if !strings.Contains(mobileBlock, "grid-template-columns: repeat(var(--cols), minmax(0, 1fr));") {
		t.Errorf("FixedGrid used inside On() must still emit its track function, got:\n%s", mobileBlock)
	}
	if !strings.Contains(mobileBlock, "--cols: 1;") {
		t.Errorf("FixedGrid used inside On() must still declare --cols, got:\n%s", mobileBlock)
	}
}

// ScrollRow must carry scroll-behavior:smooth so a same-page anchor link
// into one of its children slides instead of jumping — the mechanism a
// JS-free single-card carousel (prev/next as <a href="#childID">) depends on.
func TestScrollRowIsSmooth(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	s := style.For(w).Root(style.ScrollRow(style.Space2)).Stylesheet().String()

	if !strings.Contains(s, "scroll-behavior: smooth;") {
		t.Errorf("expected ScrollRow to emit scroll-behavior: smooth;, got:\n%s", s)
	}
}
