//go:build !wasm

package style

import (
	"sort"

	"webtyp.com/css"
	"webtyp.com/fmt"
)

// flowSelfDecls returns the declarations a flow puts on the CONTAINER itself,
// without the child rules that go with it. The main path groups flows into
// shared selectors collected from the root and the parts; a rule that is
// neither — a CueWithin — has nothing to group with and needs them inline, or
// it silently emits no `display` at all.
func flowSelfDecls(r rule) []string {
	switch r.flowType {
	case flowStack:
		return []string{"display: flex;", "flex-direction: column;", "gap: var(--gap);", "min-height: 0;"}
	case flowRow:
		return []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);", "align-items: center;"}
	case flowSplit:
		return []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);"}
	case flowGrid:
		return []string{"display: grid;", "gap: var(--gap);", "grid-template-columns: repeat(auto-fit, minmax(min(var(--track), 100%), 1fr));"}
	case flowFixedGrid:
		return []string{"display: grid;", "gap: var(--gap);", "grid-template-columns: repeat(var(--cols), minmax(0, 1fr));"}
	case flowCenter:
		return []string{"margin-inline: auto;", "max-width: var(--max-width);", "width: 100%;"}
	case flowFillCentered:
		return []string{"display: grid;", "place-items: center;", "min-height: 100%;", "width: 100%;"}
	case flowScrollRow:
		return []string{"display: flex;", "gap: var(--gap);", "overflow-x: auto;", "scroll-snap-type: x mandatory;", "scroll-behavior: smooth;"}
	case flowMediaBox:
		return []string{"aspect-ratio: var(--ratio);", "overflow: hidden;", "display: flex;", "justify-content: center;", "align-items: center;"}
	case flowCover:
		return []string{"height: 100dvh;", "display: flex;", "flex-direction: column;"}
	case flowSidebar:
		return []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);"}
	case flowSlideDeck:
		return slideDeckStripDecls()
	case flowMasterDetail:
		return masterDetailStripDecls()
	default:
		return nil
	}
}

// slideDeckStripDecls: el contenedor es el bloque contenedor de las capas y
// recorta lo que queda aparcado afuera.
func slideDeckStripDecls() []string {
	return []string{
		"position: relative;",
		"overflow: hidden;",
	}
}

// slideDeckPageDecls: cada hijo cubre el contenedor y espera aparcado en el borde
// inline-start. visibility (no display) es lo que lo saca del orden de tabulación
// sin impedir la transición: se apaga DESPUÉS de que el deslizamiento termina, por
// eso el retardo en la segunda propiedad.
func slideDeckPageDecls(m Motion) []string {
	d := motionDurationVar(m)
	return []string{
		"position: absolute;",
		"inset: 0;",
		"transform: translateX(-100%);",
		"visibility: hidden;",
		"transition: transform " + d + " " + css.EaseInOut.Var() + ", visibility 0s linear " + d + ";",
	}
}

// slideDeckCurrentDecls: el panel activo entra hasta su sitio y es visible desde
// el primer fotograma, para que se le vea llegar.
func slideDeckCurrentDecls(m Motion) []string {
	d := motionDurationVar(m)
	return []string{
		"transform: translateX(0);",
		"visibility: visible;",
		"transition: transform " + d + " " + css.EaseInOut.Var() + ", visibility 0s;",
	}
}

// drawerRevealDecls is the "arrived" state a RevealedBy applies to a Drawer:
// the same shape as slideDeckCurrentDecls, so a drawer's entry and exit share
// one transition instead of the exit being a discrete display flip. MotionNone
// drops the transition (snap into place, still no display toggle).
func drawerRevealDecls(m Motion) []string {
	d := []string{"transform: translateX(0);", "visibility: visible;"}
	if m != MotionNone {
		d = append(d, "transition: transform "+motionDurationVar(m)+" "+css.EaseInOut.Var()+", visibility 0s;")
	}
	return d
}

