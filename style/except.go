//go:build !wasm

package style

// Fill takes up the entire available height.
func Fill() Option {
	return func(r *rule) {
		r.fill = true
	}
}

// Scroll overflows internally instead of growing. Implies Fill().
func Scroll() Option {
	return func(r *rule) {
		r.scroll = true
		r.fill = true
	}
}

// ScrollGutter adds an ambient gutter to a Scroll() region's block edges,
// ADDITIVE with whatever a FloatingChrome ancestor already reserves there —
// never replacing it. Without ScrollGutter, Scroll() reserves exactly the
// FloatingChrome strip and nothing else (0px when there is none), as before;
// with it, padding-block-start/end becomes
// calc(var(--floating-<edge>, 0px) + <gutter>).
//
// This exists because a plain PadEdge()/Pad() on a Scroll() region's own part
// would be emitted in the widgets layer, which OUTRANKS the primitives layer
// where the FloatingChrome reservation lives — CSS layers do not add, a later
// layer's declaration replaces the earlier one outright. That silently erases
// the strip a floating button reserved, hiding the scroller's last row under
// it. ScrollGutter avoids that by folding the ambient gutter into the SAME
// calc the reservation already uses, so both are always the final value —
// there is no plain override to lose to.
func ScrollGutter(s Space) Option {
	return func(r *rule) {
		r.hasScrollGutter = true
		r.scrollGutter = s
	}
}

// Grow takes the free space along the inline axis and nothing else. It is the
// Row counterpart of Fill(): Fill() also claims `height: 100%`, which inside a
// Row resolves against the row and stretches the part into a full-height block.
// Use Grow() for the item in a Row that should push its siblings to the
// trailing edge.
func Grow() Option {
	return func(r *rule) {
		r.grow = true
	}
}

// PushEnd sends the part to the trailing edge of its line. It is the companion
// of Grow(): Grow() absorbs the free space so the items after it are pushed
// out, PushEnd() moves the free space in front of a single item — the only way
// to keep something right-aligned once flex-wrap has dropped it onto a line of
// its own.
func PushEnd() Option {
	return func(r *rule) {
		r.pushEnd = true
	}
}

// CenterContent centers whatever the element contains, on both axes. A button
// holding nothing but an icon needs it: the icon is a replaced element with
// display: block, so the text-align a button carries by default does not move
// it and it sits against the leading edge.
func CenterContent() Option {
	return func(r *rule) {
		r.centerContent = true
	}
}

// Glyph colours what the element draws — its text and, through currentColor,
// its icons — with a surface's base colour, and leaves the background alone. It
// is the "tinted, not filled" treatment: a nav item that is merely available
// shows a coloured icon, the selected one gets the filled surface via As().
func Glyph(s Surface) Option {
	return func(r *rule) {
		r.hasGlyph = true
		r.glyph = s
	}
}

// ChipBox gives the element the shared chip width, the box a legend or a badge
// occupies. Fixing it is what makes a column of chips line up instead of each
// one hugging its own text; the text itself is truncated by the component that
// renders it.
func ChipBox() Option {
	return func(r *rule) {
		r.chipBox = true
	}
}

// Capitalize uppercases the first letter of every word the element renders.
// It is for text that arrives from a data source in whatever case the source
// happens to store it — a model's field names becoming a form's labels — so
// the presentation layer decides the casing instead of every caller having to
// pre-format the string it passes in.
func Capitalize() Option {
	return func(r *rule) {
		r.capitalize = true
	}
}

// Hide removes the element. Its use is inside On(): a part that exists on wide
// screens and not on a phone keeps its base styling and is switched off for the
// one device, which OnlyOn cannot express — OnlyOn hides by default and reveals
// per device, the opposite direction.
func Hide() Option {
	return func(r *rule) {
		r.hidden = true
	}
}

