//go:build !wasm

package style

import (
	"sort"

	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// emitReducedMotion renders the `@media (prefers-reduced-motion: reduce)`
// sections: one `transition: none` rule over every selector that carries a
// Motion (parts, states, cues, cue-within, cue-across, device rules, and both
// halves of a Drawer), and — when AutoRotate is in play — one `animation: none`
// rule over its rotating children. Selectors are sorted so the output is
// deterministic regardless of map iteration order.
func (s *Sheet) emitReducedMotion(sb *fmt.Conv, parts []widget.Part, autoRotateSels []string) {
	var motionSel []string
	if s.rootRule.hasMotion {
		motionSel = append(motionSel, selectorOf(s.widget.WidgetName(), ""))
	}
	for _, p := range parts {
		if s.partRules[p].hasMotion {
			motionSel = append(motionSel, selectorOf(s.widget.WidgetName(), p))
		}
	}
	for k, sr := range s.stateRules {
		if sr.hasMotion {
			attr := k.state.Attr()
			sel := fmt.Sprintf("%s[%s=\"%s\"]", selectorOf(s.widget.WidgetName(), k.part), attr.Key(), attr.Value())
			motionSel = append(motionSel, sel)
		}
	}
	for k, cr := range s.cueWithin {
		if cr.hasMotion {
			motionSel = append(motionSel,
				selectorOf(s.widget.WidgetName(), k.container)+cuePseudo(k.cue)+
					" "+selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueWithinHover {
		if cr.hasMotion {
			motionSel = append(motionSel,
				selectorOf(s.widget.WidgetName(), k.container)+cuePseudo(k.cue)+
					" "+selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueAcross {
		if cr.hasMotion {
			motionSel = append(motionSel,
				selectorOf(s.widget.WidgetName(), "")+
					":has("+selectorOf(s.widget.WidgetName(), k.region)+cuePseudo(k.cue)+") "+
					selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, sr := range s.stateAcross {
		if sr.hasMotion {
			attr := k.state.Attr()
			motionSel = append(motionSel,
				selectorOf(s.widget.WidgetName(), "")+
					":has("+selectorOf(s.widget.WidgetName(), k.region)+" ["+attr.Key()+"=\""+attr.Value()+"\"]) "+
					selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, dr := range s.deviceRules {
		if dr.hasMotion {
			motionSel = append(motionSel, selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueRules {
		if cr.hasMotion {
			sel := selectorOf(s.widget.WidgetName(), k.part) + cuePseudo(k.cue)
			motionSel = append(motionSel, sel)
		}
	}
	// A Drawer(...,m) transitions on both its parked base rule and its
	// RevealedBy "arrived" state, so reduced-motion must silence both.
	addDrawerMotionSel := func(r rule, part widget.Part) {
		if !r.hasDrawer || r.drawerMotion == MotionNone || !r.hasRevealed {
			return
		}
		base := selectorOf(s.widget.WidgetName(), part)
		attr := r.revealedBy.Attr()
		motionSel = append(motionSel, base,
			fmt.Sprintf("%s[%s=\"%s\"]", base, attr.Key(), attr.Value()))
	}
	addDrawerMotionSel(s.rootRule, "")
	for _, p := range parts {
		addDrawerMotionSel(s.partRules[p], p)
	}
	for k, dr := range s.deviceRules {
		addDrawerMotionSel(dr, k.part)
	}

	addAppBarMotionSel := func(r rule, part widget.Part) {
		if !r.hasAppBar || r.appBarMotion == MotionNone {
			return
		}
		base := selectorOf(s.widget.WidgetName(), part)
		motionSel = append(motionSel, base)
	}
	addAppBarMotionSel(s.rootRule, "")
	for _, p := range parts {
		addAppBarMotionSel(s.partRules[p], p)
	}
	for k, dr := range s.deviceRules {
		addAppBarMotionSel(dr, k.part)
	}

	if len(motionSel) > 0 {
		sort.Strings(motionSel)
		sb.WriteString("@media (prefers-reduced-motion: reduce) {\n")
		sb.WriteString(formatRule(motionSel, []string{"transition: none;"}))
		sb.WriteString("}\n")
	}

	var animSel []string
	if s.rootRule.animated() {
		animSel = append(animSel, selectorOf(s.widget.WidgetName(), ""))
	}
	for _, p := range parts {
		if s.partRules[p].animated() {
			animSel = append(animSel, selectorOf(s.widget.WidgetName(), p))
		}
	}
	for k, sr := range s.stateRules {
		if sr.animated() {
			attr := k.state.Attr()
			sel := fmt.Sprintf("%s[%s=\"%s\"]", selectorOf(s.widget.WidgetName(), k.part), attr.Key(), attr.Value())
			animSel = append(animSel, sel)
		}
	}
	for k, cr := range s.cueWithin {
		if cr.animated() {
			animSel = append(animSel,
				selectorOf(s.widget.WidgetName(), k.container)+cuePseudo(k.cue)+
					" "+selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueWithinHover {
		if cr.animated() {
			animSel = append(animSel,
				selectorOf(s.widget.WidgetName(), k.container)+cuePseudo(k.cue)+
					" "+selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueAcross {
		if cr.animated() {
			animSel = append(animSel,
				selectorOf(s.widget.WidgetName(), "")+
					":has("+selectorOf(s.widget.WidgetName(), k.region)+cuePseudo(k.cue)+") "+
					selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, sr := range s.stateAcross {
		if sr.animated() {
			attr := k.state.Attr()
			animSel = append(animSel,
				selectorOf(s.widget.WidgetName(), "")+
					":has("+selectorOf(s.widget.WidgetName(), k.region)+" ["+attr.Key()+"=\""+attr.Value()+"\"]) "+
					selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, dr := range s.deviceRules {
		if dr.animated() {
			animSel = append(animSel, selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueRules {
		if cr.animated() {
			sel := selectorOf(s.widget.WidgetName(), k.part) + cuePseudo(k.cue)
			animSel = append(animSel, sel)
		}
	}

	for _, sel := range autoRotateSels {
		animSel = append(animSel, sel+" > *")
	}

	var frameHiddenSel []string
	if s.rootRule.hasFrame && s.rootRule.frameIndex > 0 {
		frameHiddenSel = append(frameHiddenSel, selectorOf(s.widget.WidgetName(), ""))
	}
	for _, p := range parts {
		if s.partRules[p].hasFrame && s.partRules[p].frameIndex > 0 {
			frameHiddenSel = append(frameHiddenSel, selectorOf(s.widget.WidgetName(), p))
		}
	}
	for k, sr := range s.stateRules {
		if sr.hasFrame && sr.frameIndex > 0 {
			attr := k.state.Attr()
			sel := fmt.Sprintf("%s[%s=\"%s\"]", selectorOf(s.widget.WidgetName(), k.part), attr.Key(), attr.Value())
			frameHiddenSel = append(frameHiddenSel, sel)
		}
	}
	for k, cr := range s.cueWithin {
		if cr.hasFrame && cr.frameIndex > 0 {
			frameHiddenSel = append(frameHiddenSel,
				selectorOf(s.widget.WidgetName(), k.container)+cuePseudo(k.cue)+
					" "+selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueWithinHover {
		if cr.hasFrame && cr.frameIndex > 0 {
			frameHiddenSel = append(frameHiddenSel,
				selectorOf(s.widget.WidgetName(), k.container)+cuePseudo(k.cue)+
					" "+selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueAcross {
		if cr.hasFrame && cr.frameIndex > 0 {
			frameHiddenSel = append(frameHiddenSel,
				selectorOf(s.widget.WidgetName(), "")+
					":has("+selectorOf(s.widget.WidgetName(), k.region)+cuePseudo(k.cue)+") "+
					selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, sr := range s.stateAcross {
		if sr.hasFrame && sr.frameIndex > 0 {
			attr := k.state.Attr()
			frameHiddenSel = append(frameHiddenSel,
				selectorOf(s.widget.WidgetName(), "")+
					":has("+selectorOf(s.widget.WidgetName(), k.region)+" ["+attr.Key()+"=\""+attr.Value()+"\"]) "+
					selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, dr := range s.deviceRules {
		if dr.hasFrame && dr.frameIndex > 0 {
			frameHiddenSel = append(frameHiddenSel, selectorOf(s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueRules {
		if cr.hasFrame && cr.frameIndex > 0 {
			sel := selectorOf(s.widget.WidgetName(), k.part) + cuePseudo(k.cue)
			frameHiddenSel = append(frameHiddenSel, sel)
		}
	}

	if len(animSel) > 0 {
		sort.Strings(animSel)
		sb.WriteString("@media (prefers-reduced-motion: reduce) {\n")
		sb.WriteString(formatRule(animSel, []string{"animation: none;"}))
		if len(frameHiddenSel) > 0 {
			sort.Strings(frameHiddenSel)
			sb.WriteString(formatRule(frameHiddenSel, []string{"opacity: 0;"}))
		}
		sb.WriteString("}\n")
	}
}
