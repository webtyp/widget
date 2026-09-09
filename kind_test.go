package widget

import "testing"

// A Form holds Invalid, Open AND Selected: a form with a conditional section
// and a set of chips is ordinary, and it must not surrender its validation
// state to get either.
func TestFormAllowsInvalidOpenAndSelected(t *testing.T) {
	for _, s := range []State{Invalid, Open, Selected} {
		if !Form.Allows(s) {
			t.Errorf("Form.Allows(%s) = false, want true", s.String())
		}
	}
}

// Widening Form must not widen everything: Current stays out. It means "the
// one you are on" in a navigation, which a form has no notion of.
func TestFormStillRejectsCurrent(t *testing.T) {
	if Form.Allows(Current) {
		t.Error("Form.Allows(Current) = true, want false")
	}
}

// The kinds this change does not touch keep their exact answer for Open.
func TestOpenAllowanceUnchangedForOtherKinds(t *testing.T) {
	allowed := []Kind{Menu, Dialog, Disclosure, Alert, Combobox}
	for _, k := range allowed {
		if !k.Allows(Open) {
			t.Errorf("%s.Allows(Open) = false, want true", k.String())
		}
	}
	for _, k := range []Kind{Listbox, Tabs, Grid} {
		if k.Allows(Open) {
			t.Errorf("%s.Allows(Open) = true, want false", k.String())
		}
	}
}
