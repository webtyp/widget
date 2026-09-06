//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

type testWidget struct {
	name widget.Name
	kind widget.Kind
}

func (t testWidget) WidgetName() widget.Name { return t.name }
func (t testWidget) WidgetKind() widget.Kind { return t.kind }

func TestBackdrop(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Dialog}

	// 1. style.For(w).Root(Backdrop(Viewport)) emits position: fixed and inset: 0 and z-index: var(--z-modal,300);
	s1 := style.For(w).Root(style.Backdrop(style.Viewport)).Stylesheet().String()
	if !strings.Contains(s1, "position: fixed;") {
		t.Errorf("expected s1 to contain 'position: fixed;', got:\n%s", s1)
	}
	if !strings.Contains(s1, "inset: 0;") {
		t.Errorf("expected s1 to contain 'inset: 0;', got:\n%s", s1)
	}
	if !strings.Contains(s1, "z-index: var(--z-modal,300);") {
		t.Errorf("expected s1 to contain 'z-index: var(--z-modal,300);', got:\n%s", s1)
	}

	// 2. style.For(w).Root(Backdrop(Parent)) emits position: absolute (no fixed)
	s2 := style.For(w).Root(style.Backdrop(style.Parent)).Stylesheet().String()
	if !strings.Contains(s2, "position: absolute;") {
		t.Errorf("expected s2 to contain 'position: absolute;', got:\n%s", s2)
	}
	if strings.Contains(s2, "position: fixed;") {
		t.Errorf("expected s2 NOT to contain 'position: fixed;', got:\n%s", s2)
	}

	// 3. style.For(w).Root(Veil(), Backdrop(Viewport)) emits a background-color with color-mix
	s4 := style.For(w).Root(style.Veil(), style.Backdrop(style.Viewport)).Stylesheet().String()
	if !strings.Contains(s4, "background-color:") {
		t.Errorf("expected s4 to contain 'background-color:', got:\n%s", s4)
	}
	if !strings.Contains(s4, "color-mix(") {
		t.Errorf("expected s4 to contain 'color-mix(', got:\n%s", s4)
	}
	// Not var(--color-surface: the color-mix() argument is
	// css.ColorSurface.NestedEnhanced(), a literal light-dark(...) with no
	// var() anywhere — a var() here, even to an always-safe property, would
	// defer this whole declaration's validity to computed-value time and
	// break the legacy-Safari fallback (see css.Token.NestedEnhanced).
	if !strings.Contains(s4, "light-dark(") {
		t.Errorf("expected s4 to contain 'light-dark(', got:\n%s", s4)
	}

	// 4. RevealedBy case
	s5 := style.For(w).
		Part(widget.Part("p"), style.Backdrop(style.Viewport), style.RevealedBy(widget.Open)).
		Stylesheet().
		String()

	idxWidgets := strings.Index(s5, "@layer widgets {")
	idxStates := strings.Index(s5, "@layer states {")
	if idxWidgets == -1 {
		t.Error("expected to find @layer widgets")
	}
	if idxStates == -1 {
		t.Error("expected to find @layer states")
	}
	if idxWidgets >= idxStates {
		t.Errorf("expected @layer widgets to come before @layer states in:\n%s", s5)
	}

	widgetsPart := s5[idxWidgets:idxStates]
	statesPart := s5[idxStates:]

	if !strings.Contains(widgetsPart, ".w__p {") || !strings.Contains(widgetsPart, "display: none;") {
		t.Errorf("expected .w__p with display: none; in @layer widgets part, got:\n%s", widgetsPart)
	}

	if !strings.Contains(statesPart, ".w__p[data-open=\"true\"] {") || !strings.Contains(statesPart, "display: block;") {
		t.Errorf("expected .w__p[data-open=\"true\"] with display: block; in @layer states part, got:\n%s", statesPart)
	}
}

