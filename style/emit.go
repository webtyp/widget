//go:build !wasm

package style

import (
	"sort"

	"webtyp.com/css"
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

// Stylesheet renders the accumulated sheet into CSS. It is a thin orchestrator:
// each cascade section is a method in its own emit_*.go file, called here in the
// order the output must carry them —
//
//	@layer tokens, primitives, widgets, states;   (the layer declaration)
//	@layer primitives { … }                       emit_primitives.go
//	@keyframes … (AutoRotate only)                emit_primitives.go
//	@layer widgets  { … }                          emit_widgets.go
//	@layer states   { … }                          emit_states.go
//	@starting-style { … } (animated reveals only)  emit_states.go
//	.n:has(…) .n__part { … }  (CueAcross/StateAcross, UNLAYERED)  emit_states.go
//	@media (hover: hover) { @layer states { … } }  emit_hover.go
//	@media (<device>) { … }                        emit_device.go
//	@media (prefers-reduced-motion: reduce) { … }  emit_motion.go
//
// The only state threaded between sections is what a later section cannot
// recompute cheaply: the AutoRotate selectors (primitives → motion) and the
// hover cues, which include Interactive()-derived entries built while emitting
// states (states → hover).
func (s *Sheet) Stylesheet() *css.Stylesheet {
	if errs := s.Validate(); len(errs) > 0 {
		var msgs []any
		msgs = append(msgs, "widget/style:")
		for _, err := range errs {
			msgs = append(msgs, err.Error())
		}
		panic(fmt.Err(msgs...))
	}

	sb := fmt.GetConv()
	defer sb.PutConv()

	sb.WriteString("@layer tokens, primitives, widgets, states;\n\n")

	parts := s.sortedParts()

	autoRotateSels := s.emitPrimitives(sb, parts)
	s.emitWidgets(sb, parts)
	hoverCues := s.emitStates(sb, parts)
	s.emitHover(sb, hoverCues)
	s.emitDevices(sb)
	s.emitReducedMotion(sb, parts, autoRotateSels)

	return css.NewStylesheet(css.Raw(sb.GetString(fmt.BuffOut)))
}

// sortedParts is the deterministic part order every section iterates in: the
// map keys, sorted. Recomputed per call — cheap, and it keeps each section
// self-contained.
func (s *Sheet) sortedParts() []widget.Part {
	parts := make([]widget.Part, 0, len(s.partRules))
	for p := range s.partRules {
		parts = append(parts, p)
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i] < parts[j] })
	return parts
}

// cueEmission is one rendered cue rule — its key plus the declarations it
// carries. Shared between emit_states.go (which builds the set, including the
// Interactive()-derived hover/focus/press entries) and emit_hover.go (which
// emits the Hover half inside the fine-pointer media query).
type cueEmission struct {
	key   cueKey
	decls []string
}
