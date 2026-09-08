package widget

import "webtyp.com/fmt"

// Kind is the type of widget according to WAI-ARIA Authoring Practices.
// It determines the role, valid states, and expected keyboard interactions.
// Closed by design: if a widget does not fit any pattern, it is almost always two widgets.
type Kind uint8

const (
	Region Kind = iota
	Listbox
	Menu
	Dialog
	Disclosure
	Tabs
	Toolbar
	Grid
	Combobox
	Form
	Alert
)

func (k Kind) String() string {
	switch k {
	case Region:
		return "Region"
	case Listbox:
		return "Listbox"
	case Menu:
		return "Menu"
	case Dialog:
		return "Dialog"
	case Disclosure:
		return "Disclosure"
	case Tabs:
		return "Tabs"
	case Toolbar:
		return "Toolbar"
	case Grid:
		return "Grid"
	case Combobox:
		return "Combobox"
	case Form:
		return "Form"
	case Alert:
		return "Alert"
	default:
		return "Unknown"
	}
}

// Layer is the stacking level of a pattern. It is an enum, not a css.Token:
// this package travels inside the WASM binary and must not import css.
// widget/style maps it to the --z-* catalog.
type Layer uint8

const (
	LayerBase Layer = iota
	LayerDropdown
	LayerSticky
	LayerModal
	LayerToast
	LayerTooltip
)

// Role returns the ARIA role of the Kind as a KeyValue attribute.
func (k Kind) Role() fmt.KeyValue {
	var val string
	switch k {
	case Region:
		val = "region"
	case Listbox:
		val = "listbox"
	case Menu:
		val = "menu"
	case Dialog:
		val = "dialog"
	case Disclosure:
		val = "group"
	case Tabs:
		val = "tablist"
	case Toolbar:
		val = "toolbar"
	case Grid:
		val = "grid"
	case Combobox:
		val = "combobox"
	case Form:
		val = "form"
	case Alert:
		val = "alert"
	}
	return fmt.KeyValue{Key: "role", Value: val}
}

// Layer returns the stacking level of the Kind.
func (k Kind) Layer() Layer {
	switch k {
	case Menu, Disclosure, Combobox:
		return LayerDropdown
	case Dialog:
		return LayerModal
	case Alert:
		return LayerToast
	default:
		return LayerBase
	}
}

// Allows returns whether the given state is meaningful for the Kind.
func (k Kind) Allows(s State) bool {
	// Universal set, allowed for every Kind
	if s == Disabled || s == Locked || s == Busy {
		return true
	}

	switch k {
	case Listbox, Tabs, Grid:
		return s == Selected || s == Current
	case Menu:
		return s == Open || s == Current
	case Dialog, Disclosure, Alert:
		return s == Open
	case Combobox:
		return s == Open || s == Selected || s == Invalid
	case Form:
		// A form with a conditional section: an "other, please specify" box, a
		// panel revealed by a choice. Open is the same reveal a Disclosure
		// carries; a Form that also holds Invalid must not have to give it up
		// to get it. Busy is already covered by the universal set.
		return s == Invalid || s == Open
	default:
		return false
	}
}
