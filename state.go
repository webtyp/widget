package widget

// StateAttr is the DOM projection of a State: the attribute the markup writes and
// the stylesheet selects on, as one value.
//
// It is a distinct type on purpose. It used to be an fmt.KeyValue, which is a bare
// string pair — and a string pair fits every attribute method dom exposes, including
// the ones that write the wrong value. A state written with BindAttrBool lands as
// `data-x=""`, the sheet selects `data-x="true"`, and nothing reports the mismatch.
// The type is what makes that call not compile.
type StateAttr struct {
	key   string
	value string
}

// Key and Value are for the two libraries that have to reach the strings:
// webtyp/dom, which writes the attribute, and widget/style, which emits the
// selector. Reading them to hand-wire a state is the mistake this type exists to
// prevent — use dom's BindState/SetState instead.
func (a StateAttr) Key() string   { return a.key }
func (a StateAttr) Value() string { return a.value }

// Key and Value let a State itself be passed to dom's BindState/BindStateFunc/
// SetState: the projection is derived inside dom, so the value written is always
// the one the sheet selects on.
func (s State) Key() string   { return s.Attr().Key() }
func (s State) Value() string { return s.Attr().Value() }

// State is a state that the widget POSSESSES: written by Go, read by the stylesheet.
//
// Open is the expanded/visible state a SHEET selects on. Do not use BindState
// with it on a <details> element: the browser writes the native `open`
// attribute onto the element itself, and usermenu/targetlist mirror that into a
// signal precisely because the element owns the attribute. A data-open state
// written alongside is a separate thing that happens to share a name.
type State uint8

const (
	Selected State = iota
	Disabled
	Locked // read-only, but fully legible
	Invalid
	Busy
	Open    // deployed / expanded
	Current // active navigation item
)

func (s State) String() string {
	switch s {
	case Selected:
		return "Selected"
	case Disabled:
		return "Disabled"
	case Locked:
		return "Locked"
	case Invalid:
		return "Invalid"
	case Busy:
		return "Busy"
	case Open:
		return "Open"
	case Current:
		return "Current"
	default:
		return "Unknown"
	}
}

// Attr returns the attribute that the DOM writes and upon which the stylesheet selects.
// Markup and CSS match by construction, not by convention.
func (s State) Attr() StateAttr {
	switch s {
	case Selected:
		return StateAttr{key: "data-selected", value: "true"}
	case Disabled:
		return StateAttr{key: "data-disabled", value: "true"}
	case Locked:
		return StateAttr{key: "data-locked", value: "true"}
	case Invalid:
		return StateAttr{key: "data-invalid", value: "true"}
	case Busy:
		return StateAttr{key: "data-busy", value: "true"}
	case Open:
		return StateAttr{key: "data-open", value: "true"}
	case Current:
		return StateAttr{key: "data-current", value: "true"}
	default:
		return StateAttr{}
	}
}

// Cue is a state that the BROWSER possesses. It is only styled;
// it cannot be written from Go — which is why it is a separate type and has no Attr().
type Cue uint8

const (
	Hover Cue = iota
	Focus
	Press
	Target
	// FocusWithin is the browser's :focus-within — the element itself or any
	// descendant holds focus. Meant for widget/style's CueAcross, where it
	// reads as "while anything inside this region is focused": a floating
	// control that yields whenever the user is interacting with the content
	// it would otherwise cover.
	FocusWithin
)
