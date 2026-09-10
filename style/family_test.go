//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// TestFamilySeparatedFromSurface is the consumer-shaped proof the publication
// rule demands: a real Sheet declares a resting look and an interaction
// family from two different surfaces, and the emitted CSS keeps both — the
// resting paint from one, the derived states from the other. Before the
// split, the second call silently overwrote the first and no test noticed
// (targethour's PartFree carried exactly this pair).
func TestFamilySeparatedFromSurface(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Part("row", style.As(style.Subtle), style.Interactive(style.Page)).
		Stylesheet().String()

	// resting look: Subtle's transparent background and muted text
	if !strings.Contains(s, "background-color: transparent;") {
		t.Errorf("expected the Subtle resting background (transparent), got:\n%s", s)
	}
	if !strings.Contains(s, "color: "+css.ColorMuted.LightValue()+";") {
		t.Errorf("expected the Subtle resting text (muted), got:\n%s", s)
	}

	// states: derived from the Page family, not from the resting surface
	for name, want := range map[string]string{
		"hover": css.Hover(css.ColorBackground),
		"focus": css.Focus(css.ColorBackground),
		"press": css.Press(css.ColorBackground),
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %s derivation from the Page family (%s), got:\n%s", name, want, s)
		}
	}
	if strings.Contains(s, css.Hover(css.ColorSurface)) {
		t.Errorf("states must not derive from the resting surface's family:\n%s", s)
	}
}

// TestSubtleStatesRepaintTextOnSurface: a Subtle resting surface keeps muted
// text at rest, but the derived state backgrounds are family colours — muted
// text on them measured 1.42:1 on hover. The states therefore repaint the
// text to on-surface (measured 9.68:1 with the ColorSurface hover fill).
func TestSubtleStatesRepaintTextOnSurface(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Part("ghost", style.As(style.Subtle), style.Interactive(style.Page)).
		Stylesheet().String()

	// the family base of Subtle is the neutral surface, never the muted text token
	if strings.Contains(s, css.Hover(css.ColorMuted)) {
		t.Errorf("Subtle states must not derive from the muted text token:\n%s", s)
	}

	// every derived state carries the on-surface text repaint (static + enhanced)
	wantStatic := "color: " + css.ColorOnSurface.LightValue() + ";"
	wantEnhanced := "color: " + css.ColorOnSurface.EnhancedVar() + ";"
	if strings.Count(s, wantStatic) < 3 {
		t.Errorf("expected on-surface static text in hover, focus and press, got:\n%s", s)
	}
	if strings.Count(s, wantEnhanced) < 3 {
		t.Errorf("expected on-surface enhanced text in hover, focus and press, got:\n%s", s)
	}
}

// TestIconBoxUnderIconCapRejected: IconCap sizes the glyph itself through a
// `> svg` child rule at (0,1,1), which silently wins over an IconBox class
// rule at (0,1,0) in the same layer. Both shapings of the mistake are loud:
// same part carrying both, and a glyph part nested (Within) inside a cap part.
func TestIconBoxUnderIconCapRejected(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}

	sheet := style.For(wd).Part("cap", style.As(style.Primary), style.IconCap(), style.IconBox(style.IconMd))
	found := false
	for _, err := range sheet.Validate() {
		if strings.Contains(err.Error(), "IconCap already sizes the glyph; drop IconBox") {
			found = true
		}
	}
	if !found {
		t.Errorf("same part with IconCap + IconBox must be rejected, got %v", sheet.Validate())
	}

	sheet2 := style.For(wd).
		Part("cap", style.As(style.Primary), style.IconCap()).
		Within("cap", "glyph", style.IconBox(style.IconMd))
	found = false
	for _, err := range sheet2.Validate() {
		if strings.Contains(err.Error(), "IconBox under IconCap") {
			found = true
		}
	}
	if !found {
		t.Errorf("IconBox glyph nested under an IconCap must be rejected, got %v", sheet2.Validate())
	}
}

// TestHandRolledCapRejected: MediaBox(AspectSquare) + ControlBox() sizes the
// box and says nothing about the glyph — the hand composition IconCap()
// replaces. The rule lives in the library that owns both options, so every
// consumer is covered, not just one repo's regex.
//
// Exempt: a square control-height block holding leading-edge TEXT
// (targetdate's PartLead: StartContent + Pad) is not a glyph cap — there is
// no <svg> for IconCap's child rule to size.
func TestHandRolledCapRejected(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}

	sheet := style.For(wd).Part("cap",
		style.As(style.Primary),
		style.MediaBox(style.AspectSquare),
		style.ControlBox(),
		style.KeepSize(),
	)
	found := false
	for _, err := range sheet.Validate() {
		if strings.Contains(err.Error(), "hand-rolled icon cap") && strings.Contains(err.Error(), "use IconCap()") {
			found = true
		}
	}
	if !found {
		t.Errorf("hand-rolled cap must be rejected with a pointer to IconCap(), got %v", sheet.Validate())
	}

	// the text block keeps validating: square + control height + StartContent
	textSheet := style.For(wd).Part("lead",
		style.StartContent(),
		style.MediaBox(style.AspectSquare),
		style.ControlBox(),
		style.Pad(style.Space2),
		style.KeepSize(),
	)
	for _, err := range textSheet.Validate() {
		if strings.Contains(err.Error(), "hand-rolled icon cap") {
			t.Errorf("text block with StartContent must stay valid, got: %v", err)
		}
	}
}
