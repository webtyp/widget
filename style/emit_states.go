//go:build !wasm

package style

import (
	"sort"

	"webtyp.com/css"
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// emitStates renders `@layer states` — every rule driven by a written state or
// a browser cue: When / WhenWithin, the RevealedBy reveals, Cue / CueWithin,
// and the Interactive() hover/focus/press derivations. Hover cues are built
// here (so the Interactive-derived ones are included) but returned rather than
// written: emit_hover.go emits them inside the fine-pointer media query.
//
// After the layer closes it emits CueAcross and StateAcross UNLAYERED. Their
// whole job is to override a RevealedBy reveal, which the device path also
// emits unlayered, and an @layer rule — however specific — always loses to an
// unlayered one. Their :has()-from-root selector is already more specific than
// a bare `.n__part[data-x]`, so among unlayered rules they win on merit.
func (s *Sheet) emitStates(sb *fmt.Conv, parts []widget.Part) (hoverCues []cueEmission) {
	statesSB := fmt.GetConv()
	defer statesSB.PutConv()

	stateDecls := make(map[stateKey][]string)
	for k, sr := range s.stateRules {
		stateDecls[k] = sr.Decls(s.widget.WidgetKind().Layer())
	}

	// animatedKeys marks the shown selectors whose reveal fades: their
	// @starting-style entry is emitted after the layer closes.
	animatedKeys := make(map[stateKey]bool)
	if s.rootRule.hasRevealed {
		sk := stateKey{state: s.rootRule.revealedBy, part: ""}
		if s.rootRule.hasDrawer {
			stateDecls[sk] = append(stateDecls[sk], drawerRevealDecls(s.rootRule.drawerMotion)...)
		} else {
			stateDecls[sk] = append(stateDecls[sk], "display: "+displayFor(s.rootRule.flowType)+";")
			if s.rootRule.animatedReveal() {
				stateDecls[sk] = append(stateDecls[sk], "opacity: 1;")
				animatedKeys[sk] = true
			}
		}
	}
	for _, p := range parts {
		pr := s.partRules[p]
		if pr.hasRevealed {
			sk := stateKey{state: pr.revealedBy, part: p}
			if pr.hasDrawer {
				stateDecls[sk] = append(stateDecls[sk], drawerRevealDecls(pr.drawerMotion)...)
			} else {
				stateDecls[sk] = append(stateDecls[sk], "display: "+displayFor(pr.flowType)+";")
				if pr.animatedReveal() {
					stateDecls[sk] = append(stateDecls[sk], "opacity: 1;")
					animatedKeys[sk] = true
				}
			}
		}
	}

	type sortedState struct {
		key   stateKey
		decls []string
	}
	var sortedStates []sortedState
	var startSels []string
	for k, decls := range stateDecls {
		if len(decls) > 0 {
			sortedStates = append(sortedStates, sortedState{key: k, decls: decls})
		}
	}
	sort.Slice(sortedStates, func(i, j int) bool {
		if sortedStates[i].key.state != sortedStates[j].key.state {
			return sortedStates[i].key.state < sortedStates[j].key.state
		}
		return sortedStates[i].key.part < sortedStates[j].key.part
	})

	for _, ss := range sortedStates {
		attr := ss.key.state.Attr()
		sel := fmt.Sprintf("%s[%s=\"%s\"]", selectorOf(s.widget.WidgetName(), ss.key.part), attr.Key(), attr.Value())
		statesSB.WriteString(formatRule([]string{sel}, ss.decls))
		if animatedKeys[ss.key] {
			startSels = append(startSels, sel)
		}
	}

	// WhenWithin: an ancestor part's written state reaching a part inside it —
	// the State twin of the cueWithin block further down.
	type stateWithinEntry struct {
		key   stateWithinKey
		decls []string
	}
	var sortedStateWithin []stateWithinEntry
	for k, r := range s.stateWithin {
		var d []string
		if r.hasFlow {
			d = append(d, flowSelfDecls(r)...)
		}
		d = append(d, r.Decls(s.widget.WidgetKind().Layer())...)
		d = append(d, primitiveDecls(r)...)
		if len(d) == 0 {
			continue
		}
		sortedStateWithin = append(sortedStateWithin, stateWithinEntry{key: k, decls: d})
	}
	sort.Slice(sortedStateWithin, func(i, j int) bool {
		a, b := sortedStateWithin[i].key, sortedStateWithin[j].key
		if a.state != b.state {
			return a.state < b.state
		}
		if a.container != b.container {
			return a.container < b.container
		}
		return a.part < b.part
	})
	for _, sw := range sortedStateWithin {
		attr := sw.key.state.Attr()
		sel := fmt.Sprintf("%s[%s=\"%s\"] %s",
			selectorOf(s.widget.WidgetName(), sw.key.container), attr.Key(), attr.Value(),
			selectorOf(s.widget.WidgetName(), sw.key.part))
		statesSB.WriteString(formatRule([]string{sel}, sw.decls))
	}

	cueDecls := make(map[cueKey][]string)
	for k, cr := range s.cueRules {
		cueDecls[k] = cr.Decls(s.widget.WidgetKind().Layer())
	}

	// An explicit Cue() call for the same (cue, part) already populated
	// cueDecls above, from s.cueRules — that hand-written rule is a deliberate
	// override (e.g. a hover tint that must read as the widget's own accent
	// instead of the generic Interactive() mix) and wins over the derived
	// one instead of being appended alongside it, which would just add a
	// second, later background-color declaration that silently wins by
	// source order regardless of which one the author meant to show.
	addInteractive := func(p widget.Part, r rule) {
		if r.interactive {
			base := familyBase(r.surface)
			if base.Name != "" {
				kHover := cueKey{cue: widget.Hover, part: p}
				if len(cueDecls[kHover]) == 0 {
					cueDecls[kHover] = append(cueDecls[kHover],
						"background-color: "+css.HoverStatic(base)+";",
						"background-color: "+css.Hover(base)+";",
					)
				}

				kFocus := cueKey{cue: widget.Focus, part: p}
				if len(cueDecls[kFocus]) == 0 {
					cueDecls[kFocus] = append(cueDecls[kFocus],
						"background-color: "+css.FocusStatic(base)+";",
						"background-color: "+css.Focus(base)+";",
					)
				}

				kPress := cueKey{cue: widget.Press, part: p}
				if len(cueDecls[kPress]) == 0 {
					cueDecls[kPress] = append(cueDecls[kPress],
						"background-color: "+css.PressStatic(base)+";",
						"background-color: "+css.Press(base)+";",
					)
				}
			}
		}
	}

	addInteractive("", s.rootRule)
	for _, p := range parts {
		addInteractive(p, s.partRules[p])
	}

	// Hover is split out and returned for emit_hover.go to emit inside the
	// same fine-pointer media query as cueWithinHover: a touch tap fires
	// :hover and leaves it stuck, which would otherwise beat a State rule of
	// equal specificity — e.g. a selected row painted grey by its own stuck
	// hover instead of its selected color, on a device that can never
	// actually hover. Focus and Press stay here: :focus-visible never fires
	// from touch and :active clears on release, so neither has that failure
	// mode.
	var sortedCues []cueEmission
	for k, decls := range cueDecls {
		if len(decls) == 0 {
			continue
		}
		if k.cue == widget.Hover {
			hoverCues = append(hoverCues, cueEmission{key: k, decls: decls})
			continue
		}
		sortedCues = append(sortedCues, cueEmission{key: k, decls: decls})
	}
	sort.Slice(sortedCues, func(i, j int) bool {
		if sortedCues[i].key.cue != sortedCues[j].key.cue {
			return sortedCues[i].key.cue < sortedCues[j].key.cue
		}
		return sortedCues[i].key.part < sortedCues[j].key.part
	})
	sort.Slice(hoverCues, func(i, j int) bool {
		return hoverCues[i].key.part < hoverCues[j].key.part
	})

	for _, sc := range sortedCues {
		sel := selectorOf(s.widget.WidgetName(), sc.key.part) + cuePseudo(sc.key.cue)
		statesSB.WriteString(formatRule([]string{sel}, sc.decls))
	}

	// The one descendant selector in the package: an ancestor's cue reaching a
	// part inside it.
	type cueWithinEntry struct {
		key   cueWithinKey
		decls []string
	}
	var sortedCueWithin []cueWithinEntry
	for k, r := range s.cueWithin {
		var d []string
		if r.hasFlow {
			d = append(d, flowSelfDecls(r)...)
		}
		d = append(d, r.Decls(s.widget.WidgetKind().Layer())...)
		d = append(d, primitiveDecls(r)...)
		if len(d) == 0 {
			continue
		}
		sortedCueWithin = append(sortedCueWithin, cueWithinEntry{key: k, decls: d})
	}
	sort.Slice(sortedCueWithin, func(i, j int) bool {
		a, b := sortedCueWithin[i].key, sortedCueWithin[j].key
		if a.cue != b.cue {
			return a.cue < b.cue
		}
		if a.container != b.container {
			return a.container < b.container
		}
		return a.part < b.part
	})
	for _, sc := range sortedCueWithin {
		sel := selectorOf(s.widget.WidgetName(), sc.key.container) + cuePseudo(sc.key.cue) +
			" " + selectorOf(s.widget.WidgetName(), sc.key.part)
		statesSB.WriteString(formatRule([]string{sel}, sc.decls))
	}

	states := statesSB.GetString(fmt.BuffOut)
	if len(states) > 0 {
		sb.WriteString("@layer states {\n")
		sb.WriteString(states)
		sb.WriteString("}\n\n")
	}
	sb.WriteString(startingStyleBlock(startSels))

	s.emitAcrossRules(sb)

	return hoverCues
}
