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

func TestMotion_BlinkBreatheFrame(t *testing.T) {
	w := testWidget{name: "avatar", kind: widget.Region}
	sheet := style.For(w).
		Part("eyes", style.Blink()).
		Part("body", style.Breathe()).
		Part("a", style.Frame(0, 3, style.CadenceFast)).
		Part("o", style.Frame(1, 3, style.CadenceFast)).
		Part("rest", style.Frame(2, 3, style.CadenceFast))

	cssOut := sheet.Stylesheet().String()

	if !strings.Contains(cssOut, ".avatar__eyes {\n  transform-box: fill-box;\n  transform-origin: center;\n  animation: tw-blink 4s ease-in-out infinite;\n}") {
		if !strings.Contains(cssOut, "animation: tw-blink") {
			t.Errorf("Expected Blink output, got:\n%s", cssOut)
		}
	}
	if strings.Count(cssOut, "@keyframes tw-blink {") != 1 {
		t.Errorf("Expected exactly one @keyframes tw-blink in output")
	}

	if !strings.Contains(cssOut, ".avatar__body {\n  transform-box: fill-box;\n  transform-origin: 50% 100%;\n  animation: tw-breathe 4s ease-in-out infinite;\n}") {
		if !strings.Contains(cssOut, "animation: tw-breathe") {
			t.Errorf("Expected Breathe output, got:\n%s", cssOut)
		}
	}
	if strings.Count(cssOut, "@keyframes tw-breathe {") != 1 {
		t.Errorf("Expected exactly one @keyframes tw-breathe in output")
	}

	if !strings.Contains(cssOut, ".avatar__a {\n  animation-name: tw-frame-3;\n  animation-duration: 450ms;\n  animation-timing-function: step-end;\n  animation-delay: 0ms;\n  animation-iteration-count: infinite;\n}") {
		if !strings.Contains(cssOut, "animation-name: tw-frame-3") || !strings.Contains(cssOut, "delay: 0ms") {
			t.Errorf("Expected Frame a output, got:\n%s", cssOut)
		}
	}
	if !strings.Contains(cssOut, ".avatar__o {\n  animation-name: tw-frame-3;\n  animation-duration: 450ms;\n  animation-timing-function: step-end;\n  animation-delay: -300ms;\n  animation-iteration-count: infinite;\n}") {
		if !strings.Contains(cssOut, "animation-name: tw-frame-3") || !strings.Contains(cssOut, "delay: -300ms") {
			t.Errorf("Expected Frame o output, got:\n%s", cssOut)
		}
	}
	if !strings.Contains(cssOut, ".avatar__rest {\n  animation-name: tw-frame-3;\n  animation-duration: 450ms;\n  animation-timing-function: step-end;\n  animation-delay: -150ms;\n  animation-iteration-count: infinite;\n}") {
		if !strings.Contains(cssOut, "animation-name: tw-frame-3") || !strings.Contains(cssOut, "delay: -150ms") {
			t.Errorf("Expected Frame rest output, got:\n%s", cssOut)
		}
	}

	if strings.Count(cssOut, "@keyframes tw-frame-3 {") != 1 {
		t.Errorf("Expected exactly one @keyframes tw-frame-3 in output")
	}
	if !strings.Contains(cssOut, "33.333%, 100% { opacity: 0; }") {
		t.Errorf("Expected correct frame breakpoint for 3 frames, got:\n%s", cssOut)
	}

	expectedReducedAnimNone := `.avatar__a, .avatar__body, .avatar__eyes, .avatar__o, .avatar__rest {
  animation: none;
}`
	if !strings.Contains(cssOut, expectedReducedAnimNone) {
		t.Errorf("Expected animation: none block for all parts, got:\n%s", cssOut)
	}

	expectedReducedOpacityZero := `.avatar__o, .avatar__rest {
  opacity: 0;
}`
	if !strings.Contains(cssOut, expectedReducedOpacityZero) {
		t.Errorf("Expected opacity: 0 block for frames > 0, got:\n%s", cssOut)
	}
}

func TestMotion_FrameMultipleCounts(t *testing.T) {
	w := testWidget{name: "avatar", kind: widget.Region}
	sheet := style.For(w).
		Part("fast", style.Frame(0, 2, style.CadenceSlow)).
		Part("slow", style.Frame(0, 4, style.CadenceSlow))

	cssOut := sheet.Stylesheet().String()

	if !strings.Contains(cssOut, "animation-duration: 800ms") {
		t.Errorf("Expected duration 800ms for 2 slow frames")
	}
	if !strings.Contains(cssOut, "animation-duration: 1600ms") {
		t.Errorf("Expected duration 1600ms for 4 slow frames")
	}
	idx2 := strings.Index(cssOut, "@keyframes tw-frame-2 {")
	idx4 := strings.Index(cssOut, "@keyframes tw-frame-4 {")
	if idx2 == -1 || idx4 == -1 || idx2 > idx4 {
		t.Errorf("Expected tw-frame-2 before tw-frame-4 in output")
	}
}
