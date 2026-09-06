//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestMotion_AnimateSlow(t *testing.T) {
	w := testWidget{name: "panel", kind: widget.Region}
	sheet := style.For(w).Root(style.Animate(style.MotionSlow))
	cssOut := sheet.Stylesheet().String()

	if !strings.Contains(cssOut, "var(--duration-slow") || !strings.Contains(cssOut, "var(--ease-in-out") {
		t.Errorf("Expected slow transition declaration using CSS variables, got:\n%s", cssOut)
	}

	if strings.Contains(cssOut, "transition: all 400ms") || strings.Contains(cssOut, "transition: 400ms") {
		t.Errorf("Transition should not contain literal 400ms duration, got:\n%s", cssOut)
	}
}

func TestMotion_AnimateNone(t *testing.T) {
	w := testWidget{name: "panel", kind: widget.Region}
	sheet := style.For(w).Root(style.Animate(style.MotionNone))
	cssOut := sheet.Stylesheet().String()

	if !strings.Contains(cssOut, "transition: none;") {
		t.Errorf("Expected 'transition: none;', got:\n%s", cssOut)
	}
}

func TestMotion_NoAnimate(t *testing.T) {
	w := testWidget{name: "panel", kind: widget.Region}
	sheet := style.For(w)
	cssOut := sheet.Stylesheet().String()

	if strings.Contains(cssOut, "prefers-reduced-motion") {
		t.Errorf("Expected no prefers-reduced-motion block when no motion is defined, got:\n%s", cssOut)
	}
	if strings.Contains(cssOut, "transition:") {
		t.Errorf("Expected no transition declarations, got:\n%s", cssOut)
	}
}

func TestMotion_OnlyInWhen(t *testing.T) {
	w := testWidget{name: "panel", kind: widget.Disclosure} // Disclosure allows Open state
	sheet := style.For(w).When(widget.Open, "", style.Animate(style.MotionSlow))
	cssOut := sheet.Stylesheet().String()

	if !strings.Contains(cssOut, "@media (prefers-reduced-motion: reduce) {") {
		t.Fatalf("Expected prefers-reduced-motion block, got:\n%s", cssOut)
	}

	expectedBlock := ".panel[data-open=\"true\"] {\n  transition: none;\n}"
	if !strings.Contains(cssOut, expectedBlock) {
		t.Errorf("Expected reduced motion selector block:\n%s\n\nGot:\n%s", expectedBlock, cssOut)
	}
}

func TestMotion_Determinism(t *testing.T) {
	w := testWidget{name: "panel", kind: widget.Disclosure}
	s1 := style.For(w).Part("item", style.As(style.Panel)).Root(style.Animate(style.MotionSlow)).When(widget.Open, "", style.Animate(style.MotionBase)).Cue(widget.Hover, "item", style.Animate(style.MotionFast))
	s2 := style.For(w).Part("item", style.As(style.Panel)).Root(style.Animate(style.MotionSlow)).When(widget.Open, "", style.Animate(style.MotionBase)).Cue(widget.Hover, "item", style.Animate(style.MotionFast))

	out1 := s1.Stylesheet().String()
	out2 := s2.Stylesheet().String()

	if out1 != out2 {
		t.Errorf("Stylesheet generation is non-deterministic.\n\nCSS 1:\n%s\n\nCSS 2:\n%s", out1, out2)
	}
}
