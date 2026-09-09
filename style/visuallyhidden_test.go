//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

const partNativeInput = widget.Part("native-input")

// TestVisuallyHidden_KeepsTheElementFocusable: the whole point of the option is
// the half Hide() cannot do. A clipped 1px box stays in the tab order and in
// the accessibility tree; display:none does not, and emitting it here would
// silently make a skinned checkbox unreachable by keyboard.
func TestVisuallyHidden_KeepsTheElementFocusable(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Part(partNativeInput, style.VisuallyHidden()).
		Stylesheet().String()

	for _, want := range []string{
		"position: absolute;",
		"width: 1px;",
		"height: 1px;",
		"clip-path: inset(50%);",
		"overflow: hidden;",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("VisuallyHidden missing %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "display: none;") {
		t.Errorf("VisuallyHidden must never emit display:none — that is Hide()'s job and it drops the element from the tab order:\n%s", s)
	}
}

// TestVisuallyHidden_RejectsHide: the two options ask for opposite things, and
// display:none would win, costing the keyboard access VisuallyHidden was
// written to preserve.
func TestVisuallyHidden_RejectsHide(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	errs := style.For(wd).
		Part(partNativeInput, style.VisuallyHidden(), style.Hide()).
		Validate()

	const want = "Hide and VisuallyHidden contradict"
	for _, err := range errs {
		if strings.Contains(err.Error(), want) {
			return
		}
	}
	t.Errorf("expected an error containing %q, got %v", want, errs)
}

// TestVisuallyHidden_RejectsASecondPositionOwner: it sets position, so it joins
// the set Anchor/Docked/Flyout/… already guard — two of them on one rule means
// the last keyword silently wins.
func TestVisuallyHidden_RejectsASecondPositionOwner(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	errs := style.For(wd).
		Part(partNativeInput, style.VisuallyHidden(), style.Anchor()).
		Validate()

	const want = "all set position; use one"
	for _, err := range errs {
		if strings.Contains(err.Error(), want) {
			return
		}
	}
	t.Errorf("expected an error containing %q, got %v", want, errs)
}
