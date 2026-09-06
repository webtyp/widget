//go:build !wasm

package style

import "webtyp.com/css"

// Surface is a complete visual decision: background, text, and border resolved together.
type Surface uint8

const (
	Page Surface = iota
	Panel
	Inset
	Primary
	Secondary
	Highlight
	Accent
	AccentWash
	AccentInverse
	AccentHover
	Success
	Danger
	DangerWash
	Subtle
	Bare
	Inactive
)

func (s Surface) String() string {
	switch s {
	case Page:
		return "Page"
	case Panel:
		return "Panel"
	case Inset:
		return "Inset"
	case Primary:
		return "Primary"
	case Secondary:
		return "Secondary"
	case Highlight:
		return "Highlight"
	case Accent:
		return "Accent"
	case AccentWash:
		return "AccentWash"
	case AccentInverse:
		return "AccentInverse"
	case AccentHover:
		return "AccentHover"
	case Success:
		return "Success"
	case Danger:
		return "Danger"
	case DangerWash:
		return "DangerWash"
	case Subtle:
		return "Subtle"
	case Bare:
		return "Bare"
	case Inactive:
		return "Inactive"
	default:
		return "Unknown"
	}
}

// borderStyle is the shared border shape every bordered surface uses. It is a
// named constant because the same prefix is stripped when the border is
// repainted as a state shadow ring (see ringShadow).
const borderStyle = "1px solid "

// triplet represents the combination of background, text, and border.
//
// The Static fields are the browser-safe counterparts of bg/text/border —
// see css.Token.LightValue(). emit_decls.go emits each one as the first of a
// double declaration, immediately before its light-dark()-enhanced sibling.
//
// borderVar is the border in css.Token.Var() form — the live, override-able
// reference — used only when a state border composes with an elevation into a
// single box-shadow declaration, where the static/enhanced pair cannot be
// used (see boxShadowDecls).
type triplet struct {
	bg     string
	text   string
	border string

	bgStatic     string
	textStatic   string
	borderStatic string

	borderVar string
}

