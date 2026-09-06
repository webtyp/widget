//go:build !wasm

package style

import (
	"webtyp.com/css"
	"webtyp.com/widget"
)

// Root defines the style for the root element of the widget.
func (s *Sheet) Root(opts ...Option) *Sheet {
	for _, opt := range opts {
		opt(&s.rootRule)
	}
	return s
}

// Part defines the style for an anatomical part of the widget.
func (s *Sheet) Part(p widget.Part, opts ...Option) *Sheet {
	r := s.partRules[p]
	for _, opt := range opts {
		opt(&r)
	}
	s.partRules[p] = r
	return s
}

// Within declares that part renders INSIDE container, and applies the options
// to part exactly as Part() would; what it adds is the containment relation,
// which the sheet needs to reason about positioning — who is whose containing
// block. It reads like the DOM: Within("menu", "options", Flyout(...)) is
// "options, inside menu".
//
// Part() remains the normal declaration. Within() is only needed where
// containment changes the result: a Flyout that hangs from an Anchor while a
// positioned part sits between them. When it matters, the sheet rejects the
// composition until the nesting is declared — see Validate().
func (s *Sheet) Within(container, p widget.Part, opts ...Option) *Sheet {
	r := s.partRules[p]
	for _, opt := range opts {
		opt(&r)
	}
	s.partRules[p] = r
	s.within[p] = container
	return s
}

// When defines the style for a part (or Root if p is "") when the widget has a specific state.
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Option) *Sheet {
	key := stateKey{state: st, part: p}
	r := s.stateRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.stateRules[key] = r
	return s
}

// WhenWithin styles a part while an ANCESTOR part carries a written state —
// `.n__container[data-x="true"] .n__part`. It is the State counterpart of
// CueWithin, and like CueWithin it is the exception, not the habit: reach for
// When() first.
//
// It exists because dom writes a state onto the element that OWNS it, which is
// not always the element that should change. A form field's read-only gate is
// written on the field, but what must stop looking editable is the control
// inside it — When(Locked, PartInput) would emit
// `.n__input[data-locked="true"]` and match nothing, since the attribute is on
// the wrapper. Pass "" as container to hang the rule off the widget root.
func (s *Sheet) WhenWithin(st widget.State, container, p widget.Part, opts ...Option) *Sheet {
	key := stateWithinKey{state: st, container: container, part: p}
	r := s.stateWithin[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.stateWithin[key] = r
	return s
}

// Cue defines the style for a part (or Root if p is "") when the browser has a cue.
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Option) *Sheet {
	key := cueKey{cue: c, part: p}
	r := s.cueRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.cueRules[key] = r
	return s
}

// CueWithin styles a part while an ANCESTOR part carries a browser cue —
// `.n__container:hover .n__part`. Reach for Cue() first; this is only for the
// case where the trigger and the thing that reacts are different elements, such
// as a rail that shows its labels while the pointer is over it.
func (s *Sheet) CueWithin(c widget.Cue, container, p widget.Part, opts ...Option) *Sheet {
	key := cueWithinKey{cue: c, container: container, part: p}
	r := s.cueWithin[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.cueWithin[key] = r
	return s
}

// CueWithinHover is CueWithin gated on the fine-pointer capability: the same
// descendant selector, emitted inside `@media (hover: hover)`. A touch tap
// fires `:hover` and synthetic mouse events, so a hover reveal that is not
// scoped this way misfires on a phone — the exact reason this variant exists.
func (s *Sheet) CueWithinHover(c widget.Cue, container, p widget.Part, opts ...Option) *Sheet {
	key := cueWithinKey{cue: c, container: container, part: p}
	r := s.cueWithinHover[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.cueWithinHover[key] = r
	return s
}

// CueAcross styles a part while some REGION part carries a browser cue, with NO
// assumed DOM relationship between them — emitted as
// `.n:has(.n__region:cue) .n__part`, checked from the widget root via :has().
//
// It is the escape hatch for what CueWithin (strict descendant) cannot express:
// floating chrome that has to yield while the module content region has focus
// within it, when the chrome and the region sit in unrelated branches of the
// tree. Reach for Cue() or CueWithin() first; this one carries a :has() on the
// root, re-evaluated on every matching state change.
//
// Pair widget.FocusWithin as the cue to mean "while anything inside region is
// focused".
func (s *Sheet) CueAcross(c widget.Cue, region, part widget.Part, opts ...Option) *Sheet {
	key := cueAcrossKey{cue: c, region: region, part: part}
	r := s.cueAcross[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.cueAcross[key] = r
	return s
}

// StateAcross is CueAcross for a WRITTEN state: it styles part while region
// CONTAINS an element carrying state — `.n:has(.n__region [data-x="true"])
// .n__part`. Use it for a state a module writes deep inside the content region
// that the chrome must also react to — the mobile hamburger staying hidden
// while a record is being edited (crudview's action button sits at
// widget.Open then), not only while a field has focus.
func (s *Sheet) StateAcross(st widget.State, region, part widget.Part, opts ...Option) *Sheet {
	key := stateAcrossKey{state: st, region: region, part: part}
	r := s.stateAcross[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.stateAcross[key] = r
	return s
}

// On defines the style for a part (or Root if p is "") only on the given viewport
// class. It is the single sanctioned way to vary a widget by device: the query
// strings live in webtyp/css and are exhaustively tested there.
//
// Reach for a flow primitive first — Split, Grid and Sidebar already reflow on
// their own. Use On only when the ARRANGEMENT itself differs, e.g. a nav rail
// that becomes a drawer.
func (s *Sheet) On(d css.Device, p widget.Part, opts ...Option) *Sheet {
	key := deviceKey{device: d, part: p}
	r := s.deviceRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	s.deviceRules[key] = r
	return s
}

// OnlyOn declares a part that exists on one viewport class and nowhere else:
// it is display:none by default and takes the given options only on d.
//
// Use it for chrome that is genuinely device-specific — a hamburger button, a
// drawer's backdrop. If the element merely CHANGES between devices rather than
// disappearing, declare it with Part() and refine it with On().
func (s *Sheet) OnlyOn(d css.Device, p widget.Part, opts ...Option) *Sheet {
	// Unconditional merge, never exists-checked: Part() merges the same way,
	// so whichever of the two runs first, the base rule keeps hidden. The
	// exists-check this replaces dropped the hide whenever Part() had
	// already created the rule — the usual order, since base options read
	// before device ones — and the "mobile-only" part painted on desktop.
	if r, exists := s.partRules[p]; exists {
		r.hidden = true
		s.partRules[p] = r
	} else {
		s.partRules[p] = rule{hidden: true}
	}
	key := deviceKey{device: d, part: p}
	r := s.deviceRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	s.deviceRules[key] = r
	return s
}
