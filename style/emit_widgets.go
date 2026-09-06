//go:build !wasm

package style

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// emitWidgets renders `@layer widgets` — the resting look of the root and every
// part: one rule each, from the declarations rule.Decls() produces at the
// widget kind's layer. State, cue and device variants live in the later
// sections.
func (s *Sheet) emitWidgets(sb *fmt.Conv, parts []widget.Part) {
	widgetsSB := fmt.GetConv()
	defer widgetsSB.PutConv()

	rootDecls := s.rootRule.Decls(s.widget.WidgetKind().Layer())
	if len(rootDecls) > 0 {
		widgetsSB.WriteString(formatRule([]string{selectorOf(s.widget.WidgetName(), "")}, rootDecls))
	}

	for _, p := range parts {
		partDecls := s.partRules[p].Decls(s.widget.WidgetKind().Layer())
		if len(partDecls) > 0 {
			widgetsSB.WriteString(formatRule([]string{selectorOf(s.widget.WidgetName(), p)}, partDecls))
		}
	}

	wids := widgetsSB.GetString(fmt.BuffOut)
	if len(wids) > 0 {
		sb.WriteString("@layer widgets {\n")
		sb.WriteString(wids)
		sb.WriteString("}\n\n")
	}
}
