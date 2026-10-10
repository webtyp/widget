//go:build !wasm

package style

import (
	"webtyp.com/css"
	"webtyp.com/widget"
)

// AppBar is the one recipe for a screen's top bar: a row with the shared
// control height plus a Space1 band above and below, welded edge to edge,
// that never shrinks. It stays IN the flow, so the content starts below it.
//
// While st is absent the bar retracts: it slides off the top edge by pulling
// its own block size out of the flow (negative margin-block-start), so the
// content below rises with it instead of jumping — a display swap
// (RevealedBy) cannot animate the band it frees. m is the slide motion;
// MotionNone retracts instantly, and prefers-reduced-motion silences it.
//
// The ancestor must clip (HideOverflow) for the retracted bar to be gone.
func AppBar(st widget.State, m Motion) Option {
	return func(r *rule) {
		r.hasAppBar = true
		r.appBarState = st
		r.appBarMotion = m
	}
}

func appBarBlockSizeExpr() string {
	return "calc(" + css.ControlHeight.Var() + " + 2 * " + spaceVar(Space1) + ")"
}

func appBarBaseDecls(m Motion) []string {
	decls := []string{
		"display: flex;",
		"align-items: center;",
		"flex-shrink: 0;",
		"block-size: " + appBarBlockSizeExpr() + ";",
		"padding-inline: " + spaceVar(Space3) + ";",
		"border-radius: 0;",
		"margin: 0;",
		"margin-block-start: calc(-1 * " + appBarBlockSizeExpr() + ");",
	}
	if m != MotionNone {
		d := motionDurationVar(m)
		decls = append(decls, "transition: margin-block-start "+d+" "+css.EaseInOut.Var()+";")
	}
	return decls
}

func appBarRevealDecls() []string {
	return []string{
		"margin-block-start: 0;",
	}
}
