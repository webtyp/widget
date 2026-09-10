//go:build !wasm

package style

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// emitPrimitives renders `@layer primitives` — the layout half of the sheet:
// every flow (Stack/Row/Split/Grid/Sidebar/SlideDeck/MasterDetail/AutoRotate…)
// and every boolean layout flag (Fill/Grow/Scroll/KeepSize/EdgeToEdge…),
// grouped so parts that ask for the same thing share one rule. It also emits
// the AutoRotate @keyframes that follow the layer.
//
// It returns the AutoRotate selectors: emit_motion.go needs them to add an
// `animation: none` reduced-motion rule, and only this pass knows which parts
// used the flow.
func (s *Sheet) emitPrimitives(sb *fmt.Conv, parts []widget.Part) (autoRotateSels []string) {
	var stackSel, rowSel, splitSel, gridSel, fixedGridSel, centerSel, fillCenteredSel, scrollRowSel, mediaBoxSel []string
	var coverSels []string

	var sidebarInfos []sidebarInfo
	var masterDetailInfos []masterDetailInfo
	var slideDeckInfos []slideDeckInfo

	var fillSel, growSel, pushEndSel, scrollSel, keepSizeSel, edgeToEdgeSel, hideOverflowSel []string
	var iconCapSel []string
	var dividerStartSel, dividerEndSel, dividerBelowSel, dividerBetweenSel []string

	type scrollGutterInfo struct {
		sel    string
		gutter Space
	}
	var scrollGutterInfos []scrollGutterInfo

	collect := func(r rule, sel string) {
		if r.hasFlow {
			switch r.flowType {
			case flowStack:
				stackSel = append(stackSel, sel)
			case flowRow:
				rowSel = append(rowSel, sel)
			case flowSplit:
				splitSel = append(splitSel, sel)
			case flowGrid:
				gridSel = append(gridSel, sel)
			case flowFixedGrid:
				fixedGridSel = append(fixedGridSel, sel)
			case flowCenter:
				centerSel = append(centerSel, sel)
			case flowFillCentered:
				fillCenteredSel = append(fillCenteredSel, sel)
			case flowScrollRow:
				scrollRowSel = append(scrollRowSel, sel)
			case flowMediaBox:
				mediaBoxSel = append(mediaBoxSel, sel)
			case flowCover:
				coverSels = append(coverSels, sel)
			case flowSidebar:
				sidebarInfos = append(sidebarInfos, sidebarInfo{sel: sel, side: r.flowSide})
			case flowSlideDeck:
				slideDeckInfos = append(slideDeckInfos, slideDeckInfo{sel: sel, motion: r.flowMotion})
			case flowAutoRotate:
				autoRotateSels = append(autoRotateSels, sel)
			case flowMasterDetail:
				masterDetailInfos = append(masterDetailInfos, masterDetailInfo{sel: sel, detail: r.flowDetail})
			}
		}
		if r.iconCap {
			iconCapSel = append(iconCapSel, sel)
		}
		if r.fill {
			fillSel = append(fillSel, sel)
		}
		if r.grow {
			growSel = append(growSel, sel)
		}
		if r.pushEnd {
			pushEndSel = append(pushEndSel, sel)
		}
		if r.scroll {
			if r.hasScrollGutter {
				scrollGutterInfos = append(scrollGutterInfos, scrollGutterInfo{sel: sel, gutter: r.scrollGutter})
			} else {
				scrollSel = append(scrollSel, sel)
			}
		}
		if r.keepSize {
			keepSizeSel = append(keepSizeSel, sel)
		}
		if r.edgeToEdge {
			edgeToEdgeSel = append(edgeToEdgeSel, sel)
		}
		if r.hideOverflow {
			hideOverflowSel = append(hideOverflowSel, sel)
		}
		if r.hasDivider {
			if r.dividerSide == SideStart {
				dividerStartSel = append(dividerStartSel, sel)
			} else {
				dividerEndSel = append(dividerEndSel, sel)
			}
		}
		if r.hasDividerBelow {
			dividerBelowSel = append(dividerBelowSel, sel)
		}
		// The child combinator is what makes this a BETWEEN rather than an
		// under: "every child that has a sibling before it" is exactly the set
		// of gaps in the list, so the first child is skipped and no line is
		// left dangling past the last row.
		if r.hasDividerBetween {
			dividerBetweenSel = append(dividerBetweenSel, sel+" > * + *")
		}
	}

	collect(s.rootRule, selectorOf(s.widget.WidgetName(), ""))
	for _, p := range parts {
		collect(s.partRules[p], selectorOf(s.widget.WidgetName(), p))
	}

	primitivesSB := fmt.GetConv()
	defer primitivesSB.PutConv()

	emitPrimitive := func(extraSel []string, decls []string) {
		if len(extraSel) > 0 {
			primitivesSB.WriteString(formatRule(extraSel, decls))
		}
	}

	// gap on the container, not margin-block-start on the children: a child rule
	// reading var(--gap) resolves it against the CHILD, so any child that is
	// itself a flow container declares its own --gap and silently replaces the
	// separation its parent asked for. A child whose gap is SpaceNone collapses
	// the parent's spacing to zero. On the container the variable resolves where
	// it was declared.
	emitPrimitive(stackSel, []string{
		"display: flex;",
		"flex-direction: column;",
		"gap: var(--gap);",
		"min-height: 0;",
	})

	emitPrimitive(rowSel, []string{
		"display: flex;",
		"flex-wrap: wrap;",
		"gap: var(--gap);",
		"align-items: center;",
	})

	emitPrimitive(splitSel, []string{
		"display: flex;",
		"flex-wrap: wrap;",
		"gap: var(--gap);",
	})
	if len(splitSel) > 0 {
		var splitKids []string
		var splitFirst []string
		for _, sel := range splitSel {
			splitKids = append(splitKids, sel+" > *")
			splitFirst = append(splitFirst, sel+" > :first-child")
		}
		emitPrimitive(splitKids, []string{
			"flex-grow: 1;",
			"flex-basis: calc((40rem - 100%) * 999);",
		})
		emitPrimitive(splitFirst, []string{
			"flex-grow: var(--ratio);",
		})
	}

	emitPrimitive(gridSel, []string{
		"display: grid;",
		"gap: var(--gap);",
		"grid-template-columns: repeat(auto-fit, minmax(min(var(--track), 100%), 1fr));",
	})

	emitPrimitive(fixedGridSel, []string{
		"display: grid;",
		"gap: var(--gap);",
		"grid-template-columns: repeat(var(--cols), minmax(0, 1fr));",
	})

	emitPrimitive(centerSel, []string{
		"margin-inline: auto;",
		"max-width: var(--max-width);",
		"width: 100%;",
	})

	emitPrimitive(fillCenteredSel, []string{
		"display: grid;",
		"place-items: center;",
		"min-height: 100%;",
		"width: 100%;",
	})

	emitPrimitive(scrollRowSel, []string{
		"display: flex;",
		"gap: var(--gap);",
		"overflow-x: auto;",
		"scroll-snap-type: x mandatory;",
		"scroll-behavior: smooth;",
	})
	if len(scrollRowSel) > 0 {
		var scrollRowKids []string
		for _, sel := range scrollRowSel {
			scrollRowKids = append(scrollRowKids, sel+" > *")
		}
		emitPrimitive(scrollRowKids, []string{
			"scroll-snap-align: start;",
			"flex: 0 0 auto;",
		})
	}

	emitPrimitive(mediaBoxSel, []string{
		"aspect-ratio: var(--ratio);",
		"overflow: hidden;",
		"display: flex;",
		"justify-content: center;",
		"align-items: center;",
	})
	if len(mediaBoxSel) > 0 {
		var mediaBoxKids []string
		for _, sel := range mediaBoxSel {
			mediaBoxKids = append(mediaBoxKids, sel+" > img", sel+" > video")
		}
		emitPrimitive(mediaBoxKids, []string{
			"width: 100%;",
			"height: 100%;",
			"object-fit: cover;",
		})
	}

	// The glyph is half its cap. Emitted BY THE CAP, as a child rule, so a
	// component cannot answer the question differently — the bug this recipe
	// exists to make unrepresentable. 50% of --control-height, not an
	// IconSize step: IconSize is relative to the font (1.5em, 2.5em) and a cap
	// is measured in control heights, so a font-relative glyph inside it
	// drifts the moment either scale moves. Same shape as MediaBox's
	// `> img, > video` rule above.
	if len(iconCapSel) > 0 {
		var iconCapKids []string
		for _, sel := range iconCapSel {
			iconCapKids = append(iconCapKids, sel+" > svg")
		}
		emitPrimitive(iconCapKids, []string{
			"width: 50%;",
			"height: 50%;",
			"flex-shrink: 0;",
		})
	}

	emitPrimitive(coverSels, []string{
		"height: 100dvh;",
		"display: flex;",
		"flex-direction: column;",
	})

	emitSidebarGroups(primitivesSB, sidebarInfos)
	emitSlideDeckGroups(primitivesSB, slideDeckInfos)
	emitAutoRotateGroups(primitivesSB, autoRotateSels)
	emitMasterDetailGroups(primitivesSB, masterDetailInfos)
	emitPrimitive(fillSel, []string{
		"height: 100%;",
		"min-height: 0;",
		"flex-grow: 1;",
	})
	emitPrimitive(growSel, []string{
		"flex-grow: 1;",
		"min-width: 0;",
	})
	emitPrimitive(pushEndSel, []string{
		"margin-inline-start: auto;",
	})
	emitPrimitive(scrollSel, append([]string{
		"overflow-y: auto;",
		"height: 100%;",
		"min-height: 0;",
		"flex-grow: 1;",
	}, floatingPadDecls()...))
	// One rule per gutter value rather than a shared bucket: unlike the plain
	// scrollSel above, these carry a value (the gutter) baked straight into
	// the calc, so two parts asking for a different Space could not share one
	// rule body. In practice this is a handful of selectors ecosystem-wide.
	for _, gi := range scrollGutterInfos {
		emitPrimitive([]string{gi.sel}, append([]string{
			"overflow-y: auto;",
			"height: 100%;",
			"min-height: 0;",
			"flex-grow: 1;",
		}, floatingPadDeclsWithGutter(gi.gutter)...))
	}
	emitPrimitive(keepSizeSel, []string{
		"flex-shrink: 0;",
		"flex-grow: 0;",
	})
	emitPrimitive(edgeToEdgeSel, []string{
		"margin: 0;",
		"border-radius: 0;",
	})
	emitPrimitive(hideOverflowSel, []string{
		"overflow: hidden;",
	})
	emitPrimitive(dividerStartSel, dividerDecls(SideStart))
	emitPrimitive(dividerEndSel, dividerDecls(SideEnd))
	emitPrimitive(dividerBelowSel, dividerBelowDecls())
	emitPrimitive(dividerBetweenSel, dividerBetweenDecls())

	prims := primitivesSB.GetString(fmt.BuffOut)
	if len(prims) > 0 {
		sb.WriteString("@layer primitives {\n")
		sb.WriteString(prims)
		sb.WriteString("}\n\n")
	}

	if len(autoRotateSels) > 0 {
		sb.WriteString(autoRotateKeyframesCSS())
	}

	return autoRotateSels
}
