package widget

import "testing"

func TestStateAttrTable(t *testing.T) {
	cases := []struct {
		state State
		key   string
		value string
	}{
		{Selected, "data-selected", "true"},
		{Disabled, "data-disabled", "true"},
		{Locked, "data-locked", "true"},
		{Invalid, "data-invalid", "true"},
		{Busy, "data-busy", "true"},
		{Open, "data-open", "true"},
		{Current, "data-current", "true"},
		{SpanFull, "data-span-full", "true"},
	}
	for _, c := range cases {
		attr := c.state.Attr()
		if attr.Key() != c.key {
			t.Errorf("%s: Attr().Key() = %q, want %q", c.state, attr.Key(), c.key)
		}
		if attr.Value() != c.value {
			t.Errorf("%s: Attr().Value() = %q, want %q", c.state, attr.Value(), c.value)
		}
	}
}

// A StateAttr carries the value the stylesheet selects on. There is no public
// constructor: the only source is State.Attr(). Forging one from a hand-typed
// string is the mistake the type exists to prevent.
//
//	// must not compile — a StateAttr is not an fmt.KeyValue anymore
//	el.Attr(st.Attr().Key(), st.Attr().Value()) // two reads, one state: also wrong
func TestStateAttrNotConstructible(t *testing.T) {
	var _ StateAttr = Selected.Attr()
}
