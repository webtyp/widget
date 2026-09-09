//go:build !wasm

package style_test

import (
	"strings"
	"testing"
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestSplitCollapses(t *testing.T) {
	// the emitted sheet contains no @container and no container-type; side by side / stacked (closes D-9)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.Split(style.SplitTwoThirds, style.Space2)).Stylesheet().String()

	if strings.Contains(s, "@container") {
		t.Error("Split must not emit @container rule")
	}
	if strings.Contains(s, "container-type") {
		t.Error("Split must not emit container-type declaration")
	}
}

func TestSplitRatioIsUnitlessForFlexGrow(t *testing.T) {
	// --ratio feeds flex-grow. An fr unit is invalid at computed-value time:
	// flex-grow silently resets to its initial 0 and the first partition
	// collapses to its content width instead of taking the larger share.
	wd := &testWidget{name: "w", kind: widget.Region}
	for _, tc := range []struct {
		ratio style.SplitRatio
		want  string
	}{
		{style.SplitHalf, "--ratio: 1;"},
		{style.SplitTwoThirds, "--ratio: 2;"},
		{style.SplitThreeQuarters, "--ratio: 3;"},
	} {
		s := style.For(wd).Root(style.Split(tc.ratio, style.Space2)).Stylesheet().String()
		if !strings.Contains(s, tc.want) {
			t.Errorf("expected %q, got:\n%s", tc.want, s)
		}
		if strings.Contains(s, "--ratio: 1fr;") || strings.Contains(s, "--ratio: 2fr;") || strings.Contains(s, "--ratio: 3fr;") {
			t.Errorf("--ratio must not carry an fr unit when it feeds flex-grow, got:\n%s", s)
		}
	}
}

func TestSurfaceCarriesNoPadding(t *testing.T) {
	// As(Panel) emits border-radius but no padding (closes C-2)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.As(style.Panel)).Stylesheet().String()

	if strings.Contains(s, "padding:") {
		t.Error("As(Panel) alone must not emit any padding")
	}
}

func TestDividerBelowEmitsBlockEndHairline(t *testing.T) {
	// DividerBelow is the block-end counterpart of Divider(side): a hairline
	// under a row (a drawer nav item) with no Surface of its own. Like the
	// other primitive flags it emits in device/state/cue rules (the base path
	// groups those flags into shared selectors instead).
	wd := &testWidget{name: "w", kind: widget.Region}
	out := style.For(wd).
		Root(style.Stack(style.Space1)).
		On(css.Mobile, widget.Part("item"), style.Row(style.Space1), style.DividerBelow()).
		Stylesheet().String()
	if !strings.Contains(out, "border-block-end: 1px solid ") {
		t.Errorf("DividerBelow must emit a border-block-end hairline, got:\n%s", out)
	}
	if strings.Contains(out, "border-inline-end: 1px solid ") || strings.Contains(out, "border-inline-start: 1px solid ") {
		t.Errorf("DividerBelow must not touch the inline edges, got:\n%s", out)
	}
}

func TestPadThenPadEdgeKeepsEmissionOrder(t *testing.T) {
	// C1 regression: formatRule used to sort declarations alphabetically
	// before deduplicating. "padding-block-end:" sorts before "padding:"
	// (- < : in ASCII) no matter which Option actually ran second, so Pad()
	// followed by PadEdge() — the "general value plus one edge exception"
	// form — silently came out with the shorthand LAST, winning over the
	// longhand it was supposed to lose to. Index check, not Contains: the
	// defect was one of order, and a Contains passes just the same when the
	// order is wrong.
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Part("box", style.Pad(style.Space1), style.PadEdge(style.EdgeBottom, style.Space12)).
		Stylesheet().String()

	padIdx := strings.Index(s, "padding: ")
	padEdgeIdx := strings.Index(s, "padding-block-end: ")
	if padIdx == -1 || padEdgeIdx == -1 {
		t.Fatalf("expected both padding: and padding-block-end: in the sheet, got:\n%s", s)
	}
	if padIdx > padEdgeIdx {
		t.Errorf("padding: must be emitted BEFORE padding-block-end: so the longhand overrides the shorthand, got:\n%s", s)
	}
}

func TestOverlappingFlexOptionsDedupToOneDeclaration(t *testing.T) {
	// Two Options emitting the same declaration on one rule must collapse to
	// one line. This keeps the set-based dedup in formatRule alive: the old
	// sort-then-drop handled the same case, and a "simplification" that
	// drops the dedup ships two identical declarations silently.
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Part("box", style.CenterContent(), style.StartContent()).
		Stylesheet().String()

	if n := strings.Count(s, "display: flex;"); n != 1 {
		t.Errorf("expected exactly one display: flex; declaration, got %d:\n%s", n, s)
	}
}
