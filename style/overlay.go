//go:build !wasm

package style

import "webtyp.com/widget"

// Scope says what an overlay dimensions against.
type Scope uint8

const (
	// Parent covers the nearest positioned ancestor (position: absolute).
	Parent Scope = iota
	// Viewport covers the entire window (position: fixed).
	Viewport
)

// Backdrop removes the element from the normal flow and stretches it over its Scope.
func Backdrop(s Scope) Option {
	return func(r *rule) {
		r.hasBackdrop = true
		r.backdropScope = s
	}
}

// Veil fills the element with a translucent wash overlaying the surface.
// Only makes sense alongside Backdrop.
func Veil() Option {
	return func(r *rule) {
		r.hasVeil = true
	}
}

// Foreground lifts the element above its own widget's Parent-scoped chrome
// without taking it out of the flow. It is the counterpart of Backdrop(Parent):
// a backdrop is pinned at the local stacking level (z-index 1, see
// stackingFor), and an in-flow sibling sits at `auto`, which loses — so a hero
// banner's caption vanished behind the photograph the moment the photograph
// actually filled its box. Positioning the content at the SAME level and
// letting DOM order decide is what puts it back on top; giving it a higher
// number would outrank a real overlay declared later.
//
// Reach for it only where the widget owns both layers. Between widgets,
// stacking is the layout's business, not a part's.
func Foreground() Option {
	return func(r *rule) {
		r.foreground = true
	}
}

// RevealedBy binds hiding/showing the element to a widget State.
func RevealedBy(st widget.State) Option {
	return func(r *rule) {
		r.hasRevealed = true
		r.revealedBy = st
	}
}

// Anchor makes the element a positioning reference: it emits position:
// relative, which is what lets a Flyout descendant hang from it. It is the
// trigger's container — a menu, a combobox.
//
// Anchor() only wins if nothing positioned sits between it and the Flyout:
// CSS resolves the Flyout's inset against the nearest POSITIONED ancestor,
// so a Docked/OnEdge/Backdrop part in between becomes the containing block
// instead and the Anchor is dead code. Validate() rejects that composition —
// see Within() for the legal way to declare nesting.
func Anchor() Option {
	return func(r *rule) {
		r.hasAnchor = true
	}
}

// Flyout lifts the element out of the flow and hangs it under its containing
// block, flush with the given inline edge, at the widget kind's stacking
// layer. Use it for a dropdown: left in the flow, an expanding menu pushes
// everything below it down and the list jumps under the pointer that opened
// it.
//
// The containing block is the nearest positioned ancestor. An Anchor() on the
// chain is what makes it hang where intended — but any positioned part
// between the two becomes the containing block instead, and nothing in the
// emitted CSS distinguishes the two. Validate() walks the declared part tree
// and rejects the interposition; declare the nesting with Within().
func Flyout(side Side) Option {
	return func(r *rule) {
		r.hasFlyout = true
		r.flyoutSide = side
	}
}

// Docked pins the element inside a corner, above the content and out of the
// flow, at the widget kind's stacking layer. Use it for a control that must not
// cost the content a band of its own: a floating action button, a row's
// overflow menu.
//
// Parent pins it to the corner of the nearest positioned ancestor — the
// Anchor only when nothing positioned sits between the two. Viewport pins it
// to the screen, so it stays put while the content behind it scrolls or
// swipes — and it disappears with the widget, because a fixed descendant of a
// display:none ancestor is not rendered either.
//
// Docked also makes the element a containing block: a Flyout inside it hangs
// from THIS box, not from whatever Anchor sits above. Validate() reports the
// theft; the fix is either a docked trigger that spans the anchor, or the
// Flyout moving out of the docked part.
func Docked(scope Scope, edge Edge, side Side, gap Space) Option {
	return func(r *rule) {
		r.hasDocked = true
		r.dockedScope = scope
		r.dockedEdge = edge
		r.dockedSide = side
		r.dockedGap = gap
	}
}

// EdgeStrip pins the element to one full inline edge of its Scope — the
// entire block-axis span (inset-block: 0) plus one inline edge — and, unlike
// Drawer(), sizes itself to its own content/padding instead of forcing a
// width, and unlike Docked(), spans the full edge instead of a single
// corner. Use it for chrome that must stay reachable along an entire edge
// and is always visible, never toggled: a calendar's full-height prev/next
// overlay strip. Drawer() is the right primitive instead when the panel is
// meant to slide in and out — it requires a paired RevealedBy() for exactly
// that reason; EdgeStrip() has no such requirement because it is not a
// panel, it is permanent chrome.
//
// Pin it inside an Anchor()ed ancestor for Parent scope, the same
// relationship Docked() needs.
func EdgeStrip(scope Scope, side Side) Option {
	return func(r *rule) {
		r.hasEdgeStrip = true
		r.edgeStripScope = scope
		r.edgeStripSide = side
	}
}

