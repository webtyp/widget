package widget

// fieldClassName is the class-prefix literal of the shared form field — the
// single place the string is written. Every consumer, tests included, must
// reach the value through the exported constant, never by typing it again.
const fieldClassName = "vy-field"

// NameField is the shared anatomy of the form field.
//
// It is emitted by webtyp.com/form (which builds the markup) and
// styled by webtyp.com/components/fieldset (which is the global skin
// of the forms). It lives here because it crosses the boundary between these
// two libraries, and neither can own it without the other depending on it.
const NameField = Name(fieldClassName)

// Parts of the field. The names are generic by design: Part is local to its
// widget and only becomes a class through a Name, so "label" here produces
// fieldClassName + "__label" and never collides with the "label" of another
// widget.
const (
	PartLabel      = Part("label")
	PartInput      = Part("input")
	PartError      = Part("error")
	PartRadioGroup = Part("radio-group")
	PartSubmit     = Part("submit")
	// PartReveal is the show/hide toggle button a masked input (input.Masked)
	// draws inside itself. form emits it only when the field is masked;
	// fieldset styles it (the eye glyph, positioned inside the control) and
	// switches it via the existing Selected state (revealed = selected).
	PartReveal = Part("reveal")
	// PartForm is the <form> element that wraps the field stack. form emits it;
	// the fieldset skin styles it as the one place the inter-field rhythm
	// lives (a gap on the container, so the ends do not double the way
	// per-field margins would).
	PartForm = Part("form")
)
