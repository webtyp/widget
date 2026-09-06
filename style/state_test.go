//go:build !wasm

package style_test

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
	"strings"
	"testing"
)

func TestCueWithinReachesADescendant(t *testing.T) {
	// The one descendant selector in the package. A rail that reveals its own
	// labels on hover has no other expression: Cue() only ever emits
	// `.n__part:hover`, and the label is a different element from the trigger.
	w := testWidget{name: "pd", kind: widget.Menu}
	s := style.For(w).
		Part("menu", style.Stack(style.Space1)).
		Part("link-text", style.FontSize(style.TextBase)).
		CueWithin(widget.Hover, "menu", "link-text", style.Row(style.SpaceNone)).
		Stylesheet().String()

	if !strings.Contains(s, ".pd__menu:hover .pd__link-text") {
		t.Errorf("expected the descendant selector, got:\n%s", s)
	}
}

func TestCueAcrossSpansAnArbitraryRelationship(t *testing.T) {
	// The escape hatch for what CueWithin cannot reach: a part reacting to a
	// cue on a region that is NOT its ancestor. Checked from the root via
	// :has(). Here: floating chrome hidden while the content region has focus
	// within it — and it must OUTRANK a RevealedBy reveal on the same part,
	// which the device path emits unlayered, so CueAcross is emitted unlayered
	// too (after the states layer closes).
	w := testWidget{name: "pd", kind: widget.Menu}
	s := style.For(w).
		Part("panel", style.Scroll()).
		Part("chrome", style.Raise(style.Floating), style.RevealedBy(widget.Open)).
		CueAcross(widget.FocusWithin, "panel", "chrome", style.Hide()).
		Stylesheet().String()

	yield := strings.Index(s, ".pd:has(.pd__panel:focus-within) .pd__chrome")
	if yield < 0 {
		t.Fatalf("expected the :has()-from-root selector, got:\n%s", s)
	}
	if !strings.Contains(s[yield:], "display: none;") {
		t.Errorf("expected Hide() to emit display:none, got:\n%s", s[yield:])
	}
	// The RevealedBy reveal is emitted unlayered; the yield rule must come
	// AFTER it in source order (and be more specific) so it wins the cascade.
	reveal := strings.Index(s, `.pd__chrome[data-open="true"]`)
	if reveal < 0 {
		t.Fatalf("expected the RevealedBy reveal rule, got:\n%s", s)
	}
	if yield < reveal {
		t.Errorf("CueAcross must be emitted AFTER the RevealedBy reveal so it overrides it; got yield@%d reveal@%d\n%s", yield, reveal, s)
	}
	// Not swallowed by a layer: no unclosed "@layer" between the last one and
	// the yield selector.
	tail := s[:yield]
	if la := strings.LastIndex(tail, "@layer "); la >= 0 && strings.Count(tail[la:], "{") > strings.Count(tail[la:], "}") {
		t.Errorf("the CueAcross rule must be unlayered, but sits inside an open @layer block:\n%s", s)
	}
}

func TestCueAcrossRejectsSameParts(t *testing.T) {
	w := testWidget{name: "pd", kind: widget.Menu}
	sheet := style.For(w).
		Part("panel", style.Scroll()).
		CueAcross(widget.FocusWithin, "panel", "panel", style.Hide())
	errs := sheet.Validate()
	if len(errs) == 0 {
		t.Fatal("expected CueAcross with region == part to be rejected")
	}
}

