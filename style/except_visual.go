//go:build !wasm

package style

// Visual options based on scales.

// Pad applies internal padding according to the space scale.
func Pad(s Space) Option {
	return func(r *rule) {
		r.hasPad = true
		r.pad = s
	}
}

// PadEdge pads one block edge only. Pad() is all four sides, which is the right
// default; this exists for the case where a fixed overlay covers the top of a
// panel and the content underneath has to start below it without gaining the
// same inset left and right.
func PadEdge(e Edge, s Space) Option {
	return func(r *rule) {
		r.hasPadEdge = true
		r.padEdge = e
		r.padEdgeSpace = s
	}
}

// PadInline pads the inline axis (start and end) and nothing else. A chip whose
// height is contracted against another element — the fieldset legend matches
// the list badge — cannot take vertical padding, but its text still needs air
// at the sides; flush text against a filled chip edge reads as a bug, not as
// density.
func PadInline(s Space) Option {
	return func(r *rule) {
		r.hasPadInline = true
		r.padInline = s
	}
}

// Round applies border radius according to the radius scale.
func Round(rad Radius) Option {
	return func(r *rule) {
		r.hasRound = true
		r.round = rad
	}
}

// Raise applies shadow elevation according to the elevation scale.
func Raise(e Elevation) Option {
	return func(r *rule) {
		r.hasRaise = true
		r.raise = e
	}
}

// Capped bounds the element's block size against the viewport, so that a
// Scroll() region inside it finally has somewhere to overflow TO.
//
// This is the missing half of Scroll(). Scroll() emits height: 100%,
// min-height: 0, flex-grow: 1 and overflow-y: auto — every one of them
// RELATIVE, so all four are inert until some ancestor has a DEFINITE block
// size. A panel left at its content height has none: the percentage resolves
// against an indefinite height and falls back to auto, flex-grow finds no
// free space to claim, and the region grows to fit its content instead of
// scrolling. Nothing looks broken in the stylesheet; the list simply pushes
// the page down and the WHOLE app scrolls under the user's thumb.
//
// An in-flow panel usually inherits a bound from its layout. An out-of-flow
// one never does: Flyout() and Drawer() take the element out of the flow, so
// no ancestor's height reaches it. Pair Capped() with them, or with any
// panel that holds a list of unknown length.
func Capped(e Extent) Option {
	return func(r *rule) {
		r.hasCapped = true
		r.capped = e
	}
}

// Width applies the relative width (Size) to the rule.
func Width(s Size) Option {
	return func(r *rule) {
		r.hasSize = true
		r.size = s
	}
}

// IconBox sizes a part as a square that never shrinks — the shape an icon needs.
// A bare <svg> with no width or height falls back to the replaced-element default
// of 300x150 and blows the layout apart, so every part that renders one must
// declare its box here.
func IconBox(s IconSize) Option {
	return func(r *rule) {
		r.hasIcon = true
		r.icon = s
	}
}

// Meter sizes a part as a thin bar whose fill fraction is supplied per
// instance at runtime — an occupancy indicator, a progress bar. thickness
// sets the bar's cross-axis size from the space scale; its length axis reads
// the --meter-fill custom property, which the stylesheet declares the SHAPE
// of (that it feeds width, and a 0% fallback) but never assigns a value to —
// a stylesheet builder works on a zero-value receiver, so a per-instance
// fill level cannot be baked in. The host sets ONLY the value at runtime,
// e.g. a bare `--meter-fill:73%;` inline style — never a property name,
// selector, or unit beyond the percent sign the value itself carries.
func Meter(thickness Space) Option {
	return func(r *rule) {
		r.hasMeter = true
		r.meterThickness = thickness
	}
}

// CenterSelf centers a part horizontally within the space its container
// gives it, via margin-inline: auto — the counterpart of CenterContent(),
// which centers what a part CONTAINS rather than the part itself. An item
// with an explicit width (IconBox(), Width()) inside a wider flex or grid
// track does not stretch to fill it and defaults to the leading edge; pair
// it with CenterSelf() to center the fixed-size box inside that track — a
// calendar day marker inside its week-row column.
func CenterSelf() Option {
	return func(r *rule) {
		r.centerSelf = true
	}
}

