//go:build !wasm

package style_test

import (
	"strings"
	"testing"
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestGlyphTintsWithoutFilling(t *testing.T) {
	// A nav item that is merely available shows a coloured icon; only the
	// selected one gets the filled surface. Icons follow currentColor, so the
	// tint has to reach fill as well as color.
	w := testWidget{name: "pd", kind: widget.Menu}
	s := style.For(w).Part("nav-link", style.Glyph(style.Primary)).Stylesheet().String()

	if !strings.Contains(s, "fill: currentColor;") {
		t.Errorf("expected Glyph to reach the icon through fill, got:\n%s", s)
	}
	if strings.Contains(s, "background-color") {
		t.Errorf("Glyph must not paint a background, got:\n%s", s)
	}
}

func TestControlBoxSharesOneHeight(t *testing.T) {
	// A list row and a form field have to be measured against the same token or
	// they drift apart the moment either one's padding changes.
	w := testWidget{name: "x", kind: widget.Region}
	s := style.For(w).
		Part("row", style.ControlBox()).
		Part("field", style.ControlBox()).
		Stylesheet().String()

	if !strings.Contains(s, "min-height: "+css.ControlHeight.Var()+";") {
		t.Errorf("expected ControlBox to emit the shared height token, got:\n%s", s)
	}
}

func TestStateSurfaceRepaintsBorderAsRing(t *testing.T) {
	// A state is painted OVER the base box: a border here grows the element
	// exactly when the pointer is on it, and the hover is what made it grow.
	// The border is therefore repainted as a shadow ring — 0 0 0 1px, the
	// border width a hair's breadth outside the box — never as a border, and
	// never as an outline (see the ring comment in emit_decls.go for why).
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Part("x", style.As(style.Page)).
		Cue(widget.Hover, "x", style.As(style.Inset)).
		Stylesheet().String()

	idx := strings.Index(s, ".w__x:hover")
	if idx < 0 {
		t.Fatalf("expected a :hover rule, got:\n%s", s)
	}
	hoverRule := s[idx : idx+strings.Index(s[idx:], "}")]

	if !strings.Contains(hoverRule, "box-shadow: 0 0 0 1px") {
		t.Errorf("expected the state rule to repaint the border as a ring, got:\n%s", hoverRule)
	}
	if strings.Contains(hoverRule, "border:") {
		t.Errorf("a state rule must not emit border:, got:\n%s", hoverRule)
	}
	// "outline:" with the colon, not the bare word: the ring's colour is
	// var(--color-outline), and matching the token NAME would flag the very
	// declaration this test wants.
	if strings.Contains(hoverRule, "outline:") {
		t.Errorf("a state rule must not emit outline, got:\n%s", hoverRule)
	}
}

func TestStateRingComposesWithElevation(t *testing.T) {
	// Raise() and a state border both paint through box-shadow. When a state
	// rule raises AND repaints its border, the two must compose into ONE
	// declaration — ring first, elevation after — or the later declaration
	// would silently stomp the earlier one and the raised border would come
	// out bare.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Part("x", style.As(style.Page), style.Raise(style.Raised)).
		Cue(widget.Hover, "x", style.As(style.Inset), style.Raise(style.Raised)).
		Stylesheet().String()

	idx := strings.Index(s, ".w__x:hover")
	if idx < 0 {
		t.Fatalf("expected a :hover rule, got:\n%s", s)
	}
	hoverRule := s[idx : idx+strings.Index(s[idx:], "}")]

	// The ring is the token's Var() form when it composes with an elevation
	// (see boxShadowDecls): the light-dark() half would defer the whole
	// declaration to computed-value time, so the static/enhanced pair is
	// structurally impossible here.
	want := "box-shadow: 0 0 0 1px " + css.ColorOutline.Var() + ", " + css.ShadowSm.Var() + ";"
	if !strings.Contains(hoverRule, want) {
		t.Errorf("expected the raised state rule to compose ring then elevation in one declaration:\nwant %q\ngot:\n%s", want, hoverRule)
	}
	if strings.Contains(hoverRule, "box-shadow: "+css.ShadowSm.Var()+";") {
		t.Errorf("a raised state rule must not repeat a bare elevation declaration, got:\n%s", hoverRule)
	}
}

func TestBaseSurfaceKeepsBorder(t *testing.T) {
	// Only STATE rules switch to outline; the base box still owns its border.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.As(style.Inset)).Stylesheet().String()

	if !strings.Contains(s, "border: 1px solid") {
		t.Errorf("expected the base rule to emit border:, got:\n%s", s)
	}
	if strings.Contains(s, "outline:") {
		t.Errorf("a base rule must not emit outline, got:\n%s", s)
	}
}

func TestChipBoxSharesOneWidth(t *testing.T) {
	w := testWidget{name: "x", kind: widget.Region}
	s := style.For(w).Part("badge", style.ChipBox()).Stylesheet().String()

	if !strings.Contains(s, "width: "+css.ChipWidth.Var()+";") {
		t.Errorf("expected ChipBox to emit the shared width token, got:\n%s", s)
	}
	if !strings.Contains(s, "overflow: hidden;") {
		t.Errorf("expected ChipBox to clip, got:\n%s", s)
	}
}

func TestVeilBlursWhatIsBehindIt(t *testing.T) {
	w := testWidget{name: "dlg", kind: widget.Dialog}
	s := style.For(w).Part("backdrop", style.Backdrop(style.Viewport), style.Veil()).Stylesheet().String()

	for _, want := range []string{
		"backdrop-filter: blur(" + css.VeilBlur.Var() + ");",
		"-webkit-backdrop-filter: blur(" + css.VeilBlur.Var() + ");",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Veil to blur with %q, got:\n%s", want, s)
		}
	}
}

