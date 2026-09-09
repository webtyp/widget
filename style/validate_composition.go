//go:build !wasm

package style

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// validateComposition is the cross-rule half of Validate(): it makes
// loud every composition the types cannot reject — two options racing
// for the same property (position, transform), containment declared
// but sabotaged, stacking levels left undeclared, and the device-only
// rules that could never show.
func (s *Sheet) validateComposition(errs []error) []error {
	// Every one of these sets `position`, so two of them on one rule means
	// the last-emitted keyword silently wins over the other. Anchor() alongside
	// Docked()/Flyout() is the easy mistake: both of those are already
	// containing blocks, so the Anchor is redundant as well as destructive.
	checkPosition := func(p widget.Part, r rule) {
		n := 0
		for _, on := range []bool{r.hasAnchor, r.hasDocked, r.hasOnEdge, r.hasFlyout, r.hasBackdrop, r.hasDrawer, r.hasEdgeStrip, r.hasFloatMiddle} {
			if on {
				n++
			}
		}
		if n > 1 {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Anchor/Docked/OnEdge/Flyout/Backdrop/Drawer/EdgeStrip/FloatMiddle all set position; use one", string(s.widget.WidgetName()), string(p)))
		}
		if r.hasRotate && (r.hasOnEdge || r.hasDrawer || r.hasFloatMiddle) {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Rotate cannot combine with OnEdge/Drawer/FloatMiddle — all own transform", string(s.widget.WidgetName()), string(p)))
		}
	}
	// Button already paints and already claims the control height. Saying
	// either again is not additive: the second Interactive silently overwrites
	// the surface Button chose, and a second min-height from the same token
	// reads as if it did something.
	checkButton := func(p widget.Part, r rule) {
		if !r.buttonBox {
			return
		}
		if r.hasInteractive {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Button already paints an interactive surface; drop Interactive", string(s.widget.WidgetName()), string(p)))
		}
		if r.controlBox {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Button already carries the control height; drop ControlBox", string(s.widget.WidgetName()), string(p)))
		}
	}
	// Both ends of a descendant rule have to exist, or it silently styles
	// nothing. CueWithinHover carries the same obligation.
	checkCueWithin := func(method string, k cueWithinKey) {
		if _, ok := s.partRules[k.container]; !ok && k.container != "" {
			errs = append(errs, fmt.Errf("sheet %s: %s container %q is not a declared part", string(s.widget.WidgetName()), method, string(k.container)))
		}
		if _, ok := s.partRules[k.part]; !ok && k.part != "" {
			errs = append(errs, fmt.Errf("sheet %s: %s part %q is not a declared part", string(s.widget.WidgetName()), method, string(k.part)))
		}
		if k.container == k.part {
			errs = append(errs, fmt.Errf("sheet %s: %s container and part are both %q; use Cue()", string(s.widget.WidgetName()), method, string(k.part)))
		}
	}
	for k := range s.cueWithin {
		checkCueWithin("CueWithin", k)
	}
	for k := range s.cueWithinHover {
		checkCueWithin("CueWithinHover", k)
	}
	// A :has()-spanned rule styles nothing unless both ends exist somewhere in
	// the sheet (a part styled ONLY on one device — the mobile hamburger — is
	// still a real part), and both being the same means Cue()/When() was meant.
	declared := func(p widget.Part) bool {
		if p == "" {
			return true
		}
		if _, ok := s.partRules[p]; ok {
			return true
		}
		for k := range s.deviceRules {
			if k.part == p {
				return true
			}
		}
		return false
	}
	for k := range s.cueAcross {
		if !declared(k.region) {
			errs = append(errs, fmt.Errf("sheet %s: CueAcross region %q is not a declared part", string(s.widget.WidgetName()), string(k.region)))
		}
		if !declared(k.part) {
			errs = append(errs, fmt.Errf("sheet %s: CueAcross part %q is not a declared part", string(s.widget.WidgetName()), string(k.part)))
		}
		if k.region == k.part {
			errs = append(errs, fmt.Errf("sheet %s: CueAcross region and part are both %q; use Cue()", string(s.widget.WidgetName()), string(k.part)))
		}
	}
	for k := range s.stateAcross {
		if !declared(k.region) {
			errs = append(errs, fmt.Errf("sheet %s: StateAcross region %q is not a declared part", string(s.widget.WidgetName()), string(k.region)))
		}
		if !declared(k.part) {
			errs = append(errs, fmt.Errf("sheet %s: StateAcross part %q is not a declared part", string(s.widget.WidgetName()), string(k.part)))
		}
		if k.region == k.part {
			errs = append(errs, fmt.Errf("sheet %s: StateAcross region and part are both %q; use When()", string(s.widget.WidgetName()), string(k.part)))
		}
	}

	// Within() declares the part tree — which part renders inside which —
	// and the sheet reasons about positioning from it: a Flyout's
	// inset-block-start:100% resolves against the nearest POSITIONED
	// ancestor, which is not always the Anchor the author meant. These
	// checks make the two failure shapes loud instead of silent.
	for p, container := range s.within {
		if container == p {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Within() declares the part inside itself", string(s.widget.WidgetName()), string(p)))
		}
		if container != "" {
			if _, ok := s.partRules[container]; !ok {
				errs = append(errs, fmt.Errf("sheet %s: Within() container %q is not a declared part", string(s.widget.WidgetName()), string(container)))
			}
		}
	}

	// Precise theft, containment declared: climb the declared chain from
	// each Flyout part until the nearest Anchor() or the end of the chain.
	// A positioned part found BEFORE the Anchor steals the containing block
	// the 100% resolves against — name both parts. A chain that ends at a
	// positioned part with no Anchor above it is the author's explicit
	// choice of container (a docked trigger that spans its anchor) and is
	// accepted; the theft only exists when an Anchor sits beyond the thief.
	positioned := func(r rule) bool {
		return r.hasDocked || r.hasOnEdge || r.hasBackdrop || r.hasFlyout || r.hasDrawer || r.hasEdgeStrip
	}
	for p, pr := range s.partRules {
		if !pr.hasFlyout {
			continue
		}
		container, ok := s.within[p]
		if !ok {
			continue
		}
		seen := map[widget.Part]bool{}
		thief := "" // the first positioned part between the Flyout and the Anchor
		found := false
		for container != "" {
			if seen[container] {
				errs = append(errs, fmt.Errf("sheet %s: part %q: Within() containment cycle", string(s.widget.WidgetName()), string(p)))
				break
			}
			seen[container] = true
			cr, ok := s.partRules[container]
			if !ok {
				break // undeclared container already reported above
			}
			if cr.hasAnchor {
				found = true
				break
			}
			if thief == "" && positioned(cr) {
				thief = string(container)
			}
			next, ok := s.within[container]
			if !ok {
				break
			}
			container = next
		}
		if found && thief != "" {
			errs = append(errs, fmt.Errf("sheet %s: part %q: part %q is positioned between the Flyout and its Anchor and steals the containing block the Flyout hangs from; make it not positioned or re-anchor the Flyout", string(s.widget.WidgetName()), string(p), thief))
		}
	}

	// A Flyout descendant of a Scroll() region is clipped by it: overflow
	// clips every descendant that escapes the padding box, containing block
	// or not. Same two-primitive seam as the theft above — Scroll() on an
	// ancestor, Flyout() on a descendant, no type naming the crossing. Unlike
	// the theft, an Anchor above the scroller does not save it: the panel is
	// still a DOM descendant of the scroller, so the whole declared chain is
	// walked, Anchor or not. Reject the composition; the author then chooses
	// consciously — a flow accordion, or a panel out of the scroller.
	for p, pr := range s.partRules {
		if !pr.hasFlyout {
			continue
		}
		container, ok := s.within[p]
		if !ok {
			continue
		}
		seen := map[widget.Part]bool{}
		for container != "" {
			if seen[container] {
				break // cycle already reported above
			}
			seen[container] = true
			cr, ok := s.partRules[container]
			if !ok {
				break // undeclared container already reported above
			}
			if cr.scroll {
				errs = append(errs, fmt.Errf("sheet %s: part %q: part %q is a Scroll() region and clips the Flyout inside it; move the panel out of the scroller or use a flow accordion", string(s.widget.WidgetName()), string(p), string(container)))
				break
			}
			next, ok := s.within[container]
			if !ok {
				break
			}
			container = next
		}
	}

	// Ambiguous composition, containment NOT declared: a Flyout the author
	// has not nested inside anything, coexisting with a part that IS a
	// containing block (Docked(Parent, …), OnEdge, Backdrop(Parent)), cannot
	// be resolved — the docked part may sit between the Flyout and its
	// Anchor and the sheet has no way to see it. Closed by default: the
	// author declares the nesting with Within() or Validate() stays red.
	var uncontainedFlyouts []widget.Part
	for p, pr := range s.partRules {
		if pr.hasFlyout {
			if _, contained := s.within[p]; !contained {
				uncontainedFlyouts = append(uncontainedFlyouts, p)
			}
		}
	}
	if len(uncontainedFlyouts) > 0 {
		var stealers []widget.Part
		for p, pr := range s.partRules {
			if p == "" {
				continue // the root can never sit BETWEEN a Flyout and anything
			}
			if (pr.hasDocked && pr.dockedScope == Parent) || pr.hasOnEdge ||
				(pr.hasBackdrop && pr.backdropScope == Parent) ||
				(pr.hasEdgeStrip && pr.edgeStripScope == Parent) {
				stealers = append(stealers, p)
			}
		}
		for _, f := range uncontainedFlyouts {
			for _, st := range stealers {
				if f == st {
					continue
				}
				errs = append(errs, fmt.Errf("sheet %s: part %q is a Flyout and part %q is a containing block, but the sheet cannot know which contains which — declare the nesting with Within()", string(s.widget.WidgetName()), string(f), string(st)))
			}
		}
	}

	// A device rule that only paints is invisible: OnlyOn hides the part by
	// default, and inside the query only a flow, CenterContent or the state rule
	// RevealedBy generates puts a `display` back on it.
	for k, dr := range s.deviceRules {
		pr, ok := s.partRules[k.part]
		if !ok || !pr.hidden || dr.hidden {
			continue
		}
		if !dr.hasFlow && !dr.centerContent && !dr.startContent && !dr.hasRevealed {
			errs = append(errs, fmt.Errf("sheet %s: part %q is OnlyOn but its device rule declares no flow, so it can never show", string(s.widget.WidgetName()), string(k.part)))
		}
	}

	checkPosition("", s.rootRule)
	checkButton("", s.rootRule)
	for p, pr := range s.partRules {
		checkPosition(p, pr)
		checkButton(p, pr)
	}
	for k, dr := range s.deviceRules {
		checkPosition(k.part, dr)
	}

	// An element taken out of the flow must always end up with a DECLARED
	// stacking level — never `auto`, which is the silent failure stackingFor
	// exists to remove. Every out-of-flow primitive today resolves to one;
	// this guard exists so a future positioning option cannot ship without
	// its level.
	checkStacking := func(p widget.Part, r rule) {
		if (r.hasOnEdge || r.hasDocked || r.hasFlyout || r.hasBackdrop || r.hasDrawer || r.hasFloatMiddle) &&
			stackingFor(r, s.widget.WidgetKind().Layer()) == "" {
			errs = append(errs, fmt.Errf("sheet %s: part %q: out-of-flow element without a declared stacking level", string(s.widget.WidgetName()), string(p)))
		}
	}
	checkStacking("", s.rootRule)
	for p, pr := range s.partRules {
		checkStacking(p, pr)
	}
	for k, dr := range s.deviceRules {
		checkStacking(k.part, dr)
	}
	for k, sr := range s.stateRules {
		if sr.hasVeil && !sr.hasBackdrop {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Veil() requires Backdrop()", string(s.widget.WidgetName()), string(k.part)))
		}
	}
	for k, cr := range s.cueRules {
		if cr.hasVeil && !cr.hasBackdrop {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Veil() requires Backdrop()", string(s.widget.WidgetName()), string(k.part)))
		}
	}
	return errs
}