// revealTransitionDecls is the transition an animated reveal (RevealedBy
// paired with Animate on the same rule) runs on its hidden base rule.
// display flips discretely — at the start on entry, at the end on exit —
// while opacity fades over the motion duration; durations and easing are
// catalog tokens, the same ones Drawer and Animate resolve to. `all` keeps
// every other transition the rule already had (an Interactive hover tint,
// ...); it is listed first so the discrete display timing wins for display.
// Opacity is the only animated property on purpose: a slide would need a
// length the closed scales do not have, and AGENTS.md §2 forbids inventing
// one — a fade softens the swap with nothing invented.
func revealTransitionDecls(m Motion) []string {
	d := motionDurationVar(m)
	return []string{"transition: all " + d + " " + css.EaseInOut.Var() + ", display " + d + " allow-discrete;"}
}

// startingStyleBlock renders the entry from-state for animated reveals:
// @starting-style sets opacity 0 on each shown selector, so the first frames
// of an entry fade in from nothing instead of flashing the final state.
// Browsers without @starting-style ignore the block and the reveal is the
// same instant swap RevealedBy always emitted — the graceful fallback, not a
// breakage. Selectors are sorted for deterministic output.
func startingStyleBlock(sels []string) string {
	if len(sels) == 0 {
		return ""
	}
	sort.Strings(sels)
	sb := fmt.GetConv()
	defer sb.PutConv()
	sb.WriteString("@starting-style {\n")
	for _, sel := range sels {
		sb.WriteString(formatRule([]string{sel}, []string{"opacity: 0;"}))
	}
	sb.WriteString("}\n")
	return sb.GetString(fmt.BuffOut)
}

// autoRotateStepSeconds is how long one layer holds the screen before the
// next takes over. Not a token: webtyp/css's duration scale (150-400ms) is
// for UI transitions, and an unattended background rotation runs one to two
// orders of magnitude slower than that — there is nothing in the catalog to
// reuse here, so this is a fixed implementation constant of AutoRotate, the
// same way slideDeckPageDecls' "0s" fallback above is.
const autoRotateStepSeconds = 5

// autoRotateCycleSeconds is the time for every layer to get exactly one
// turn. AutoRotateLayers is fixed (see its doc comment), so this is fixed
// too: every AutoRotate() rule in a page shares one @keyframes definition
// and one cycle length, regardless of how many of the layers a given
// instance's markup actually fills.
const autoRotateCycleSeconds = autoRotateStepSeconds * AutoRotateLayers

// autoRotateKeyframesName is the single @keyframes identifier every
// AutoRotate() rule references. Keyframe names are global in CSS, not
// scoped to a selector, and the crossfade shape below never varies, so one
// definition — duplicated harmlessly if more than one AutoRotate part
// appears in the same stylesheet — is simpler than inventing a per-instance
// name.
const autoRotateKeyframesName = "tw-auto-rotate"

// autoRotateStripDecls: the container is the containing block for its
// layered children and clips whatever a layer's transform pushes outside it.
func autoRotateStripDecls() []string {
	return []string{
		"position: relative;",
		"overflow: hidden;",
	}
}

// autoRotateLayerDecls: every layer covers the container and starts hidden.
// The animation is what makes each one visible in turn; a layer whose delay
// has not elapsed yet, or whose slot has no element mounted in it, simply
// keeps showing this resting opacity — which is why :first-child overrides
// it below, so the reduced-motion fallback (animation: none, see the
// prefers-reduced-motion block this rule feeds into) is never a blank layer.
func autoRotateLayerDecls() []string {
	return []string{
		"position: absolute;",
		"inset: 0;",
		// A rotating layer is almost always a photograph, and inset alone does
		// not size a replaced element: an <img> with auto width/height keeps
		// its intrinsic dimensions no matter what the four insets say, so a
		// 1024x1368 portrait rendered in a 982x517 hero showed its top-left
		// corner and nothing else. Sizing the box explicitly and cropping with
		// object-fit is what "the photo fills the banner" means; on a
		// non-replaced child (a <div> with a background) these are inert.
		"width: 100%;",
		"height: 100%;",
		"object-fit: cover;",
		"opacity: 0;",
		fmt.Sprintf("animation: %s %ds infinite;", autoRotateKeyframesName, autoRotateCycleSeconds),
	}
}

