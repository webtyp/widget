//go:build !wasm

package style_test

import (
	"strings"
	"testing"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestSlideDeckParksPagesOffCanvas(t *testing.T) {
	// SlideDeck is not a scroller: each child is an absolute layer parked at the
	// inline-start edge, and the current one is the only one in the box.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.SlideDeck(style.MotionBase)).Stylesheet().String()

	for _, want := range []string{
		".w__x > *",
		"position: absolute;",
		"inset: 0;",
		"transform: translateX(-100%);",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected SlideDeck pages to emit %q, got:\n%s", want, s)
		}
	}

	// Exactly three rules for the part: container, pages, current page.
	if n := strings.Count(s, ".w__x "); n != 3 {
		t.Errorf("expected exactly 3 rules for a SlideDeck part, got %d:\n%s", n, s)
	}
}

func TestSlideDeckRevealsTheCurrentPage(t *testing.T) {
	// The state is widget.Current, derived from its Attr(), so markup and CSS
	// match by construction — the same principle that sustains RevealedBy.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.SlideDeck(style.MotionBase)).Stylesheet().String()

	if !strings.Contains(s, `.w__x > *[data-current="true"]`) {
		t.Errorf("expected the current-page selector, got:\n%s", s)
	}
	if !strings.Contains(s, "transform: translateX(0);") {
		t.Errorf("expected the current page to sit in the box, got:\n%s", s)
	}
}

func TestSlideDeckIsNotAScroller(t *testing.T) {
	// The whole point: no scroller here, or the gesture inside a module chains
	// with the stage's own snap strip and changes section alone.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.SlideDeck(style.MotionBase)).Stylesheet().String()

	if strings.Contains(s, "scroll-snap-type") {
		t.Errorf("SlideDeck must not emit scroll-snap-type, got:\n%s", s)
	}
	if strings.Contains(s, "overflow-x") {
		t.Errorf("SlideDeck must not emit overflow-x, got:\n%s", s)
	}
}

func TestAutoRotateRunsOnZeroValueReceiver(t *testing.T) {
	// The package-wide contract: RenderCSS runs on &T{}, so AutoRotate cannot
	// take a count and must never panic or need instance data to build its
	// stylesheet.
	w := testWidget{name: "w", kind: widget.Region}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AutoRotate panicked on a zero-value receiver: %v", r)
		}
	}()
	s := style.For(w).Part("x", style.AutoRotate()).Stylesheet().String()
	if !strings.Contains(s, "@keyframes tw-auto-rotate") {
		t.Errorf("expected the shared keyframes rule, got:\n%s", s)
	}
}

func TestAutoRotateStagersDelayByPosition(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.AutoRotate()).Stylesheet().String()

	for _, want := range []string{
		".w__x > *",
		"position: absolute;",
		"inset: 0;",
		"opacity: 0;",
		"animation: tw-auto-rotate",
		".w__x > :first-child {\n  opacity: 1;\n}",
		".w__x > :nth-child(2) {\n  animation-delay: 5s;\n}",
		".w__x > :nth-child(6) {\n  animation-delay: 25s;\n}",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected AutoRotate to emit %q, got:\n%s", want, s)
		}
	}

	// No 7th slot: AutoRotateLayers caps the stagger.
	if strings.Contains(s, ":nth-child(7)") {
		t.Errorf("expected no :nth-child(7) rule beyond AutoRotateLayers, got:\n%s", s)
	}
}

func TestAutoRotateRespectsReducedMotion(t *testing.T) {
	// Disabling the animation must leave :first-child visible and every other
	// layer hidden — the "first image, frozen" fallback — with no JS and no
	// extra markup from the caller.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.AutoRotate()).Stylesheet().String()

	if !strings.Contains(s, "@media (prefers-reduced-motion: reduce)") {
		t.Fatalf("expected a prefers-reduced-motion block, got:\n%s", s)
	}
	if !strings.Contains(s, ".w__x > * {\n  animation: none;\n}") {
		t.Errorf("expected AutoRotate layers to disable their animation under reduced motion, got:\n%s", s)
	}
}
