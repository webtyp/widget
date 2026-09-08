package widget

import "testing"

// A Form holds Invalid AND Open: a form with a conditional section is
// ordinary, and it must not surrender its validation state to get the reveal.
func TestFormAllowsInvalidAndOpen(t *testing.T) {
	for _, s := range []State{Invalid, Open} {
		if !Form.Allows(s) {
			t.Errorf("Form.Allows(%s) = false, want true", s.String())
		}
	}
}

// Widening Form must not widen anything else: Selected and Current stay out.
func TestFormStillRejectsSelectionStates(t *testing.T) {
	for _, s := range []State{Selected, Current} {
		if Form.Allows(s) {
			t.Errorf("Form.Allows(%s) = true, want false", s.String())
		}
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