// autoRotateFirstDecls: the resting (non-animated) opacity for the first
// layer. It is redundant while the animation runs — animation always wins
// over a plain opacity declaration on the same property — and it is the
// entire reduced-motion fallback once the animation is switched off.
func autoRotateFirstDecls() []string {
	return []string{"opacity: 1;"}
}

// autoRotateDelayDecls staggers slot's turn by slot * autoRotateStepSeconds.
// slot is the layer's 1-based DOM position (1 is :first-child, which needs
// no delay — its turn is the animation's own 0%).
func autoRotateDelayDecls(slot int) []string {
	return []string{fmt.Sprintf("animation-delay: %ds;", (slot-1)*autoRotateStepSeconds)}
}

// autoRotateKeyframesCSS is the shared crossfade shape every layer plays out
// on its own staggered clock: fully visible for the first chunk of its slot,
// a short fade, then hidden for the rest of the cycle until its next turn.
// The 12%/16% breakpoints are fixed because autoRotateCycleSeconds is fixed
// — see AutoRotateLayers' doc comment for why this package cannot compute
// them from a real image count.
func autoRotateKeyframesCSS() string {
	return "@keyframes " + autoRotateKeyframesName + " {\n" +
		"  0%, 12% { opacity: 1; }\n" +
		"  16%, 100% { opacity: 0; }\n" +
		"}\n\n"
}

// masterDetailStripDecls lays the two panels out as a horizontal scroll-snap
// strip. direction: rtl is load-bearing — it puts the start edge on the right,
// which is where scroll position 0 already rests, so the master panel is what
// shows on arrival with no scroll nudge at mount time.
func masterDetailStripDecls() []string {
	return []string{
		"display: flex;",
		"flex-direction: row;",
		"flex-wrap: nowrap;",
		"direction: rtl;",
		"gap: 0;",
		"overflow-x: auto;",
		"overflow-y: hidden;",
		"scroll-snap-type: x mandatory;",
		"scroll-behavior: smooth;",
	}
}

// masterDetailResetDecls clears whatever the wide-screen flow left on the
// children. Split in particular gives every child a flex-basis of
// calc((40rem - 100%) * 999), which below 40rem is a six-figure pixel value: any
// child the two panel rules do not cover — a modal mount point, a portal anchor
// — keeps it and blows the strip's scroll width apart.
func masterDetailResetDecls() []string {
	return []string{
		"flex: 0 0 auto;",
		// direction: ltr on EVERY child, not just the two panels. The strip is
		// laid out rtl so the master lands at scroll position 0; a third child —
		// a modal mount point, a portal anchor — would otherwise inherit it and
		// render its text backwards.
		"direction: ltr;",
	}
}

// masterDetailDetailDecls sizes the detail panel. It is first in the DOM, the
// order a desktop Split wants, and order: 2 moves it beside the master without
// touching the markup. Its width is a share of the SCROLL CONTAINER, not of the
// viewport: the host panel is not guaranteed to be viewport-wide, and a vw here
// overflows the strip by the difference.
func masterDetailDetailDecls(detail Size) []string {
	return []string{
		"direction: ltr;",
		"flex: 0 0 " + sizeValue(detail) + ";",
		"scroll-snap-align: end;",
		"order: 2;",
	}
}

func masterDetailMasterDecls() []string {
	return []string{
		"direction: ltr;",
		"flex: 0 0 100%;",
		"scroll-snap-align: start;",
		"order: 1;",
	}
}
