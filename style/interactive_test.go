//go:build !wasm

package style_test

import (
	"strings"
	"testing"
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestInteractiveDerivesFamily(t *testing.T) {
	// Interactive(Primary) emits the three Primary states and no other family (closes D-4)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.Interactive(style.Primary)).Stylesheet().String()

	for name, want := range map[string]string{
		"hover": css.Hover(css.ColorPrimary),
		"focus": css.Focus(css.ColorPrimary),
		"press": css.Press(css.ColorPrimary),
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected primary %s derivation to be emitted: %s", name, want)
		}
	}

	// no other family leaks in
	if strings.Contains(s, css.Hover(css.ColorSuccess)) {
		t.Error("should not contain another family's derivation")
	}
}

func TestFocusVisible(t *testing.T) {
	// the focus cue emits :focus-visible; bare :focus appears nowhere (closes D-4)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Cue(widget.Focus, "", style.As(style.Primary)).Stylesheet().String()

	if !strings.Contains(s, ":focus-visible") {
		t.Error("expected :focus-visible to be emitted for Focus cue")
	}
	if strings.Contains(s, ":focus ") || strings.HasSuffix(s, ":focus {\n") {
		t.Error("bare :focus must not be emitted")
	}
}

// A simple parser to extract all var(...) calls properly handling nested parenthesis

func TestInteractiveRejectsNonInteractive(t *testing.T) {
	// Interactive(Inactive) is reported: it is the deliberately-dead shade,
	// and interacting with it is always a mistake (closes D-3)
	wd := &testWidget{name: "w", kind: widget.Region}
	sheet := style.For(wd).Root(style.Interactive(style.Inactive))
	if len(sheet.Validate()) == 0 {
		t.Error("Interactive(Inactive) should be reported as invalid")
	}
}

func TestInteractivePageIsLegalAndWhite(t *testing.T) {
	// Interactive(Page) validates and derives a family from the whitest
	// surface: the white page background plus the cursor: pointer an
	// interactive rule carries (closes 0.2 — white-and-clickable was not
	// expressible).
	wd := &testWidget{name: "w", kind: widget.Region}
	sheet := style.For(wd).Root(style.Interactive(style.Page))
	if errs := sheet.Validate(); len(errs) != 0 {
		t.Fatalf("Interactive(Page) must validate, got: %v", errs)
	}

	s := sheet.Stylesheet().String()
	if !strings.Contains(s, "cursor: pointer;") {
		t.Errorf("expected Interactive(Page) to emit cursor: pointer, got:\n%s", s)
	}
	if !strings.Contains(s, "background-color: "+css.ColorBackground.LightValue()+";") {
		t.Errorf("expected Interactive(Page) to emit the white page background (%s), got:\n%s", css.ColorBackground.LightValue(), s)
	}
	// the family derives from the page background, not from nothing
	if !strings.Contains(s, css.Hover(css.ColorBackground)) {
		t.Errorf("expected the hover derivation from ColorBackground, got:\n%s", s)
	}
}
