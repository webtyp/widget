//go:build !wasm

package style

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// sidebarInfo describes one Sidebar() container and the rail/content geometry
// it lays out.
type sidebarInfo struct {
	sel  string
	side Side
}

// slideDeckInfo describes one SlideDeck() container and the motion its pages
// slide on.
type slideDeckInfo struct {
	sel    string
	motion Motion
}

// masterDetailInfo describes one MasterDetail() container and the size its
// detail panel claims.
type masterDetailInfo struct {
	sel    string
	detail Size
}

// emitSidebarGroups, emitSlideDeckGroups, emitAutoRotateGroups and
// emitMasterDetailGroups write the child/kin rules of the assembled flows that
// need per-container geometry. They write into the in-flight @layer primitives
// buffer, which the shared emitPrimitive closure of emitPrimitives cannot reach
// from here — each keeps its own formatRule call.
func emitSidebarGroups(sb *fmt.Conv, groups []sidebarInfo) {
	for _, si := range groups {
		sb.WriteString(formatRule([]string{si.sel}, []string{
			"display: flex;",
			"flex-wrap: wrap;",
			"gap: var(--gap);",
		}))
		sb.WriteString(formatRule([]string{sidebarRailSel(si.sel, si.side)}, []string{
			"flex-basis: var(--rail);",
			"flex-grow: 1;",
		}))
		sb.WriteString(formatRule([]string{sidebarContentSel(si.sel, si.side)}, []string{
			"flex-basis: 0;",
			"flex-grow: 999;",
			"min-width: 50%;",
		}))
	}
}

func emitSlideDeckGroups(sb *fmt.Conv, groups []slideDeckInfo) {
	for _, sd := range groups {
		sb.WriteString(formatRule([]string{sd.sel}, slideDeckStripDecls()))
		sb.WriteString(formatRule([]string{sd.sel + " > *"}, slideDeckPageDecls(sd.motion)))
		cur := widget.Current.Attr()
		sb.WriteString(formatRule(
			[]string{sd.sel + ` > *[` + cur.Key() + `="` + cur.Value() + `"]`},
			slideDeckCurrentDecls(sd.motion)))
	}
}

func emitAutoRotateGroups(sb *fmt.Conv, sels []string) {
	if len(sels) == 0 {
		return
	}
	sb.WriteString(formatRule(sels, autoRotateStripDecls()))

	var kids, firsts []string
	for _, sel := range sels {
		kids = append(kids, sel+" > *")
		firsts = append(firsts, sel+" > :first-child")
	}
	sb.WriteString(formatRule(kids, autoRotateLayerDecls()))
	sb.WriteString(formatRule(firsts, autoRotateFirstDecls()))

	for slot := 2; slot <= AutoRotateLayers; slot++ {
		var nths []string
		for _, sel := range sels {
			nths = append(nths, fmt.Sprintf("%s > :nth-child(%d)", sel, slot))
		}
		sb.WriteString(formatRule(nths, autoRotateDelayDecls(slot)))
	}
}

func emitMasterDetailGroups(sb *fmt.Conv, groups []masterDetailInfo) {
	for _, mi := range groups {
		sb.WriteString(formatRule([]string{mi.sel}, masterDetailStripDecls()))
		sb.WriteString(formatRule([]string{mi.sel + " > *"}, masterDetailResetDecls()))
		sb.WriteString(formatRule([]string{mi.sel + " > :nth-child(1)"}, masterDetailDetailDecls(mi.detail)))
		sb.WriteString(formatRule([]string{mi.sel + " > :nth-child(2)"}, masterDetailMasterDecls()))
	}
}
