//go:build !wasm

package style

import (
	"sort"

	"webtyp.com/fmt"
)

// emitAcrossRules renders the CueAcross and StateAcross rules UNLAYERED, after
// the states layer has closed. Their whole job is to override a RevealedBy
// reveal, which the device path also emits unlayered, and an @layer rule —
// however specific — always loses to an unlayered one. Their :has()-from-root
// selector is already more specific than a bare `.n__part[data-x]`, so among
// unlayered rules they win on merit.
func (s *Sheet) emitAcrossRules(sb *fmt.Conv) {

	// The one selector that spans an arbitrary DOM relationship: a cue on a
	// region styles a part anywhere, checked from the root via :has(). Built
	// here, emitted below — after the states layer closes.
	type cueAcrossEntry struct {
		key   cueAcrossKey
		decls []string
	}
	var sortedCueAcross []cueAcrossEntry
	for k, r := range s.cueAcross {
		var d []string
		if r.hasFlow {
			d = append(d, flowSelfDecls(r)...)
		}
		d = append(d, r.Decls(s.widget.WidgetKind().Layer())...)
		d = append(d, primitiveDecls(r)...)
		if len(d) == 0 {
			continue
		}
		sortedCueAcross = append(sortedCueAcross, cueAcrossEntry{key: k, decls: d})
	}
	sort.Slice(sortedCueAcross, func(i, j int) bool {
		a, b := sortedCueAcross[i].key, sortedCueAcross[j].key
		if a.cue != b.cue {
			return a.cue < b.cue
		}
		if a.region != b.region {
			return a.region < b.region
		}
		return a.part < b.part
	})

	// CueAcross is emitted UNLAYERED, after the states layer: its whole job is
	// to override a RevealedBy reveal (which the device path also emits
	// unlayered), and an @layer rule — however specific — always loses to an
	// unlayered one. The :has()-on-root selector it builds is already more
	// specific than a bare `.n__part[data-x]`, so among unlayered rules it
	// wins on its own merits; a plain part rule it should not fight anyway.
	for _, sc := range sortedCueAcross {
		sel := selectorOf(s.widget.WidgetName(), "") +
			":has(" + selectorOf(s.widget.WidgetName(), sc.key.region) + cuePseudo(sc.key.cue) + ") " +
			selectorOf(s.widget.WidgetName(), sc.key.part)
		sb.WriteString(formatRule([]string{sel}, sc.decls))
	}
	if len(sortedCueAcross) > 0 {
		sb.WriteString("\n")
	}

	// StateAcross: same unlayered :has()-from-root shape, but the region is
	// probed for a DESCENDANT carrying the written state.
	type stateAcrossEntry struct {
		key   stateAcrossKey
		decls []string
	}
	var sortedStateAcross []stateAcrossEntry
	for k, r := range s.stateAcross {
		var d []string
		if r.hasFlow {
			d = append(d, flowSelfDecls(r)...)
		}
		d = append(d, r.Decls(s.widget.WidgetKind().Layer())...)
		d = append(d, primitiveDecls(r)...)
		if len(d) == 0 {
			continue
		}
		sortedStateAcross = append(sortedStateAcross, stateAcrossEntry{key: k, decls: d})
	}
	sort.Slice(sortedStateAcross, func(i, j int) bool {
		a, b := sortedStateAcross[i].key, sortedStateAcross[j].key
		if a.state != b.state {
			return a.state < b.state
		}
		if a.region != b.region {
			return a.region < b.region
		}
		return a.part < b.part
	})
	for _, sa := range sortedStateAcross {
		attr := sa.key.state.Attr()
		sel := selectorOf(s.widget.WidgetName(), "") +
			":has(" + selectorOf(s.widget.WidgetName(), sa.key.region) +
			" [" + attr.Key() + "=\"" + attr.Value() + "\"]) " +
			selectorOf(s.widget.WidgetName(), sa.key.part)
		sb.WriteString(formatRule([]string{sel}, sa.decls))
	}
	if len(sortedStateAcross) > 0 {
		sb.WriteString("\n")
	}

}
