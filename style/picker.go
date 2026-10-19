//go:build !wasm

package style

import (
	"webtyp.com/css"
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// Picker styles the native option list of a <select> — the popup the browser
// draws when the control opens — with the accent: the option under the pointer
// or the keyboard is outlined, the chosen one is filled. Without it the list
// highlights in the operating system's blue, a colour nothing else in the
// application uses, and CSS cannot reach it.
//
// Put it on a part that is only ever a <select>: it declares
// `appearance: base-select`, which is what makes the list stylable at all and
// which would strip the native look of any other control it landed on.
//
// It is a progressive enhancement: the rules sit in `@supports (appearance:
// base-select)` (Chrome and Edge 135+). Elsewhere nothing is emitted that
// applies, and the browser keeps its native list.
func Picker() Option {
	return func(r *rule) {
		r.hasPicker = true
	}
}

// emitPickers writes the Picker rules, unlayered and inside @supports: the
// pseudo-elements it needs (::picker(select), ::checkmark) do not exist in a
// browser without base-select, and one unknown pseudo-element would drop a
// whole selector list.
func (s *Sheet) emitPickers(sb *fmt.Conv, parts []widget.Part) {
	for _, p := range parts {
		if !s.partRules[p].hasPicker {
			continue
		}
		sel := selectorOf(s.widget.WidgetName(), p)
		accent := css.ColorAccent.Var()
		sb.WriteString("@supports (appearance: base-select) {\n")
		sb.WriteString(formatRule([]string{sel, sel + "::picker(select)"}, []string{"appearance: base-select;"}))
		// base-select turns the closed control into a flex button whose content
		// sits at the top: centre it the way the native select centred its text.
		sb.WriteString(formatRule([]string{sel}, []string{"display: flex;", "align-items: center;"}))
		sb.WriteString(formatRule([]string{sel + "::picker(select)"}, []string{
			"background-color: " + css.ColorBackground.Var() + ";",
			"color: " + css.ColorOnSurface.Var() + ";",
			"border: 1px solid " + css.ColorOutline.Var() + ";",
			"border-radius: " + radiusVar(RadiusMd) + ";",
			"padding: " + spaceVar(Space1) + ";",
		}))
		sb.WriteString(formatRule([]string{sel + " option"}, []string{
			"padding: " + spaceVar(Space2) + " " + spaceVar(Space3) + ";",
			"border-radius: " + radiusVar(RadiusSm) + ";",
		}))
		// Outline for a passing state, fill for the confirmed one.
		sb.WriteString(formatRule([]string{sel + " option:hover", sel + " option:focus-visible"}, []string{
			"background-color: transparent;",
			"box-shadow: inset 0 0 0 1px " + accent + ";",
		}))
		sb.WriteString(formatRule([]string{sel + " option:checked"}, []string{
			"background-color: " + accent + ";",
			"color: " + css.ColorOnAccent.Var() + ";",
		}))
		sb.WriteString(formatRule([]string{sel + " option::checkmark"}, []string{"display: none;"}))
		sb.WriteString("}\n\n")
	}
}
