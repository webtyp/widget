//go:build !wasm

package style_test

import (
	"fmt"
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// TestStateAttrMatchesEmittedSelector is the round-trip that would have caught
// the targetlist defect: the value State.Attr() publishes must equal the value
// the emitted selector matches. These are two independent switch statements
// that happen to agree; nothing asserted it. Iterate the states, do not
// spot-check. A state written as data-x="" while the sheet selects
// data-x="true" is exactly the silent no-op this plan removes.
func TestStateAttrMatchesEmittedSelector(t *testing.T) {
	kindFor := map[widget.State]widget.Kind{
		widget.Selected: widget.Listbox,
		widget.Disabled: widget.Region,
		widget.Locked:   widget.Region,
		widget.Invalid:  widget.Form,
		widget.Busy:     widget.Region,
		widget.Open:     widget.Disclosure,
		widget.Current:  widget.Tabs,
	}

	for st, kind := range kindFor {
		w := testWidget{name: "w", kind: kind}
		s := style.For(w).
			Part("item", style.As(style.Panel)).
			When(st, "item", style.As(style.Highlight)).
			Stylesheet().String()

		attr := st.Attr()
		want := fmt.Sprintf(".w__item[%s=\"%s\"]", attr.Key(), attr.Value())
		if !strings.Contains(s, want) {
			t.Errorf("state %s: selector %q absent from emitted CSS — Attr() and the emitter disagree:\n%s", st, want, s)
		}
	}
}
