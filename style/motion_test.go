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

func TestMotion_WavePulsePivot(t *testing.T) {
	w := testWidget{name: "avatar", kind: widget.Region}
	sheet := style.For(w).
		Part("hand", style.Wave()).
		Part("indicator", style.Pulse()).
		Part("base", style.Pivot("top left"), style.Rotate(style.TurnQuarter))

	cssOut := sheet.Stylesheet().String()

	if !strings.Contains(cssOut, ".avatar__base {\n  transform: rotate(90deg);\n  transform-box: fill-box;\n  transform-origin: top left;\n}") {
		t.Errorf("Expected Pivot and Rotate output, got:\n%s", cssOut)
	}

	if !strings.Contains(cssOut, ".avatar__hand {\n  transform-box: fill-box;\n  transform-origin: 50% 100%;\n  animation: tw-wave 1s ease-in-out infinite;\n}") {
		if strings.Contains(cssOut, "animation: tw-wave") {
			// alphabetical order formatting
			if !strings.Contains(cssOut, ".avatar__hand {\n  animation: tw-wave 1s ease-in-out infinite;\n  transform-box: fill-box;\n  transform-origin: 50% 100%;\n}") {
				t.Errorf("Expected Wave output, got:\n%s", cssOut)
			}
		} else {
			t.Errorf("Expected Wave output, got:\n%s", cssOut)
		}
	}
	if !strings.Contains(cssOut, "@keyframes tw-wave {") {
		t.Errorf("Expected @keyframes tw-wave in output")
	}

	if !strings.Contains(cssOut, ".avatar__indicator {\n  animation: tw-pulse 2s ease-in-out infinite;\n  transform-box: fill-box;\n  transform-origin: center;\n}") {
		if !strings.Contains(cssOut, "animation: tw-pulse 2s") {
			t.Errorf("Expected Pulse output, got:\n%s", cssOut)
		}
	}
	if !strings.Contains(cssOut, "@keyframes tw-pulse {") {
		t.Errorf("Expected @keyframes tw-pulse in output")
	}

	expectedReduced := `@media (prefers-reduced-motion: reduce) {
.avatar__hand, .avatar__indicator {
  animation: none;
}
}`
	if !strings.Contains(cssOut, expectedReduced) {
		t.Errorf("Expected reduced motion block for wave/pulse:\n%s\n\nGot:\n%s", expectedReduced, cssOut)
	}
}
