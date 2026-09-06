//go:build !wasm

package style_test

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
	"strings"
	"testing"
)

func TestFloatingChromeReservesScrollEndSpace(t *testing.T) {
	// A floating chrome strip (the host: a FAB, a miniplayer) declares the
	// band of its edge it occupies as an inherited custom property; every
	// Scroll() region — in this widget or a descendant one — reserves it
	// through var(--floating-bottom, 0px). No chrome means no declaration and
	// no reservation.
	w := testWidget{name: "fab", kind: widget.Region}
	s := style.For(w).
		Part("host", style.FloatingChrome(style.EdgeBottom, style.IconLg, style.Space4)).
		Part("list", style.Scroll()).
		Stylesheet().String()

	want := "--floating-bottom: calc(2.5em + 2 * " + css.Space4.Var() + ");"
	if !strings.Contains(s, want) {
		t.Errorf("expected the host to declare its strip with %q, got:\n%s", want, s)
	}
	for _, want := range []string{
		"padding-block-start: var(--floating-top, 0px);",
		"padding-block-end: var(--floating-bottom, 0px);",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Scroll() to reserve the floating strip with %q, got:\n%s", want, s)
		}
	}

	// EdgeTop is the mirror: the strip declaration switches property, the
	// Scroll() padding stays the same.
	top := style.For(w).Part("host", style.FloatingChrome(style.EdgeTop, style.IconLg, style.Space4)).Stylesheet().String()
	if !strings.Contains(top, "--floating-top: calc(2.5em + 2 * "+css.Space4.Var()+");") {
		t.Errorf("expected FloatingChrome(EdgeTop) to declare --floating-top, got:\n%s", top)
	}

	// And with no chrome at all, Scroll() pads by the 0px default — the var
	// references stay, but no sheet declares the property.
	bare := style.For(w).Part("list", style.Scroll()).Stylesheet().String()
	if strings.Contains(bare, "--floating-top:") || strings.Contains(bare, "--floating-bottom:") {
		t.Errorf("no FloatingChrome means no floating declaration, got:\n%s", bare)
	}
}

func TestScrollGutterAddsToTheFloatingReservationWithoutReplacingIt(t *testing.T) {
	// ScrollGutter exists so a consumer (webtyp/components/listgap) can give
	// a Scroll() region its own ambient top/bottom gutter WITHOUT the plain
	// widgets-layer PadEdge()/Pad() trap: a later-layer declaration replaces
	// an earlier one outright (CSS layers do not add), which would silently
	// erase whatever a FloatingChrome ancestor reserved. Folding both into one
	// calc is what keeps them both live at once.
	w := testWidget{name: "fab", kind: widget.Region}

	// With an active FloatingChrome reservation, the gutter must be ADDED to
	// it, not replace it.
	s := style.For(w).
		Part("host", style.FloatingChrome(style.EdgeBottom, style.IconLg, style.Space4)).
		Part("list", style.Scroll(), style.ScrollGutter(style.Space1)).
		Stylesheet().String()

	want := "padding-block-end: calc(var(--floating-bottom, 0px) + " + css.Space1.Var() + ");"
	if !strings.Contains(s, want) {
		t.Errorf("expected the gutter to add to the floating reservation with %q, got:\n%s", want, s)
	}
	// The top edge carries the SAME gutter even though only EdgeBottom has an
	// active FloatingChrome — the gutter is ambient on both edges by design.
	wantTop := "padding-block-start: calc(var(--floating-top, 0px) + " + css.Space1.Var() + ");"
	if !strings.Contains(s, wantTop) {
		t.Errorf("expected the gutter on the top edge too with %q, got:\n%s", wantTop, s)
	}

	// A Scroll() region with no gutter must keep emitting the exact old decl —
	// byte-identical — so every existing consumer is unaffected.
	bare := style.For(w).Part("list", style.Scroll()).Stylesheet().String()
	if !strings.Contains(bare, "padding-block-end: var(--floating-bottom, 0px);") {
		t.Errorf("a plain Scroll() must keep the old decl untouched, got:\n%s", bare)
	}
	if strings.Contains(bare, "calc(var(--floating-bottom") {
		t.Errorf("a plain Scroll() must never emit the calc form, got:\n%s", bare)
	}
}
