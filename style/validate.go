//go:build !wasm

package style

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

func (s *Sheet) Validate() []error {
	var errs []error
	errs = s.validateParts(errs)
	errs = s.validateComposition(errs)
	errs = s.validateStates(errs)
	return errs
}

// validateParts is the declaration-integrity half of Validate(): a rule
// must reference only DECLARED parts, a declared part must actually emit
// something, and the base surface rules must satisfy their Backdrop()
// pairing.
func (s *Sheet) validateParts(errs []error) []error {
	for k := range s.stateRules {
		if k.part != "" {
			if _, exists := s.partRules[k.part]; !exists {
				errs = append(errs, fmt.Errf("sheet %s: rule for undeclared part %q", string(s.widget.WidgetName()), string(k.part)))
			}
		}
	}
	for k := range s.cueRules {
		if k.part != "" {
			if _, exists := s.partRules[k.part]; !exists {
				errs = append(errs, fmt.Errf("sheet %s: rule for undeclared part %q", string(s.widget.WidgetName()), string(k.part)))
			}
		}
	}

	for p, pr := range s.partRules {
		if pr.hidden {
			continue
		}
		if pr.emitsNothing(s.widget.WidgetKind().Layer()) {
			errs = append(errs, fmt.Errf("sheet %s: part %q emits nothing", string(s.widget.WidgetName()), string(p)))
		}
	}

	checkVeil := func(p widget.Part, r rule) {
		if r.hasVeil && !r.hasBackdrop {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Veil() requires Backdrop()", string(s.widget.WidgetName()), string(p)))
		}
	}
	checkVeil("", s.rootRule)
	for p, pr := range s.partRules {
		checkVeil(p, pr)
	}

	// DividerBetween() needs a child combinator, which only the base-rule path
	// can write: On() and the state/cue rules emit a flat declaration list with
	// no selector of their own. Declaring it there used to be accepted and then
	// emit nothing — the exact silent failure that let Divider()/DividerBelow()
	// go unnoticed on base rules for as long as they did. Fail loudly instead.
	for k, dr := range s.deviceRules {
		if dr.hasDividerBetween {
			errs = append(errs, fmt.Errf("sheet %s: part %q: DividerBetween() cannot be used inside On()/OnlyOn(); declare it on the base Part() rule", string(s.widget.WidgetName()), string(k.part)))
		}
	}
	for k, sr := range s.stateRules {
		if sr.hasDividerBetween {
			errs = append(errs, fmt.Errf("sheet %s: part %q: DividerBetween() cannot be used in a state rule; declare it on the base Part() rule", string(s.widget.WidgetName()), string(k.part)))
		}
	}
	for k, cr := range s.cueRules {
		if cr.hasDividerBetween {
			errs = append(errs, fmt.Errf("sheet %s: part %q: DividerBetween() cannot be used in a Cue() rule; declare it on the base Part() rule", string(s.widget.WidgetName()), string(k.part)))
		}
	}
	return errs
}
