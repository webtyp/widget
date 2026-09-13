//go:build !wasm

package style

import "strings"

// surfaceDecls resolves the rule's Surface into its declarations and
// composes the one box-shadow decision point: Raise()'s elevation and a
// state border's ring both paint through box-shadow, and the two combine
// here — ring first, elevation after — or the later declaration would
// silently stomp the earlier one.
func (r rule) surfaceDecls() []string {
	var decls []string

	var t triplet
	if r.hasSurface {
		t = r.surface.resolve()
		if t.bg != "" {
			// Static first, enhanced second: a browser without light-dark()/
			// color-mix() support drops the second (invalid at parse time) and
			// keeps the first — permanently the light theme. A supporting
			// browser applies the second, which wins by being later.
			if t.bgStatic != "" {
				decls = append(decls, "background-color: "+t.bgStatic+";")
			}
			decls = append(decls, "background-color: "+t.bg+";")
			// Always present, always inert until an app calls
			// css.Theme(css.SetGradient(...)) for this surface's family
			// token: background-image paints over background-color rather
			// than replacing it, so this costs nothing for the (default)
			// solid case and needs no per-surface opt-in flag.
			//
			// Unless the surface has no family: familyBase returns the zero
			// Token for the derived ones (AccentWash, AccentInverse,
			// AccentHover, Inactive), and its ImageVarName() is the bare
			// "-image" — not a custom property name at all, since those must
			// start with two dashes. Browsers discard it, so the bug was
			// invisible; it was still shipping junk in every stylesheet.
			//
			// A derived surface still emits `background-image: none`, though: it
			// is a flat fill with no family image of its own, and a lower-layer
			// rule's family image can otherwise bleed through. A nav item filled
			// As(Primary) in @layer widgets and overridden As(AccentInverse) in
			// @layer states keeps the widgets-layer `background-image:
			// var(--color-primary-image, none)` unless the states rule sets the
			// longhand too — under css.SetGradient that is the gradient painting
			// over the amber "current" fill.
			if family := familyBase(r.surface); family.Name != "" {
				decls = append(decls, "background-image: var("+family.ImageVarName()+", none);")
				// GradientAngle: repaint the family gradient at this surface's
				// own angle. Emitted AFTER the line above so it wins when
				// valid; when the app set no gradient the -image-stops var is
				// unset, `linear-gradient(<angle>, )` is invalid, this
				// declaration is dropped, and the `var(--x-image, none)`
				// above still applies. Double-declaration fallback, no flag.
				if r.hasGradientAngle {
					decls = append(decls, "background-image: linear-gradient("+r.gradientAngle+", var("+family.ImageStopsVarName()+"));")
				}
			} else {
				decls = append(decls, "background-image: none;")
			}
		}
		// edgeToEdge's border-radius: 0 lives in the primitives layer, which the
		// widgets layer outranks — a default radius emitted here would win over
		// it and leave the box rounded against the frame (the crudview root
		// measured 4px with EdgeToEdge already applied). An explicit Round()
		// still wins: it is emitted in the widgets layer, same as the surface.
		if radius := r.resolvedRadius(); radius != RadiusNone && !r.hasRound && !r.edgeToEdge {
			decls = append(decls, "border-radius: "+radiusVar(radius)+";")
		}
		if t.text != "" {
			if t.textStatic != "" {
				decls = append(decls, "color: "+t.textStatic+";")
			}
			decls = append(decls, "color: "+t.text+";")
			// Safari paints a :disabled control's text with its own
			// -webkit-text-fill-color, which WINS over color — a read-only
			// field styled for legibility still came out washed grey on iOS
			// while every other browser honoured the color above. Only state
			// rules pay for the extra pair: disabled/locked is a state, and a
			// surface's promise to resolve text together with its background
			// is not kept on the one platform where the text half is silently
			// overridden.
			if r.overlay {
				if t.textStatic != "" {
					decls = append(decls, "-webkit-text-fill-color: "+t.textStatic+";")
				}
				decls = append(decls, "-webkit-text-fill-color: "+t.text+";")
			}
		}
		if t.border != "" && !r.overlay {
			// A state rule repaints its border as a shadow ring through
			// boxShadowDecls below — never as a border, which would grow the
			// box exactly when the pointer entered it.
			if t.borderStatic != "" {
				decls = append(decls, "border: "+t.borderStatic+";")
			}
			decls = append(decls, "border: "+t.border+";")
		}
	}

	// The one box-shadow decision point: Raise() and a state border both paint
	// through box-shadow, and the two compose here — ring first, elevation
	// after — or the later declaration would silently stomp the earlier one.
	return append(decls, boxShadowDecls(r, t)...)
}

func boxShadowDecls(r rule, t triplet) []string {
	if t.border != "" && r.overlay {
		if r.hasRaise {
			// One declaration, and the ring's color is the token's Var() form,
			// not the static+enhanced pair: the elevation's var() inside the
			// light-dark() half would defer the WHOLE declaration to
			// computed-value time (the parse-time-safe rule), silently
			// discarding the static half on a browser without light-dark().
			// A bare var() with its hex fallback is safe on every engine: a
			// legacy browser cannot compute --color-outline's light-dark()
			// :root value, so the fallback applies.
			return []string{"box-shadow: " + ringShadow(t.borderVar) + ", " + elevationVar(r.raise) + ";"}
		}
		if t.borderStatic != "" {
			// Static ring first, enhanced ring second — the same double
			// declaration every other themed color uses.
			return []string{
				"box-shadow: " + ringShadow(t.borderStatic) + ";",
				"box-shadow: " + ringShadow(t.border) + ";",
			}
		}
		return []string{"box-shadow: " + ringShadow(t.border) + ";"}
	}
	if r.hasRaise {
		return []string{"box-shadow: " + elevationVar(r.raise) + ";"}
	}
	return nil
}

// ringShadow turns a border value ("1px solid <color>") into the shadow ring
// a state border paints with: 0 0 0 1px <color>.
func ringShadow(border string) string {
	return "0 0 0 1px " + strings.TrimPrefix(border, borderStyle)
}