func TestPrimarySurface_GradientHookIsInertByDefaultAndLiveWhenSet(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("cta", style.As(style.Primary)).Stylesheet().String()

	if !strings.Contains(s, "background-color: "+css.ColorPrimary.EnhancedVar()+";") {
		t.Errorf("expected the solid background-color to still be emitted, got:\n%s", s)
	}
	if !strings.Contains(s, "background-image: var(--color-primary-image, none);") {
		t.Errorf("expected the always-present, inert-by-default gradient hook, got:\n%s", s)
	}

	themed := css.Theme(
		css.Set(css.ColorPrimary, "#16a34a"),
		css.SetGradient(css.ColorPrimary, "135deg", css.ColorPrimary, css.ColorAccent),
	).String()

	if !strings.Contains(themed, "--color-primary-image: linear-gradient(135deg, var(--color-primary") {
		t.Errorf("expected the app's Theme() to set the same --color-primary-image the widget rule reads, got:\n%s", themed)
	}
}

// TestGradientAngle_RepaintsFamilyGradientPerSurface: GradientAngle on an
// As(<family>) surface adds a second background-image that re-angles the theme
// gradient using the app's own stops (css ImageStopsVarName), emitted AFTER the
// plain hook so it wins when the app set a gradient and is silently ignored
// (invalid) when it did not.

func TestGradientAngle_RepaintsFamilyGradientPerSurface(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("rail", style.As(style.Primary), style.GradientAngle("315deg")).Stylesheet().String()

	// plain hook still there...
	if !strings.Contains(s, "background-image: var(--color-primary-image, none);") {
		t.Errorf("GradientAngle must not remove the inert family hook, got:\n%s", s)
	}
	// ...and the re-angled override comes AFTER it, referencing the stops.
	iHook := strings.Index(s, "background-image: var(--color-primary-image, none);")
	iAngle := strings.Index(s, "background-image: linear-gradient(315deg, var(--color-primary-image-stops));")
	if iAngle == -1 {
		t.Fatalf("GradientAngle must emit the re-angled override referencing --color-primary-image-stops, got:\n%s", s)
	}
	if iAngle < iHook {
		t.Errorf("the re-angled override must come AFTER the plain hook (fallback order), got:\n%s", s)
	}

	// A derived surface has no family gradient — GradientAngle is a no-op there.
	d := style.For(w).Part("x", style.As(style.AccentInverse), style.GradientAngle("315deg")).Stylesheet().String()
	if strings.Contains(d, "linear-gradient(315deg") {
		t.Errorf("GradientAngle on a derived surface must be a no-op, got:\n%s", d)
	}
}
