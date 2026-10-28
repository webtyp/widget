//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// An SVG shape ignores background-color and border: Paint must reach it through
// fill and stroke, from the same tokens As() gives a box.
func TestPaintFillsAndStrokesFromTheSurface(t *testing.T) {
	w := testWidget{name: "d", kind: widget.Region}
	s := style.For(w).Part("box", style.Paint(style.Panel)).Stylesheet().String()

	for _, want := range []string{"fill: " + css.ColorSurface.EnhancedVar() + ";", "stroke: " + css.ColorOutline.EnhancedVar() + ";"} {
		if !strings.Contains(s, want) {
			t.Errorf("Paint(Panel) lacks %q, got:\n%s", want, s)
		}
	}
	if strings.Contains(s, "background-color") {
		t.Errorf("Paint must not set a background an SVG shape ignores, got:\n%s", s)
	}
}

// SVG text is filled, not coloured: Ink gives it the surface's text colour.
func TestInkFillsTextWithTheSurfaceTextColour(t *testing.T) {
	w := testWidget{name: "d", kind: widget.Region}
	s := style.For(w).Part("name", style.Ink(style.Panel)).Stylesheet().String()

	if !strings.Contains(s, "fill: "+css.ColorOnSurface.EnhancedVar()+";") {
		t.Errorf("Ink(Panel) must fill with the on-surface colour, got:\n%s", s)
	}
}

// A line is a stroke with nothing filled: a curved path filled would paint the
// area under the curve. A solid role (Danger) has no border, so its line takes
// the role's own colour.
func TestLineStrokesWithoutFilling(t *testing.T) {
	w := testWidget{name: "d", kind: widget.Region}
	s := style.For(w).
		Part("edge", style.Line(style.Panel), style.Dashed()).
		Part("edge-error", style.Line(style.Danger)).
		Stylesheet().String()

	for _, want := range []string{"fill: none;", "stroke: " + css.ColorOutline.EnhancedVar() + ";", "stroke: " + css.ColorDanger.EnhancedVar() + ";", "stroke-dasharray:"} {
		if !strings.Contains(s, want) {
			t.Errorf("Line lacks %q, got:\n%s", want, s)
		}
	}
}

func TestPaintInkAndLineAreExclusive(t *testing.T) {
	w := testWidget{name: "d", kind: widget.Region}
	cases := map[string]*style.Sheet{
		"Paint+Ink":    style.For(w).Part("x", style.Paint(style.Panel), style.Ink(style.Panel)),
		"Paint+Line":   style.For(w).Part("x", style.Paint(style.Panel), style.Line(style.Panel)),
		"Dashed alone": style.For(w).Part("x", style.Dashed()),
	}
	for name, s := range cases {
		if errs := s.Validate(); len(errs) == 0 {
			t.Errorf("%s must fail Validate()", name)
		}
	}
}

// A document with no body to inherit from (a standalone SVG) takes the theme's
// typeface from its root, or the viewer falls back to its own serif default.
func TestTypefaceSetsTheThemeFont(t *testing.T) {
	w := testWidget{name: "d", kind: widget.Region}
	s := style.For(w).Root(style.Typeface()).Stylesheet().String()
	if !strings.Contains(s, "font-family: "+css.FontSans.Var()+";") {
		t.Errorf("Typeface must set the theme font, got:\n%s", s)
	}
}
