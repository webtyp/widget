//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestAppBar_BaseAndReveal(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}

	s := style.For(w).
		Root(style.AppBar(widget.Open, style.MotionBase)).
		Stylesheet().
		String()

	for _, want := range []string{
		"display: flex;",
		"align-items: center;",
		"flex-shrink: 0;",
		"block-size: calc(var(--control-height",
		"2 * var(--space-1",
		"padding-inline: var(--space-3",
		"border-radius: 0;",
		"margin-block-start: calc(-1 * calc(var(--control-height",
		"transition: margin-block-start var(--duration-base",
		"var(--ease-in-out",
		`.w[data-open="true"] {`,
		"margin-block-start: 0;",
		"@media (prefers-reduced-motion: reduce) {",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected AppBar output to contain %q, got:\n%s", want, s)
		}
	}
}

func TestAppBar_MotionNone(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}

	s := style.For(w).
		Root(style.AppBar(widget.Open, style.MotionNone)).
		Stylesheet().
		String()

	if strings.Contains(s, "transition: margin-block-start") {
		t.Errorf("MotionNone must not emit transition, got:\n%s", s)
	}
	if !strings.Contains(s, "margin-block-start: calc(-1 * calc(var(--control-height") {
		t.Errorf("MotionNone still retracts by negative margin, got:\n%s", s)
	}
	if !strings.Contains(s, "margin-block-start: 0;") {
		t.Errorf("MotionNone still resets margin on reveal, got:\n%s", s)
	}
}

func TestAppBar_DevicePath(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}

	s := style.For(w).
		On(css.Mobile, widget.Part("bar"), style.AppBar(widget.Open, style.MotionFast)).
		Stylesheet().
		String()

	if !strings.Contains(s, "@media (max-width: 639.98px)") {
		t.Fatalf("expected mobile media query, got:\n%s", s)
	}
	if !strings.Contains(s, ".w__bar {") {
		t.Errorf("expected .w__bar in mobile media block, got:\n%s", s)
	}
	if !strings.Contains(s, `.w__bar[data-open="true"] {`) {
		t.Errorf("expected .w__bar[data-open=\"true\"] in mobile media block, got:\n%s", s)
	}
	if !strings.Contains(s, "margin-block-start: 0;") {
		t.Errorf("expected reveal margin in mobile media block, got:\n%s", s)
	}
}

func TestAppBar_ValidationForbiddenPairings(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}

	// 1. With ControlBox
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for AppBar + ControlBox")
			}
		}()
		style.For(w).Root(style.AppBar(widget.Open, style.MotionBase), style.ControlBox()).Stylesheet()
	}()

	// 2. With Docked
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for AppBar + Docked")
			}
		}()
		style.For(w).Root(style.AppBar(widget.Open, style.MotionBase), style.Docked(style.Viewport, style.EdgeTop, style.SideEnd, style.Space3)).Stylesheet()
	}()

	// 3. With RevealedBy
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for AppBar + RevealedBy")
			}
		}()
		style.For(w).Root(style.AppBar(widget.Open, style.MotionBase), style.RevealedBy(widget.Open)).Stylesheet()
	}()

	// 4. Disallowed state for widget kind
	// Region only allows SpanFull; Open is disallowed.
	wRegion := testWidget{name: "reg", kind: widget.Region}
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for disallowed AppBar state")
			}
		}()
		style.For(wRegion).Root(style.AppBar(widget.Open, style.MotionBase)).Stylesheet()
	}()
}