// EdgeStrip spans the full block axis of its Scope plus one inline edge, and
// — the property that differentiates it from Drawer() — never forces a
// width. It also carries no RevealedBy() requirement: unlike Drawer, it is
// permanent chrome, not a toggled panel.
func TestEdgeStripSpansFullEdgeWithNoForcedWidth(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	sheet := style.For(w).
		Part("next", style.EdgeStrip(style.Parent, style.SideEnd))

	if errs := sheet.Validate(); len(errs) > 0 {
		t.Fatalf("EdgeStrip alone (no RevealedBy) must validate, got: %v", errs)
	}

	out := sheet.Stylesheet().String()
	if !strings.Contains(out, "position: absolute;") {
		t.Errorf("expected Parent-scoped EdgeStrip to emit position: absolute;, got:\n%s", out)
	}
	if !strings.Contains(out, "inset-block: 0;") {
		t.Errorf("expected inset-block: 0;, got:\n%s", out)
	}
	if !strings.Contains(out, "inset-inline-end: 0;") {
		t.Errorf("expected EdgeStrip(SideEnd) to emit inset-inline-end: 0;, got:\n%s", out)
	}
	if strings.Contains(out, "inset-inline-start:") {
		t.Errorf("expected exactly one inline inset, not both, got:\n%s", out)
	}
	if strings.Contains(out, "width:") {
		t.Errorf("EdgeStrip must never force a width — that is what differentiates it from Drawer(), got:\n%s", out)
	}
}

// Viewport-scoped EdgeStrip is position: fixed and claims the widget's own
// stacking layer, the same split Docked/Backdrop already draw between local
// chrome and a real overlay.
func TestEdgeStripViewportScopeIsFixed(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	s := style.For(w).Root(style.EdgeStrip(style.Viewport, style.SideStart)).Stylesheet().String()

	if !strings.Contains(s, "position: fixed;") {
		t.Errorf("expected Viewport-scoped EdgeStrip to emit position: fixed;, got:\n%s", s)
	}
	if !strings.Contains(s, "inset-inline-start: 0;") {
		t.Errorf("expected EdgeStrip(SideStart) to emit inset-inline-start: 0;, got:\n%s", s)
	}
}

// TestRevealedByWinsOverCenterContent: a part with BOTH RevealedBy and
// CenterContent must end hidden — CenterContent opens a block with
// display: flex, and a flex declaration emitted AFTER the reveal's
// display: none would resurrect a part that is supposed to stay parked
// until its state arrives. The last display: wins in CSS, so the hiding
// declaration has to be the very last one in the @layer widgets block.
// Regression: the reveal block used to sit before CenterContent, and the
// platform's hamburger (RevealedBy + CenterContent) never hid on scroll.
func TestRevealedByWinsOverCenterContent(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	s := style.For(w).
		Part(widget.Part("p"), style.RevealedBy(widget.Open), style.CenterContent()).
		Stylesheet().
		String()

	// The .w__p rule lives in @layer widgets; the reveal state lives in
	// @layer states. The LAST display declaration inside the widgets block
	// decides what the part looks like while parked.
	idxWidgets := strings.Index(s, "@layer widgets {")
	idxStates := strings.Index(s, "@layer states {")
	if idxWidgets == -1 || idxStates == -1 {
		t.Fatalf("expected @layer widgets and @layer states, got:\n%s", s)
	}
	widgetsPart := s[idxWidgets:idxStates]

	if !strings.Contains(widgetsPart, ".w__p {") {
		t.Fatalf("expected .w__p in @layer widgets part, got:\n%s", widgetsPart)
	}
	ruleBlock := widgetsPart[strings.Index(widgetsPart, ".w__p {"):]
	if !strings.Contains(ruleBlock, "display: none;") {
		t.Errorf("expected .w__p to carry display: none; while parked, got:\n%s", ruleBlock)
	}
	if !strings.Contains(ruleBlock, "display: flex;") {
		t.Errorf("expected .w__p to carry CenterContent's display: flex, got:\n%s", ruleBlock)
	}
	// display: none must come AFTER display: flex inside the block.
	flexIdx := strings.Index(ruleBlock, "display: flex;")
	noneIdx := strings.Index(ruleBlock, "display: none;")
	if noneIdx < flexIdx {
		t.Errorf("expected display: none; to come AFTER display: flex;, got:\n%s", ruleBlock)
	}
}
