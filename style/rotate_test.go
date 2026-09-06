//go:build !wasm

package style_test

import (
	"webtyp.com/widget"
	"webtyp.com/widget/style"
	"strings"
	"testing"
)

func TestRotateTurnsOnState(t *testing.T) {
	// A chevron points down at rest and flips when its host opens. The
	// rotation IS the state: a When() rule, not a class the component toggles
	// by hand, and Animate() on the base rule makes the flip a transition.
	w := testWidget{name: "ss", kind: widget.Combobox}
	s := style.For(w).
		Root(style.Anchor()).
		Part("icon", style.Rotate(style.TurnNone), style.Animate(style.MotionBase)).
		WhenWithin(widget.Open, "", "icon", style.Rotate(style.TurnHalf)).
		Stylesheet().String()

	if !strings.Contains(s, "transform: rotate(0deg);") {
		t.Errorf("expected the resting rotation, got:\n%s", s)
	}
	if !strings.Contains(s, "transform: rotate(180deg);") {
		t.Errorf("expected the open-state rotation, got:\n%s", s)
	}
	if !strings.Contains(s, `.ss[data-open="true"] .ss__icon`) {
		t.Errorf("expected the state to reach the icon from the root, got:\n%s", s)
	}
	if !strings.Contains(s, "transition:") {
		t.Errorf("expected Animate to make the turn a transition, got:\n%s", s)
	}
}

func TestRotateAloneIsNotDroppedAsEmpty(t *testing.T) {
	// emitsNothing() decides whether a rule is worth a selector. A part whose
	// only declaration is a rotation is a real rule; forgetting it there is a
	// silent failure — nothing renders and nothing complains.
	w := testWidget{name: "ss", kind: widget.Combobox}
	s := style.For(w).Part("icon", style.Rotate(style.TurnQuarter)).Stylesheet().String()
	if !strings.Contains(s, ".ss__icon") || !strings.Contains(s, "transform: rotate(90deg);") {
		t.Errorf("a Rotate-only rule must still be emitted, got:\n%s", s)
	}
}

func TestRotateRejectsTransformConflicts(t *testing.T) {
	// OnEdge and Drawer already own `transform`. Two declarations on one rule
	// means the later silently wins — a defect that is invisible in the CSS
	// and only shows on screen. Validate() is where it must die.
	w := testWidget{name: "ss", kind: widget.Combobox}
	sheet := style.For(w).
		Root(style.Anchor()).
		Part("chip", style.OnEdge(style.EdgeTop, style.SideStart, style.SpaceNone, style.Space2), style.Rotate(style.TurnHalf))
	if errs := sheet.Validate(); len(errs) == 0 {
		t.Fatal("expected Rotate + OnEdge on one rule to be rejected")
	}
}