// resolve translates a Surface to its corresponding triplet of styles.
func (s Surface) resolve() triplet {
	switch s {
	case Page:
		return triplet{
			bg: css.ColorBackground.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(),
			bgStatic: css.ColorBackground.LightValue(), textStatic: css.ColorOnSurface.LightValue(),
		}
	case Panel:
		return triplet{
			bg: css.ColorSurface.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(), border: borderStyle + css.ColorOutline.EnhancedVar(),
			bgStatic: css.ColorSurface.LightValue(), textStatic: css.ColorOnSurface.LightValue(), borderStatic: borderStyle + css.ColorOutline.LightValue(),
			borderVar: borderStyle + css.ColorOutline.Var(),
		}
	case Inset:
		return triplet{
			bg: css.ColorSurfaceSunken.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(), border: borderStyle + css.ColorOutline.EnhancedVar(),
			bgStatic: css.ColorSurfaceSunken.LightValue(), textStatic: css.ColorOnSurface.LightValue(), borderStatic: borderStyle + css.ColorOutline.LightValue(),
			borderVar: borderStyle + css.ColorOutline.Var(),
		}
	case Primary:
		return triplet{
			bg: css.ColorPrimary.EnhancedVar(), text: css.ColorOnPrimary.EnhancedVar(),
			bgStatic: css.ColorPrimary.LightValue(), textStatic: css.ColorOnPrimary.LightValue(),
		}
	case Secondary:
		return triplet{
			bg: css.ColorSurface.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(),
			bgStatic: css.ColorSurface.LightValue(), textStatic: css.ColorOnSurface.LightValue(),
		}
	case Highlight:
		return triplet{
			bg: css.ColorSelection.EnhancedVar(), text: css.ColorOnSelection.EnhancedVar(),
			bgStatic: css.ColorSelection.LightValue(), textStatic: css.ColorOnSelection.LightValue(),
		}
	case Accent:
		return triplet{
			bg: css.ColorAccent.EnhancedVar(), text: css.ColorOnAccent.EnhancedVar(),
			bgStatic: css.ColorAccent.LightValue(), textStatic: css.ColorOnAccent.LightValue(),
		}
	case AccentWash:
		return triplet{
			bg: css.ColorAccentWash.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(),
			bgStatic: css.ColorAccentWash.LightValue(), textStatic: css.ColorOnSurface.LightValue(),
		}
	// AccentInverse is Accent's own amber background paired with
	// ColorOnPrimary (white) instead of ColorOnAccent (near-black): the
	// dark-on-amber pairing Accent normally carries is for a large filled
	// area (a selected row) where AA text contrast matters; a small filled
	// control icon — the "add" button gone amber while its panel is open —
	// is asked here to match the white-on-color language every other
	// control surface (Primary, Success, Danger) already uses, a deliberate
	// caller choice made in spite of AccentInverse's lower contrast.
	case AccentInverse:
		return triplet{
			bg: css.ColorAccent.EnhancedVar(), text: css.ColorOnPrimary.EnhancedVar(),
			bgStatic: css.ColorAccent.LightValue(), textStatic: css.ColorOnPrimary.LightValue(),
		}
	// AccentHover pairs ColorAccentHover (Accent faded only 30%, not
	// AccentWash's 85%) with the same white ColorOnPrimary AccentInverse
	// uses: a hover preview that must stay readable in white can only fade
	// so far before the white itself loses contrast, so this is the
	// hover/hover-adjacent counterpart to AccentInverse — visibly softer
	// than the fully committed fill, never as pale as AccentWash.
	case AccentHover:
		return triplet{
			bg: css.ColorAccentHover.EnhancedVar(), text: css.ColorOnPrimary.EnhancedVar(),
			bgStatic: css.ColorAccentHover.LightValue(), textStatic: css.ColorOnPrimary.LightValue(),
		}
	case Success:
		return triplet{
			bg: css.ColorSuccess.EnhancedVar(), text: css.ColorOnSuccess.EnhancedVar(),
			bgStatic: css.ColorSuccess.LightValue(), textStatic: css.ColorOnSuccess.LightValue(),
		}
	case Danger:
		return triplet{
			bg: css.ColorDanger.EnhancedVar(), text: css.ColorOnDanger.EnhancedVar(),
			bgStatic: css.ColorDanger.LightValue(), textStatic: css.ColorOnDanger.LightValue(),
		}
	// DangerWash is the Danger counterpart of AccentWash: Danger faded 70%
	// toward transparency with on-surface text — a tint for a selection
	// state that must read as "leans toward danger" without claiming the
	// solid Danger fill, which stays reserved for a destructive commit.
	case DangerWash:
		return triplet{
			bg: css.ColorDangerWash.EnhancedVar(), text: css.ColorOnSurface.EnhancedVar(),
			bgStatic: css.ColorDangerWash.LightValue(), textStatic: css.ColorOnSurface.LightValue(),
		}
	case Subtle:
		return triplet{
			bg: "transparent", text: css.ColorMuted.EnhancedVar(),
			bgStatic: "transparent", textStatic: css.ColorMuted.LightValue(),
		}
	// Bare is the non-surface: no background, no border, and — unlike
	// Subtle — no text tint either. It exists for the device-scoped rule
	// that must strip a container's card look without pretending to be a
	// meaningful surface of its own. Subtle is for text that is meant to
	// stay muted; Bare is for "this is not a surface at all".
	case Bare:
		return triplet{bg: "transparent", bgStatic: "transparent"}
	case Inactive:
		return triplet{
			bg: css.ColorSurface.EnhancedVar(), text: css.ColorMuted.EnhancedVar(),
			bgStatic: css.ColorSurface.LightValue(), textStatic: css.ColorMuted.LightValue(),
		}
	default:
		return triplet{}
	}
}

// defaultRadius returns the default radius associated with a Surface.
func (s Surface) defaultRadius() Radius {
	switch s {
	case Page, Subtle, Bare:
		return RadiusNone
	case Panel:
		return RadiusMd
	default:
		return RadiusSm
	}
}

// Interactive applies s and derives its hover, focus, and press treatments.
func Interactive(s Surface) Option {
	return func(r *rule) {
		r.hasSurface, r.surface, r.interactive = true, s, true
	}
}
