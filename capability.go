package widget

// Widget is the identity. It is the only mandatory interface.
type Widget interface {
	WidgetName() Name
	WidgetKind() Kind
}

// Capabilities — each hook asserts only the capability it needs.
type Selectable interface{ Select(id string) }
type Dismissible interface{ Dismiss() }
type Expandable interface{ Expand(open bool) }

// Filterable is implemented by a control that narrows a host's collection: a
// search bar, a date picker, a category select. The host registers one sink and
// the control calls it whenever the term changes.
//
// The term is a plain string because that is what the consuming end takes
// (view.Presenter.Filter(term string)). A calendar formats its range into a
// term; a select emits the chosen id. Widening this to a structured query is a
// change to that contract, not to this one.
//
// The host registers the sink once, at wiring time, and never reads it back —
// which is why this is a setter and not a field: a control is free to fan the
// sink out to several inputs, or to debounce it, without the host knowing.
type Filterable interface{ OnFilterChange(func(term string)) }

// TypingAware is implemented by a Filterable that only reports a term when the
// person commits it (Enter): the host registers one sink that fires on every
// keystroke, with no term, so it can tell the person that Enter is what searches.
type TypingAware interface{ OnFilterTyping(func()) }
