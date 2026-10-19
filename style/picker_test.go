//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// The picker's pseudo-elements do not exist in a browser without base-select,
// and one unknown pseudo-element drops a whole selector list: every rule that
// names one must sit inside the @supports block.
func TestPickerOnlyUnderSupports(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("sel", style.Picker()).Stylesheet().String()

	supports := strings.Index(s, "@supports (appearance: base-select)")
	picker := strings.Index(s, "::picker(select)")
	if supports < 0 || picker < 0 {
		t.Fatalf("expected a @supports block carrying ::picker(select), got:\n%s", s)
	}
	if picker < supports {
		t.Errorf("::picker(select) appears before the @supports guard, got:\n%s", s)
	}
}
