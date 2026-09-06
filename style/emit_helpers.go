//go:build !wasm

package style

import (
	"webtyp.com/css"
	"webtyp.com/widget"
)

func selectorOf(name widget.Name, part widget.Part) string {
	if part == "" {
		return "." + string(name)
	}
	return "." + string(name) + "__" + string(part)
}

func cuePseudo(c widget.Cue) string {
	switch c {
	case widget.Hover:
		return ":hover"
	case widget.Focus:
		return ":focus-visible"
	case widget.Press:
		return ":active"
	case widget.Target:
		return ":target"
	case widget.FocusWithin:
		return ":focus-within"
	default:
		return ""
	}
}

func displayFor(f flowType) string {
	switch f {
	case flowStack, flowRow, flowScrollRow, flowMediaBox, flowCover, flowMasterDetail:
		return "flex"
	case flowGrid, flowFillCentered:
		return "grid"
	case flowSplit, flowSidebar:
		// Both emit display:flex in @layer primitives. Saying "grid" here would
		// win from @layer states and strand every flex-basis/flex-grow the
		// primitive laid down.
		return "flex"
	default:
		return "block"
	}
}

// animatedReveal reports whether the rule pairs RevealedBy with a real
// motion: Animate(m) on the same rule upgrades the instant display swap into
// a choreographed fade — entry and exit alike. MotionNone, or no Animate at
// all, keeps the instant swap RevealedBy has always emitted. Drawer rules
// are excluded: Drawer owns its own slide choreography already.
func (r rule) animatedReveal() bool {
	return r.hasRevealed && !r.hasDrawer && r.hasMotion && r.motion != MotionNone
}

func layerVar(l widget.Layer) string {
	switch l {
	case widget.LayerBase:
		return css.ZBase.Var()
	case widget.LayerDropdown:
		return css.ZDropdown.Var()
	case widget.LayerSticky:
		return css.ZSticky.Var()
	case widget.LayerModal:
		return css.ZModal.Var()
	case widget.LayerToast:
		return css.ZToast.Var()
	case widget.LayerTooltip:
		return css.ZTooltip.Var()
	default:
		return css.ZBase.Var()
	}
}

// stackingFor is the package's ONLY source of a z-index value. An element
// taken out of the flow must never be left at `auto`: auto means "whatever
// the DOM order and the engine's compositor happen to do" — the silent
// failure of the OnEdge-under-the-input bug, where Safari composites a UA
// form control over an auto-z-index sibling that every other engine painted
// on top. Two declared levels exist:
//
//   - local (1): chrome that rides on its own widget's content — OnEdge, and
//     Parent-scoped Docked/Backdrop/EdgeStrip. Deliberately NOT the overlay
//     layer: overlay tokens start at --z-dropdown (100+), and a chip level
//     with a real dropdown would tie it and win on DOM order, rendering
//     under it. The same applies to a Parent backdrop: on the overlay layer
//     a dialog's click-catcher outranked its own panel, and the blur it
//     carried blurred the very thing it was meant to isolate.
//   - overlay (var(--z-dropdown) and up, via Kind.Layer()): real overlays —
//     Flyout, Drawer, Backdrop(Viewport), Docked(Viewport),
//     EdgeStrip(Viewport).
//
// Returns "" only when nothing in the rule is out of flow; the emitter then
// emits nothing, and Validate() flags an out-of-flow rule whose level came
// back unresolvable.
func stackingFor(r rule, layer widget.Layer) string {
	switch {
	case r.hasOnEdge:
		return "z-index: 1;"
	case r.hasDocked && r.dockedScope == Parent:
		return "z-index: 1;"
	case r.hasEdgeStrip && r.edgeStripScope == Parent:
		return "z-index: 1;"
	case r.hasBackdrop && r.backdropScope == Parent:
		return "z-index: 1;"
	case r.hasDrawer:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasFloatMiddle:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasDocked && r.dockedScope == Viewport:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasEdgeStrip && r.edgeStripScope == Viewport:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasFlyout:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasBackdrop && r.backdropScope == Viewport:
		return "z-index: " + layerVar(layer) + ";"
	default:
		return ""
	}
}

// The custom properties the floating-chrome seam is built on: a FloatingChrome
// host declares the strip it occupies along its edge, and every Scroll()
// region descendant reserves it through var(--floating-<edge>, 0px). They are
// DSL-owned variables crossing a component boundary, so they are not css
// tokens — the drift guard knows them by name.
const (
	floatingTopVar    = "--floating-top"
	floatingBottomVar = "--floating-bottom"
)

// floatingPadDecls are the reservations a Scroll() region carries: whatever a
// FloatingChrome ancestor declares it occupies along the region's edges —
// nothing declared means 0px, which is the default padding anyway, so the
// cost of the seam for authors who do not use it is zero.
func floatingPadDecls() []string {
	return []string{
		"padding-block-start: var(" + floatingTopVar + ", 0px);",
		"padding-block-end: var(" + floatingBottomVar + ", 0px);",
	}
}

// floatingPadDeclsWithGutter is floatingPadDecls with ScrollGutter's ambient
// gutter folded into the SAME calc as the FloatingChrome reservation, so a
// later layer can never plainly override one without the other — see
// ScrollGutter's doc for why that distinction matters. Kept as its own
// function (not a parameter on floatingPadDecls) so the no-gutter path — the
// overwhelming majority of Scroll() call sites — keeps emitting the exact
// byte-identical decls it always has.
func floatingPadDeclsWithGutter(g Space) []string {
	gutter := spaceVar(g)
	return []string{
		"padding-block-start: calc(var(" + floatingTopVar + ", 0px) + " + gutter + ");",
		"padding-block-end: calc(var(" + floatingBottomVar + ", 0px) + " + gutter + ");",
	}
}

func sidebarRailSel(sel string, side Side) string {
	if side == SideStart {
		return sel + " > :first-child"
	}
	return sel + " > :last-child"
}

func sidebarContentSel(sel string, side Side) string {
	if side == SideStart {
		return sel + " > :last-child"
	}
	return sel + " > :first-child"
}
