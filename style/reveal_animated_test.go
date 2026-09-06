//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// Animate paired with RevealedBy on the same rule upgrades the instant
// display swap into a choreographed fade — entry and exit alike. MotionNone,
// or no Animate at all, keeps the instant swap.
func TestAnimatedReveal_Part(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure} // Disclosure allows Open
	out := style.For(w).
		Part("p", style.Animate(style.MotionBase), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	for _, want := range []string{
		"display: none;",
		"opacity: 0;",
		"transition: all var(--duration-base",
		"display var(--duration-base",
		"allow-discrete;",
		".w__p[data-open=\"true\"]",
		"opacity: 1;",
		"@starting-style",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected animated reveal to contain %q, got:\n%s", want, out)
		}
	}
}

func TestAnimatedReveal_MotionNoneStaysInstant(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure}
	out := style.For(w).
		Part("p", style.Animate(style.MotionNone), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(out, "display: none;") {
		t.Errorf("expected the hidden base, got:\n%s", out)
	}
	for _, want := range []string{"opacity:", "allow-discrete", "@starting-style"} {
		if strings.Contains(out, want) {
			t.Errorf("expected MotionNone reveal NOT to contain %q, got:\n%s", want, out)
		}
	}
}

func TestAnimatedReveal_NoAnimateStaysInstant(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure}
	out := style.For(w).
		Part("p", style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(out, "display: none;") {
		t.Errorf("expected the hidden base, got:\n%s", out)
	}
	for _, want := range []string{"opacity:", "allow-discrete", "@starting-style", "transition:"} {
		if strings.Contains(out, want) {
			t.Errorf("expected plain RevealedBy NOT to contain %q, got:\n%s", want, out)
		}
	}
}

func TestAnimatedReveal_DevicePath(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure}
	out := style.For(w).
		Part("p", style.Row(style.Space2)).
		OnlyOn(css.Mobile, "p", style.Animate(style.MotionBase), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	media := strings.Index(out, "@media")
	if media == -1 {
		t.Fatalf("expected a device media block, got:\n%s", out)
	}
	dev := out[media:]
	for _, want := range []string{
		"opacity: 0;",
		"allow-discrete;",
		"opacity: 1;",
		"@starting-style",
	} {
		if !strings.Contains(dev, want) {
			t.Errorf("expected device reveal to contain %q, got:\n%s", dev, out)
		}
	}
}

func TestOnlyOnKeepsBaseHideAfterPart(t *testing.T) {
	// Base options read before device ones — the usual order — so OnlyOn
	// must merge its hide into the rule Part() already created, not skip it.
	// Dropping it paints the "mobile-only" part on desktop (calendarslider's
	// collapsed chip showed on every viewport).
	w := testWidget{name: "w", kind: widget.Region}
	out := style.For(w).
		Part("p", style.Row(style.Space2)).
		OnlyOn(css.Mobile, "p", style.RevealedBy(widget.Open)).
		Stylesheet().String()

	base := out[:strings.Index(out, "@media")]
	if !strings.Contains(base, "display: none;") {
		t.Errorf("expected the base rule to hide the OnlyOn part, got:\n%s", base)
	}
}

func TestAnimatedReveal_ReducedMotionSilencesIt(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure}
	out := style.For(w).
		Part("p", style.Animate(style.MotionBase), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(out, "@media (prefers-reduced-motion: reduce)") {
		t.Fatalf("expected a reduced-motion block, got:\n%s", out)
	}
	if !strings.Contains(out, "transition: none;") {
		t.Errorf("expected reduced motion to silence the transition, got:\n%s", out)
	}
}

func TestAnimatedReveal_Deterministic(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure}
	build := func() string {
		return style.For(w).
			Part("p", style.Animate(style.MotionBase), style.RevealedBy(widget.Open)).
			OnlyOn(css.Mobile, "q", style.Animate(style.MotionSlow), style.RevealedBy(widget.Open)).
			Stylesheet().String()
	}
	if a, b := build(), build(); a != b {
		t.Errorf("two emissions differ:\n%s\n---\n%s", a, b)
	}
}
