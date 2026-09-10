//go:build !wasm

package style

import (
	"webtyp.com/css"
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// validateStates is the legality half of Validate(): a state that the
// widget kind cannot hold, an Interactive() surface with no interaction
// family, a Drawer() without its reveal, and OnlyOn() collisions.
func (s *Sheet) validateStates(errs []error) []error {
	for k := range s.stateRules {
		if !s.widget.WidgetKind().Allows(k.state) {
			errs = append(errs, fmt.Errf("sheet %s: part %q: state %s is not meaningful for kind %s", string(s.widget.WidgetName()), string(k.part), k.state.String(), s.widget.WidgetKind().String()))
		}
	}

	// Inactive is the only surface with no interaction family: interacting
	// with the deliberately-dead shade is always a mistake. Page stays legal —
	// it is the whitest surface and a perfectly live (even default) one.
	//
	// Subtle used to be rejected here by omission: its family base was
	// --color-muted, a text token, so Interactive(Subtle) derived dark
	// backgrounds under muted text. That gap is closed in familyBase(), which
	// now returns the neutral surface family for Subtle, and in emit_states.go,
	// which repaints Subtle resting text to on-surface inside the states.
	checkInteractive := func(p widget.Part, r rule) {
		if r.interactive && r.interactiveFamily == Inactive {
			errs = append(errs, fmt.Errf("sheet %s: part %q: surface %s has no interaction states", string(s.widget.WidgetName()), string(p), r.interactiveFamily.String()))
		}
	}
	checkInteractive("", s.rootRule)
	for p, pr := range s.partRules {
		checkInteractive(p, pr)
	}

	hasDrawer := make(map[widget.Part]bool)
	hasRevealed := make(map[widget.Part]bool)
	checkPart := func(p widget.Part, r rule) {
		if r.hasDrawer {
			hasDrawer[p] = true
		}
		if r.hasRevealed {
			hasRevealed[p] = true
		}
	}
	checkPart("", s.rootRule)
	for p, pr := range s.partRules {
		checkPart(p, pr)
	}
	for p := range hasDrawer {
		if !hasRevealed[p] {
			if p == "" {
				errs = append(errs, fmt.Errf("sheet %s: root: Drawer() without RevealedBy(); the panel would be permanently visible", string(s.widget.WidgetName())))
			} else {
				errs = append(errs, fmt.Errf("sheet %s: part %q: Drawer() without RevealedBy(); the panel would be permanently visible", string(s.widget.WidgetName()), string(p)))
			}
		}
	}

	onlyOnDevices := make(map[widget.Part]css.Device)
	for dk := range s.deviceRules {
		if pr, exists := s.partRules[dk.part]; exists && pr.hidden {
			if prev, exists := onlyOnDevices[dk.part]; exists {
				errs = append(errs, fmt.Errf("sheet %s: part %q declared OnlyOn for both %s and %s", string(s.widget.WidgetName()), string(dk.part), prev.String(), dk.device.String()))
			} else {
				onlyOnDevices[dk.part] = dk.device
			}
		}
	}

	return errs
}
