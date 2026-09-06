//go:build !wasm

package style

import (
	"sort"

	"webtyp.com/css"
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// emitDevices renders the per-device sections — one `@media <query> { … }` block
// per viewport class that On() / OnlyOn() named, in ascending css.Device order.
// Inside each: an `@layer widgets` block for the device's flow + primitive
// flags + declarations, then (unlayered, like the main states path) the reveal
// for a RevealedBy on that device rule.
func (s *Sheet) emitDevices(sb *fmt.Conv) {
	var deviceOrder []css.Device
	for dk := range s.deviceRules {
		deviceOrder = append(deviceOrder, dk.device)
	}
	sort.Slice(deviceOrder, func(i, j int) bool {
		return deviceOrder[i] < deviceOrder[j]
	})
	var deduped []css.Device
	for _, d := range deviceOrder {
		if len(deduped) == 0 || deduped[len(deduped)-1] != d {
			deduped = append(deduped, d)
		}
	}

	for _, d := range deduped {
		var deviceParts []deviceKey
		for dk := range s.deviceRules {
			if dk.device == d {
				deviceParts = append(deviceParts, dk)
			}
		}
		sort.Slice(deviceParts, func(i, j int) bool {
			return deviceParts[i].part < deviceParts[j].part
		})

		devSB := fmt.GetConv()

		var devStartSels []string
		for _, dk := range deviceParts {
			r := s.deviceRules[dk]
			sel := selectorOf(s.widget.WidgetName(), dk.part)

			devWidSB := fmt.GetConv()
			if r.hasFlow {
				switch r.flowType {
				case flowStack:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "flex-direction: column;", "gap: var(--gap);", "min-height: 0;"}))
				case flowRow:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);", "align-items: center;"}))
				case flowSplit:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);"}))
					devWidSB.WriteString(formatRule([]string{sel + " > *"}, []string{"flex-grow: 1;", "flex-basis: calc((40rem - 100%) * 999);"}))
					devWidSB.WriteString(formatRule([]string{sel + " > :first-child"}, []string{"flex-grow: var(--ratio);"}))
				case flowGrid:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: grid;", "gap: var(--gap);", "grid-template-columns: repeat(auto-fit, minmax(min(var(--track), 100%), 1fr));"}))
				case flowFixedGrid:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: grid;", "gap: var(--gap);", "grid-template-columns: repeat(var(--cols), minmax(0, 1fr));"}))
				case flowCenter:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"margin-inline: auto;", "max-width: var(--max-width);", "width: 100%;"}))
				case flowFillCentered:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: grid;", "place-items: center;", "min-height: 100%;", "width: 100%;"}))
				case flowScrollRow:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "gap: var(--gap);", "overflow-x: auto;", "scroll-snap-type: x mandatory;", "scroll-behavior: smooth;"}))
					devWidSB.WriteString(formatRule([]string{sel + " > *"}, []string{"scroll-snap-align: start;", "flex: 0 0 auto;"}))
				case flowMediaBox:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"aspect-ratio: var(--ratio);", "overflow: hidden;", "display: flex;", "justify-content: center;", "align-items: center;"}))
					devWidSB.WriteString(formatRule([]string{sel + " > img", sel + " > video"}, []string{"width: 100%;", "height: 100%;", "object-fit: cover;"}))
				case flowCover:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"height: 100dvh;", "display: flex;", "flex-direction: column;"}))
				case flowSlideDeck:
					devWidSB.WriteString(formatRule([]string{sel}, slideDeckStripDecls()))
					devWidSB.WriteString(formatRule([]string{sel + " > *"}, slideDeckPageDecls(r.flowMotion)))
					cur := widget.Current.Attr()
					devWidSB.WriteString(formatRule(
						[]string{sel + ` > *[` + cur.Key() + `="` + cur.Value() + `"]`},
						slideDeckCurrentDecls(r.flowMotion)))
				case flowMasterDetail:
					devWidSB.WriteString(formatRule([]string{sel}, masterDetailStripDecls()))
					devWidSB.WriteString(formatRule([]string{sel + " > *"}, masterDetailResetDecls()))
					devWidSB.WriteString(formatRule([]string{sel + " > :nth-child(1)"}, masterDetailDetailDecls(r.flowDetail)))
					devWidSB.WriteString(formatRule([]string{sel + " > :nth-child(2)"}, masterDetailMasterDecls()))
				case flowSidebar:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);"}))
					devWidSB.WriteString(formatRule([]string{sidebarRailSel(sel, r.flowSide)}, []string{"flex-basis: var(--rail);", "flex-grow: 1;"}))
					devWidSB.WriteString(formatRule([]string{sidebarContentSel(sel, r.flowSide)}, []string{"flex-basis: 0;", "flex-grow: 999;", "min-width: 50%;"}))
				}
			}

			// The primitive flags are grouped into shared selectors on the main
			// path; inside a device rule there is nothing to group with, so they
			// are emitted here. Without this an option like PushEnd() or Grow()
			// passed to On()/OnlyOn() is silently dropped.
			if p := primitiveDecls(r); len(p) > 0 {
				devWidSB.WriteString(formatRule([]string{sel}, p))
			}

			wd := r.Decls(s.widget.WidgetKind().Layer())
			if len(wd) > 0 {
				devWidSB.WriteString(formatRule([]string{sel}, wd))
			}

			devWid := devWidSB.GetString(fmt.BuffOut)
			devWidSB.PutConv()
			if len(devWid) > 0 {
				devSB.WriteString("@layer widgets {\n")
				devSB.WriteString(devWid)
				devSB.WriteString("}\n")
			}

			if r.hasRevealed {
				sk := stateKey{state: r.revealedBy, part: dk.part}
				attr := sk.state.Attr()
				stateSel := fmt.Sprintf("%s[%s=\"%s\"]", sel, attr.Key(), attr.Value())
				if r.hasDrawer {
					// A drawer slides in on a transform, never a display
					// flip — so the reveal state is the "arrived" slide, the
					// same one open and close both transition through.
					devSB.WriteString(formatRule([]string{stateSel}, drawerRevealDecls(r.drawerMotion)))
				} else {
					// The display to restore is the PART's, not the device
					// rule's: a device rule that only says "hidden here,
					// revealed by Open" declares no flow of its own, so
					// reading its flowType returned the zero value and
					// revealed a flex row as a block — the nav's links
					// reappeared stacked one per line instead of in the row
					// they are laid out as everywhere else.
					reveal := r.flowType
					if !r.hasFlow {
						reveal = s.partRules[dk.part].flowType
					}
					decls := []string{"display: " + displayFor(reveal) + ";"}
					if r.animatedReveal() {
						decls = append(decls, "opacity: 1;")
						devStartSels = append(devStartSels, stateSel)
					}
					devSB.WriteString(formatRule([]string{stateSel}, decls))
				}
			}
		}

		devSB.WriteString(startingStyleBlock(devStartSels))
		devRules := devSB.GetString(fmt.BuffOut)
		devSB.PutConv()
		if len(devRules) > 0 {
			sb.WriteString("@media " + d.Query() + " {\n")
			sb.WriteString(devRules)
			sb.WriteString("}\n")
		}
	}
}
