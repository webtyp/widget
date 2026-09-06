//go:build !wasm

package style

import (
	"sort"

	"webtyp.com/css"
	"webtyp.com/fmt"
)

// emitHover renders the fine-pointer half of `@layer states`:
//
//	@media (hover: hover) { @layer states { … } }
//
// It carries the Hover cues that emitStates split out (hoverCues, which already
// include the Interactive()-derived ones) plus every CueWithinHover rule. A
// touch tap fires :hover and leaves it stuck, so anything scoped here — a hover
// tint, a hover-revealed control — simply never applies on a phone, instead of
// applying and never clearing.
func (s *Sheet) emitHover(sb *fmt.Conv, hoverCues []cueEmission) {
	// CueWithinHover rules are the same descendant selectors as CueWithin,
	// gated on the fine-pointer capability. A touch tap fires :hover but never
	// (hover: hover), so reveals scoped here cannot misfire on a phone.
	type cueWithinHoverEntry struct {
		key   cueWithinKey
		decls []string
	}
	var sortedCueWithinHover []cueWithinHoverEntry
	for k, r := range s.cueWithinHover {
		var d []string
		if r.hasFlow {
			d = append(d, flowSelfDecls(r)...)
		}
		d = append(d, r.Decls(s.widget.WidgetKind().Layer())...)
		d = append(d, primitiveDecls(r)...)
		if len(d) == 0 {
			continue
		}
		sortedCueWithinHover = append(sortedCueWithinHover, cueWithinHoverEntry{key: k, decls: d})
	}
	sort.Slice(sortedCueWithinHover, func(i, j int) bool {
		a, b := sortedCueWithinHover[i].key, sortedCueWithinHover[j].key
		if a.cue != b.cue {
			return a.cue < b.cue
		}
		if a.container != b.container {
			return a.container < b.container
		}
		return a.part < b.part
	})

	if len(hoverCues) > 0 || len(sortedCueWithinHover) > 0 {
		hoverSB := fmt.GetConv()
		for _, sc := range hoverCues {
			sel := selectorOf(s.widget.WidgetName(), sc.key.part) + cuePseudo(sc.key.cue)
			hoverSB.WriteString(formatRule([]string{sel}, sc.decls))
		}
		for _, sc := range sortedCueWithinHover {
			sel := selectorOf(s.widget.WidgetName(), sc.key.container) + cuePseudo(sc.key.cue) +
				" " + selectorOf(s.widget.WidgetName(), sc.key.part)
			hoverSB.WriteString(formatRule([]string{sel}, sc.decls))
		}
		sb.WriteString("@media " + css.FinePointer.Query() + " {\n")
		sb.WriteString("@layer states {\n")
		sb.WriteString(hoverSB.GetString(fmt.BuffOut))
		sb.WriteString("}\n")
		sb.WriteString("}\n\n")
		hoverSB.PutConv()
	}
}