// VisuallyHidden removes the element from sight while leaving it in the
// accessibility tree and in the tab order — the opposite trade to Hide(), which
// removes it from both.
//
// Its use is the native control a component skins: a checkbox or radio whose
// own appearance cannot be styled, paired with a <label for> carrying the
// visible pill. The input stays the real control — it keeps focus, keyboard
// toggling and the screen reader's announcement — and the label is what the eye
// sees. Hide() cannot express that: display:none takes the input out of the tab
// order, so the control becomes unreachable by keyboard.
//
// To hide something nobody needs to reach, that is Hide().
func VisuallyHidden() Option {
	return func(r *rule) {
		r.visuallyHidden = true
	}
}

// Show puts back an element that the base rule hid. It is Hide()'s missing
// counterpart, and it is for a STATE rule: a control that swaps one glyph for
// another — a hamburger becoming a close cross — hides the inactive one by
// default and shows it through When()/WhenWithin(). RevealedBy() cannot
// express that, because it selects on the element that CARRIES the state, and
// here the state lives on the button while what changes are its children.
//
// The display it restores follows the rule's own flow, exactly as RevealedBy
// does, so revealing a Row comes back as a row and not as a block.
func Show() Option {
	return func(r *rule) {
		r.shown = true
	}
}

// StartContent packs what the element contains against its leading edge. It is
// the counterpart of CenterContent, for the case where the same part is centred
// in one state and aligned in another — an icon alone in a narrow rail, icon
// and label once the rail expands.
func StartContent() Option {
	return func(r *rule) {
		r.startContent = true
	}
}

// ControlBox gives the element the shared control height, the rhythm every
// interactive row in the app is measured against — a list row, a form field.
// Pinning both to one token is what keeps them from drifting apart.
func ControlBox() Option {
	return func(r *rule) {
		r.controlBox = true
	}
}

// Button is the one recipe for a button: the shared control height, inline
// padding for its label, and a box that neither stretches to its container nor
// shrinks under pressure. s paints it — Primary, Secondary, Danger, Subtle —
// with the hover, focus and press treatments Interactive derives, and with that
// surface's default radius.
//
// It exists because the parts are not safely composable by hand. KeepSize()
// looks like the way to stop a button filling its panel and is not: it emits
// flex-shrink/flex-grow, which govern the MAIN axis, while a button inside a
// Stack is stretched across the CROSS axis by align-items: stretch. Every
// component that composed its own button reached a different answer, and one of
// them shipped an 800px-wide "Add row" bar.
//
// Use it for anything the user presses: a button, a summary that acts as one.
// Not for an interactive surface that is not a button — a clickable table row
// or a calendar day keeps Interactive(), which paints without claiming a
// control's box.
func Button(s Surface) Option {
	return func(r *rule) {
		r.hasSurface, r.surface, r.interactive = true, s, true
		r.buttonBox = true
	}
}

// LogoBox caps a media element to the shared control height, width auto to
// preserve whatever aspect ratio the source art has. A brand mark's file —
// often an SVG traced from artwork with no relationship to a nav row's
// geometry — has an intrinsic size the browser will render at verbatim if
// nothing constrains it, the same failure IconBox exists to prevent for a
// bare <svg>. IconBox is not the fix here: it forces width equal to height,
// and a logo is rarely square. ControlBox is not either: min-height sets a
// floor, not the ceiling an oversized intrinsic image needs.
func LogoBox() Option {
	return func(r *rule) {
		r.logoBox = true
	}
}

// KeepSize does NOT reflow: maintains its size under any width.
func KeepSize() Option {
	return func(r *rule) {
		r.keepSize = true
	}
}

// EdgeToEdge has no border radius or margin: flush against parent container.
func EdgeToEdge() Option {
	return func(r *rule) {
		r.edgeToEdge = true
	}
}

// HideOverflow clips descendants (overflow: hidden).
func HideOverflow() Option {
	return func(r *rule) {
		r.hideOverflow = true
	}
}