// OnEdge centres the element ON one of its Anchor's edge lines — half outside
// the box, half inside — the way a fieldset legend rides the border it labels.
//
// inline is how far the chip is indented along that line. block is the extra
// distance from the Anchor's border to the line being ridden, ON TOP of the
// half chip-height the straddle already needs — SpaceNone for the common case.
//
// For EdgeTop the straddle is now driven entirely by --chip-height: the chip's
// top edge sits flush against the Anchor's padding box and its centre lands
// half a chip-height in, so the Anchor pairs this with ChipSeat(EdgeTop) to
// reserve exactly that space and the chip never pokes out of the box it labels.
// EdgeBottom keeps the older "block is the gap from the border" behaviour that
// the list badges depend on.
func OnEdge(edge Edge, side Side, block Space, inline Space) Option {
	return func(r *rule) {
		r.hasOnEdge = true
		r.onEdgeEdge = edge
		r.onEdgeSide = side
		r.onEdgeBlock = block
		r.onEdgeInline = inline
	}
}

// ChipSeat reserves half a chip-height of padding on one block edge — the space
// an OnEdge chip on a child straddles into. Pair it with OnEdge(EdgeTop) on the
// chip: the chip seats flush against this container's padding box and its
// centre lands exactly on the content's top line, fully inside the container.
func ChipSeat(e Edge) Option {
	return func(r *rule) {
		r.hasChipSeat = true
		r.chipSeatEdge = e
	}
}

// FloatingChrome declares that this element occupies a strip along its edge,
// and every Scroll() region DESCENDANT of it must reserve that strip — the
// contract between a floating action button and the scroll container behind
// it. It emits, on its own box:
//
//	--floating-bottom: calc(<size> + 2 * <gap>);
//
// (and the --floating-top counterpart for EdgeTop). The custom property is
// inherited, so the reservation crosses widget and repository boundaries:
// the host says "I occupy this band of my edge" and a Scroll() descendant —
// whatever widget it belongs to — pads itself by var(--floating-bottom, 0px)
// without either knowing the other's class name. No FloatingChrome means no
// declaration, and every scroller reserves nothing (the 0px default).
//
// size is IconSize, not Size: floating chrome pinned to a screen edge is by
// construction a small icon-only control (a FAB, a hamburger) — the same
// IconBox(...) an author already gave its glyph. Size's members are
// percentages/keywords meant for a panel's share of a Split or a Flyout's
// width; a percentage inside padding-block-end resolves against the SCROLL
// REGION'S OWN inline size, not a fixed footprint, and max-content is not a
// <length> at all — neither compiles into a calc() that means what this
// needs. gap is doubled because the same value both pads the control on the
// edge closest to its icon and sets its own inset off the container edge.
//
// This is the seam the badge-over-FAB overlap needed: the FAB's box is
// invisible to the badge's scrollHeight, so no padding computed from the
// badge's own box could reserve the space where the FAB really paints.
func FloatingChrome(edge Edge, size IconSize, gap Space) Option {
	return func(r *rule) {
		r.hasFloatingChrome = true
		r.floatingChromeEdge = edge
		r.floatingChromeSize = size
		r.floatingChromeGap = gap
	}
}

// Drawer anchors the element to one inline edge of the viewport, full height,
// at the widget kind's stacking layer. It is the slide-in panel of a mobile
// navigation; pair it with RevealedBy(widget.Open) to control visibility and
// with a sibling Backdrop(Viewport)+Veil() for the dimmed page behind it.
//
// m is the slide motion, exactly like SlideDeck(m): the panel rests parked
// off the inline edge (transform, not display) and RevealedBy transitions it
// in AND back out, so the exit is choreographed instead of a hard cut. Pass
// MotionNone to park it with no animation.
//
// Drawer sets the element's width. Do NOT also pass Width() — Validate rejects it.
func Drawer(side Side, size Size, m Motion) Option {
	return func(r *rule) {
		r.hasDrawer = true
		r.drawerSide = side
		r.drawerSize = size
		r.drawerMotion = m
	}
}

// FloatMiddle pins the element to the vertical middle of one viewport edge —
// position: fixed, top at half the viewport, pulled back by half its own
// height — at the widget kind's stacking layer. It is the always-reachable
// floating control: a menu button riding the middle of the inline-end edge
// sits under the thumb wherever the page has scrolled to, which no corner
// pin can promise once the content is taller than the screen.
//
// Corner pins already exist (Docked) and stay the only answer for corners:
// extending Docked with a middle would tangle two semantics — a corner owns
// all four insets, a middle owns one edge plus its own height — in one
// option. Viewport scope only, by the same reasoning: middle-centering
// inside a scrolling parent fights the scroll instead of floating over it.
//
// Owns transform (translateY), like Drawer and OnEdge: do NOT combine with
// Rotate() — Validate rejects it. gap is the Space step off the edge.
func FloatMiddle(side Side, gap Space) Option {
	return func(r *rule) {
		r.hasFloatMiddle = true
		r.floatMiddleSide = side
		r.floatMiddleGap = gap
	}
}
