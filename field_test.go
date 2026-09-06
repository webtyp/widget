package widget

import (
	"testing"

	"webtyp.com/fmt"
)

func TestFieldAnatomy(t *testing.T) {
	// Every expectation derives from the NameField constant — the class
	// contract is name + "__" + part, and a rename must not silently break
	// the test or let it drift from the constant it is guarding.
	wantRoot := string(NameField)
	if got, want := NameField.Root().String(), wantRoot; got != want {
		t.Errorf("NameField.Root().String() = %q; want %q", got, want)
	}

	for _, tc := range []struct {
		part Part
		want string
	}{
		{PartLabel, wantRoot + "__label"},
		{PartInput, wantRoot + "__input"},
		{PartError, wantRoot + "__error"},
		{PartRadioGroup, wantRoot + "__radio-group"},
	} {
		if got := NameField.Class(tc.part).String(); got != tc.want {
			t.Errorf("NameField.Class(%s).String() = %q; want %q", tc.part, got, tc.want)
		}
	}

	expectedAttr := fmt.KeyValue{Key: "class", Value: string(NameField.Class(PartLabel))}
	if got := NameField.Class(PartLabel).AsAttr(); got != expectedAttr {
		t.Errorf("NameField.Class(PartLabel).AsAttr() = %+v; want %+v", got, expectedAttr)
	}
}