func TestStateAcrossProbesForADescendantState(t *testing.T) {
	// StateAcross is CueAcross for a WRITTEN state: the chrome yields while the
	// region CONTAINS an element carrying the state, deep inside — a record
	// being edited, not only a field with focus.
	w := testWidget{name: "pd", kind: widget.Menu}
	s := style.For(w).
		Part("panel", style.Scroll()).
		Part("chrome", style.Raise(style.Floating), style.RevealedBy(widget.Open)).
		StateAcross(widget.Open, "panel", "chrome", style.Hide()).
		Stylesheet().String()

	want := `.pd:has(.pd__panel [data-open="true"]) .pd__chrome`
	i := strings.Index(s, want)
	if i < 0 {
		t.Fatalf("expected the :has()-descendant-state selector %q, got:\n%s", want, s)
	}
	if !strings.Contains(s[i:], "display: none;") {
		t.Errorf("expected Hide() in the StateAcross rule, got:\n%s", s[i:])
	}
	// Unlayered and after the RevealedBy reveal, same as CueAcross.
	reveal := strings.Index(s, `.pd__chrome[data-open="true"]`)
	if reveal >= 0 && i < reveal {
		t.Errorf("StateAcross must come AFTER the RevealedBy reveal to override it\n%s", s)
	}
}

func TestCueWithinHoverScopesToFinePointer(t *testing.T) {
	// CueWithinHover is the same descendant selector as CueWithin, gated on
	// the fine-pointer capability: a touch tap fires :hover but never
	// (hover: hover), so a hover reveal scoped here cannot misfire on a phone.
	w := testWidget{name: "pd", kind: widget.Menu}
	s := style.For(w).
		Part("menu", style.Stack(style.Space1)).
		Part("drawer-panel", style.Raise(style.Floating)).
		CueWithinHover(widget.Hover, "menu", "drawer-panel", style.Row(style.SpaceNone)).
		Stylesheet().String()

	mediaIdx := strings.Index(s, "@media (hover: hover)")
	if mediaIdx < 0 {
		t.Fatalf("expected the fine-pointer gate, got:\n%s", s)
	}
	block := s[mediaIdx:]
	if !strings.Contains(block, ".pd__menu:hover .pd__drawer-panel") {
		t.Errorf("expected the descendant selector inside the hover media query, got:\n%s", s)
	}
	if !strings.Contains(block, "@layer states") {
		t.Errorf("expected the rule in @layer states (so it outranks @layer widgets device rules), got:\n%s", block)
	}
	plain := s[:mediaIdx]
	if strings.Contains(plain, ".pd__menu:hover .pd__drawer-panel") {
		t.Errorf("the selector must not escape into the plain states layer, got:\n%s", plain)
	}
}

func TestValidateCueWithinHoverContainerUndeclared(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}
	sheet := style.For(w).
		Part("link-text", style.FontSize(style.TextBase)).
		CueWithinHover(widget.Hover, "menu", "link-text", style.Row(style.SpaceNone))

	errs := sheet.Validate()
	if len(errs) == 0 {
		t.Fatal("expected Validate to reject an undeclared CueWithinHover container")
	}
	if !strings.Contains(errs[0].Error(), `CueWithinHover container "menu" is not a declared part`) {
		t.Errorf("unexpected error message: %v", errs[0])
	}
}

func TestStateAttrsListsEveryRevealedBy(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure}
	sheet := style.For(w).
		Part("panel", style.RevealedBy(widget.Open)).
		Part("tab", style.RevealedBy(widget.Current)).
		On(css.Mobile, "panel", style.RevealedBy(widget.Open))

	attrs := sheet.StateAttrs()
	if len(attrs) != 2 {
		t.Errorf("expected 2 state attrs, got %d: %v", len(attrs), attrs)
	}
	found := make(map[string]bool)
	for _, a := range attrs {
		found[a.Key+"="+a.Value] = true
	}
	if !found["data-open=true"] {
		t.Error("expected data-open=true in StateAttrs")
	}
	if !found["data-current=true"] {
		t.Error("expected data-current=true in StateAttrs")
	}
}

func TestRevealedSplitStaysFlex(t *testing.T) {
	// Split emits display:flex in @layer primitives; a state rule saying "grid"
	// would win from @layer states and strand every flex-basis under it.
	w := testWidget{name: "w", kind: widget.Disclosure}
	s := style.For(w).
		Part("panel", style.Split(style.SplitTwoThirds, style.Space2), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(s, ".w__panel[data-open=\"true\"] {\n  display: flex;\n}") {
		t.Errorf("expected a revealed Split to stay flex, got:\n%s", s)
	}
}