// As sets the surface decision (background, text, border, and radius default).
func As(s Surface) Option {
	return func(r *rule) {
		r.hasSurface = true
		r.surface = s
	}
}

// GradientAngle repaints THIS surface's theme gradient at its own angle,
// reusing the app's colour stops (css.SetGradient publishes them as the
// family's ImageStopsVarName()). Every As(<family>) surface otherwise
// re-origins the one baked-in direction over its own box; a rail that wants
// the light end where it meets a panel's light end can flip just itself with
// GradientAngle("315deg"). Requires As() of a family surface (Primary,
// Accent, Panel, …); a no-op on a derived surface. Inert until the app calls
// css.SetGradient — the override references an unset custom property and is
// ignored, so the plain solid/gradient hook still applies.
func GradientAngle(angle string) Option {
	return func(r *rule) {
		r.hasGradientAngle = true
		r.gradientAngle = angle
	}
}

// FontSize sets the text size.
func FontSize(ts TextSize) Option {
	return func(r *rule) {
		r.hasTextSize = true
		r.textSize = ts
	}
}

// FontWeight sets the font weight.
func FontWeight(w Weight) Option {
	return func(r *rule) {
		r.hasWeight = true
		r.weight = w
	}
}

// Animate applies a transition according to the motion scale.
func Animate(m Motion) Option {
	return func(r *rule) {
		r.hasMotion = true
		r.motion = m
	}
}

// Rotate turns the element by a fixed quarter-turn step. Pair it with a state
// rule — When()/WhenWithin() — so the rotation IS the state, and with
// Animate() on the base rule so the turn is a transition instead of a jump.
//
// Not combinable with OnEdge() or Drawer(): both already own the element's
// `transform` and a second declaration would silently replace the first.
// Validate() rejects the combination rather than letting it fail on screen.
func Rotate(t Turn) Option {
	return func(r *rule) {
		r.hasRotate = true
		r.rotate = t
	}
}

// Divider draws a single hairline rule on one inline side — SideEnd for a
// leading region separated from the content that follows it, SideStart for
// the mirror case. It is independent of As(): a Surface's border is part of
// a background+text+border package deal, and a part that wants a plain
// separator with no background of its own (a badge that no longer fills a
// color, just marks where it ends) has no Surface to reach for.
func Divider(side Side) Option {
	return func(r *rule) {
		r.hasDivider = true
		r.dividerSide = side
	}
}

// DividerBetween draws a hairline BETWEEN a container's children — declared on
// the container, emitted on every child except the first, so N rows get N-1
// rules and no line dangles at either end.
//
// It exists because DividerBelow() cannot express this. A separator belongs to
// the PAIR of rows it comes between, not to one of them, and a border declared
// on the row is a property of the row: the last one in the list carries a rule
// under it with nothing after it to separate, which reads as a list that was
// cut off rather than one that ended. Pair it with Stack(SpaceNone) — a
// hairline only reads as a separator when the rows it divides are flush;
// leave a gap and the same line reads as an underline on the row above.
//
// Container-only, and only on a base Part()/Root() rule: the child combinator
// it needs cannot be expressed from the flat declaration list that On() and
// the state rules emit. Validate() rejects it there rather than letting it
// silently emit nothing.
func DividerBetween() Option {
	return func(r *rule) {
		r.hasDividerBetween = true
	}
}

// DividerBelow draws the same hairline rule as Divider, on the block-end
// (bottom) edge instead of an inline side. It is for a row in a stack that
// wants a visible seam under it — a navigation item in a drawer, a list row —
// without the row carrying a Surface of its own. Like Divider it is
// independent of As() and uses the outline colour.
func DividerBelow() Option {
	return func(r *rule) {
		r.hasDividerBelow = true
	}
}

// Wave waves a part back and forth with a gentle rotation anchored at bottom-center.
func Wave() Option {
	return func(r *rule) {
		r.hasWave = true
	}
}

// Pulse breathes/pulses scale anchored at center.
func Pulse() Option {
	return func(r *rule) {
		r.hasPulse = true
	}
}

// Pivot sets the transform origin for a part.
func Pivot(origin string) Option {
	return func(r *rule) {
		r.hasPivot = true
		r.pivot = origin
	}
}
