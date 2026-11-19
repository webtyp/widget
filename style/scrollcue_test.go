//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestScrollCueMasksTheRegionAndShipsItsAnimationOnce(t *testing.T) {
	w := testWidget{name: "rail", kind: widget.Region}

	s := style.For(w).
		Part("a", style.Scroll(), style.ScrollCue()).
		Part("b", style.Scroll(), style.ScrollCue()).
		Stylesheet().String()

	for _, want := range []string{"mask-image: linear-gradient(", "animation-timeline: scroll(self);", "@keyframes tw-scroll-cue", "@property --scroll-cue-start", "@property --scroll-cue-end"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in:\n%s", want, s)
		}
	}
	if n := strings.Count(s, "@keyframes tw-scroll-cue"); n != 1 {
		t.Errorf("the keyframes are shared, want 1 definition, got %d", n)
	}

	plain := style.For(w).Part("a", style.Scroll()).Stylesheet().String()
	if strings.Contains(plain, "scroll-cue") || strings.Contains(plain, "mask-image") {
		t.Errorf("a plain Scroll() must stay untouched, got:\n%s", plain)
	}
}

func TestScrollCueWithoutScrollFailsTheBuild(t *testing.T) {
	w := testWidget{name: "rail", kind: widget.Region}
	if errs := style.For(w).Part("a", style.ScrollCue()).Validate(); len(errs) == 0 {
		t.Error("ScrollCue() on a part that does not scroll must be rejected")
	}
}
