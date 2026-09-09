//go:build !wasm

package style

import (
	"webtyp.com/css"
	"webtyp.com/widget"
)

// placementDecls emits the declarations that take the element out of the
// flow and position it against its containing block — Backdrop, Veil,
// Drawer, Anchor, Docked, EdgeStrip, OnEdge, Flyout — together with the
// self-alignment flags that sit among them (Glyph, ChipBox, ControlBox, Button,
// CenterContent, StartContent, Meter, CenterSelf). The sequence is part of
// the byte-identical contract: the CSS engine breaks equal-specificity
// ties by source order, so these blocks run in the order they appear and
// nothing may be reordered.
func (r rule) placementDecls(layer widget.Layer) []string {
	var decls []string

	if r.hasBackdrop {
		if r.backdropScope == Viewport {
			decls = append(decls, "position: fixed;")
		} else {
			decls = append(decls, "position: absolute;")
		}
		decls = append(decls, "inset: 0;")
		// The stacking level is stackingFor's call — the single site that
		// decides local chrome from real overlay.
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasVeil {
		decls = append(decls, "background-color: "+css.FadeStatic(css.ColorSurface, 0.4)+";")
		decls = append(decls, "background-color: color-mix(in srgb, "+css.ColorSurface.NestedEnhanced()+" 60%, transparent);")
		// A wash alone still lets the page behind compete for attention.
		// Softening it is what makes the thing on top read as the only thing in
		// focus. -webkit- first: Safari has never unprefixed this.
		decls = append(decls, "-webkit-backdrop-filter: blur("+css.VeilBlur.Var()+");")
		decls = append(decls, "backdrop-filter: blur("+css.VeilBlur.Var()+");")
	}

	if r.hasDrawer {
		decls = append(decls, "position: fixed;")
		decls = append(decls, "inset-block: 0;")
		off := "translateX(100%)" // SideEnd parks off the inline-end edge
		if r.drawerSide == SideStart {
			decls = append(decls, "inset-inline-start: 0;")
			off = "translateX(-100%)"
		} else {
			decls = append(decls, "inset-inline-end: 0;")
		}
		decls = append(decls, "width: "+sizeValue(r.drawerSize)+";")
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
		// Parked off-screen but STILL in the layout: a transform (not
		// display) is what RevealedBy can transition back to translateX(0),
		// so open AND close are the same choreographed slide. visibility
		// drops it from the tab order; on the way out it is delayed to the
		// end of the slide (visibility 0s linear <dur>) — the SlideDeck
		// pattern, see slideDeckPageDecls.
		if r.hasRevealed {
			decls = append(decls, "transform: "+off+";", "visibility: hidden;")
			if r.drawerMotion != MotionNone {
				d := motionDurationVar(r.drawerMotion)
				decls = append(decls, "transition: transform "+d+" "+css.EaseInOut.Var()+", visibility 0s linear "+d+";")
			}
		}
	}

	if r.hasGlyph {
		decls = append(decls, "color: "+familyBase(r.glyph).LightValue()+";")
		decls = append(decls, "color: "+familyBase(r.glyph).EnhancedVar()+";")
		decls = append(decls, "fill: currentColor;")
	}

	if r.chipBox {
		decls = append(decls, "width: "+css.ChipWidth.Var()+";")
		decls = append(decls, "overflow: hidden;")
	}

	if r.capitalize {
		decls = append(decls, "text-transform: capitalize;")
	}

	if r.controlBox {
		decls = append(decls, "min-height: "+css.ControlHeight.Var()+";")
		decls = append(decls, "min-width: "+css.ControlWidth.Var()+";")
	}

	if r.visuallyHidden {
		// The standard visually-hidden recipe (Bootstrap's .visually-hidden,
		// Tailwind's sr-only): collapsed to a clipped 1px box rather than
		// display:none, which is what keeps it focusable and announced.
		decls = append(decls, "position: absolute;")
		decls = append(decls, "width: 1px;")
		decls = append(decls, "height: 1px;")
		decls = append(decls, "margin: -1px;")
		decls = append(decls, "padding: 0;")
		decls = append(decls, "border: 0;")
		decls = append(decls, "overflow: hidden;")
		decls = append(decls, "clip-path: inset(50%);")
		decls = append(decls, "white-space: nowrap;")
	}

	if r.buttonBox {
		decls = append(decls, "min-height: "+css.ControlHeight.Var()+";")
		decls = append(decls, "padding-inline: "+css.Space3.Var()+";")
		// The 799px fix: flex-shrink/grow govern the MAIN axis, and a button
		// inside a Stack is stretched across the CROSS one by the default
		// align-items: stretch. Any align-self but stretch takes content width.
		decls = append(decls, "align-self: center;")
		decls = append(decls, "flex-shrink: 0;")
		decls = append(decls, "flex-grow: 0;")
	}

	if r.logoBox {
		decls = append(decls, "height: "+css.ControlHeight.Var()+";")
		decls = append(decls, "width: auto;")
	}

	if r.centerContent {
		decls = append(decls, "display: flex;")
		decls = append(decls, "align-items: center;")
		decls = append(decls, "justify-content: center;")
	}

	if r.startContent {
		decls = append(decls, "display: flex;")
		decls = append(decls, "align-items: center;")
		decls = append(decls, "justify-content: flex-start;")
	}

	if r.hasAnchor {
		decls = append(decls, "position: relative;")
	}

	if r.foreground {
		// z-index only applies to a positioned element, so the position comes
		// with it. Level 1 is the local chrome level (stackingFor) — the same
		// one a Backdrop(Parent) sits at — so DOM order decides, and content
		// declared after the backdrop wins without outranking a real overlay.
		decls = append(decls, "position: relative;")
		decls = append(decls, "z-index: 1;")
	}

	if r.hasDocked {
		if r.dockedScope == Viewport {
			decls = append(decls, "position: fixed;")
		} else {
			decls = append(decls, "position: absolute;")
		}
		decls = append(decls, "margin: 0;")
		// All four insets, the unpinned pair explicitly auto. A corner pin owns
		// the whole box: leaving the opposite edges unset lets whatever set them
		// before survive, and an over-constrained absolute box collapses instead
		// of moving. That is what happened when this overrode a Flyout on one
		// device — top and bottom were both live and the panel became 8x2.
		if r.dockedEdge == EdgeTop {
			decls = append(decls, "inset-block-start: "+spaceVar(r.dockedGap)+";", "inset-block-end: auto;")
		} else {
			decls = append(decls, "inset-block-end: "+spaceVar(r.dockedGap)+";", "inset-block-start: auto;")
		}
		if r.dockedSide == SideStart {
			decls = append(decls, "inset-inline-start: "+spaceVar(r.dockedGap)+";", "inset-inline-end: auto;")
		} else {
			decls = append(decls, "inset-inline-end: "+spaceVar(r.dockedGap)+";", "inset-inline-start: auto;")
		}
		// Parent vs Viewport docking levels are stackingFor's call: a Parent
		// dock is local chrome, a Viewport dock claims the widget's layer.
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasFloatMiddle {
		decls = append(decls, "position: fixed;")
		// Half the viewport (SizeHalf, not a literal), then back by half the
		// element's own height — the same translateY(-50%) centring OnEdge
		// uses for its straddle. The unpinned inline side is explicitly
		// auto, Docked's discipline: an over-constrained fixed box
		// collapses instead of moving.
		decls = append(decls, "top: "+sizeValue(Half)+";")
		decls = append(decls, "transform: translateY(-50%);")
		decls = append(decls, "inset-block-end: auto;")
		if r.floatMiddleSide == SideStart {
			decls = append(decls, "inset-inline-start: "+spaceVar(r.floatMiddleGap)+";", "inset-inline-end: auto;")
		} else {
			decls = append(decls, "inset-inline-end: "+spaceVar(r.floatMiddleGap)+";", "inset-inline-start: auto;")
		}
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasEdgeStrip {
		if r.edgeStripScope == Viewport {
			decls = append(decls, "position: fixed;")
		} else {
			decls = append(decls, "position: absolute;")
		}
		decls = append(decls, "inset-block: 0;")
		if r.edgeStripSide == SideStart {
			decls = append(decls, "inset-inline-start: 0;")
		} else {
			decls = append(decls, "inset-inline-end: 0;")
		}
		// Deliberately no width: — that is what differentiates this from
		// Drawer(); the element sizes to its own content/padding.
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasMeter {
		decls = append(decls, "height: "+spaceVar(r.meterThickness)+";")
		decls = append(decls, "width: var(--meter-fill, 0%);")
	}

	if r.centerSelf {
		decls = append(decls, "margin-inline: auto;")
	}

	if r.hasOnEdge {
		decls = append(decls, "position: absolute;")
		decls = append(decls, "margin: 0;")
		// min-height keeps every chip the same pill size; the flex row centres
		// its text. Geometry below does the alignment.
		decls = append(decls, "min-height: "+css.ChipHeight.Var()+";")
		if !r.centerContent && !r.startContent {
			decls = append(decls, "display: flex;")
			decls = append(decls, "align-items: center;")
		}
		if r.onEdgeEdge == EdgeTop {
			// Put the chip's CENTRE on the centre of the 1px border the content
			// below draws. ChipSeat(EdgeTop) reserves 0.5·--chip-height for that
			// content's border-box top; +0.5px reaches the middle of the stroke.
			// translateY(-50%), not a negative margin: with ChipSeat the whole
			// chip sits inside the container's box, so nothing pokes out for an
			// ancestor's scrollHeight to miss, and it can centre by its OWN
			// rendered height — landing exactly on the line whatever that height
			// turns out to be, instead of only when it equals the token. The
			// z-index below already opened the stacking context a transform
			// would; there is no transition here for reduced-motion to keep.
			seat := "calc(0.5 * " + css.ChipHeight.Var() + " + 0.5px)"
			if r.onEdgeBlock != SpaceNone {
				seat = "calc(0.5 * " + css.ChipHeight.Var() + " + 0.5px + " + spaceVar(r.onEdgeBlock) + ")"
			}
			decls = append(decls, "inset-block-start: "+seat+";")
			decls = append(decls, "transform: translateY(-50%);")
		} else {
			decls = append(decls, "inset-block-end: "+spaceVar(r.onEdgeBlock)+";")
			decls = append(decls, "margin-block-end: calc(-0.5 * "+css.ChipHeight.Var()+");")
		}
		if r.onEdgeSide == SideStart {
			decls = append(decls, "inset-inline-start: "+spaceVar(r.onEdgeInline)+";")
		} else {
			decls = append(decls, "inset-inline-end: "+spaceVar(r.onEdgeInline)+";")
		}
		// Local level, never the overlay layer — stackingFor is the single
		// source of that decision.
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasFlyout {
		decls = append(decls, "position: absolute;")
		decls = append(decls, "inset-block-start: 100%;", "inset-block-end: auto;")
		if r.flyoutSide == SideStart {
			decls = append(decls, "inset-inline-start: 0;", "inset-inline-end: auto;")
		} else {
			decls = append(decls, "inset-inline-end: 0;", "inset-inline-start: auto;")
		}
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	// display: none last, after every declaration that could turn it back into
	// display: flex. CenterContent/StartContent open a block by adding
	// display: flex, so a part with both RevealedBy and one of them must end
	// hidden or the reveal contract silently dies. Same reason r.hidden sits
	// last of all.
	return decls
}
