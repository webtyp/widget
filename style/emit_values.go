//go:build !wasm

package style

import (
	"webtyp.com/css"
)

func spaceVar(s Space) string {
	switch s {
	case SpaceNone:
		return "0"
	case Space1:
		return css.Space1.Var()
	case Space2:
		return css.Space2.Var()
	case Space3:
		return css.Space3.Var()
	case Space4:
		return css.Space4.Var()
	case Space6:
		return css.Space6.Var()
	case Space8:
		return css.Space8.Var()
	case Space12:
		return css.Space12.Var()
	default:
		return "0"
	}
}

func radiusVar(r Radius) string {
	switch r {
	case RadiusNone:
		return "0"
	case RadiusSm:
		return css.RadiusSm.Var()
	case RadiusMd:
		return css.RadiusMd.Var()
	case RadiusLg:
		return css.RadiusLg.Var()
	case RadiusFull:
		return css.RadiusFull.Var()
	default:
		return "0"
	}
}

func extentValue(e Extent) string {
	switch e {
	case ExtentHalf:
		return "50dvh"
	case ExtentMost:
		return "70dvh"
	case ExtentFull:
		return "100dvh"
	default:
		return "70dvh"
	}
}

func elevationVar(e Elevation) string {
	switch e {
	case Flat:
		return "none"
	case Raised:
		return css.ShadowSm.Var()
	case Floating:
		return css.ShadowMd.Var()
	case Popover:
		return css.ShadowLg.Var()
	default:
		return "none"
	}
}

// splitRatioValue feeds flex-grow, which takes a unitless number. An fr unit
// here is invalid at computed-value time, which silently resets flex-grow to
// its initial 0 and leaves the first partition at its content width.
func splitRatioValue(r SplitRatio) string {
	switch r {
	case SplitHalf:
		return "1"
	case SplitTwoThirds:
		return "2"
	case SplitThreeQuarters:
		return "3"
	default:
		return "1"
	}
}

func aspectValue(a Aspect) string {
	switch a {
	case AspectSquare:
		return "1/1"
	case Aspect3x2:
		return "3/2"
	case Aspect4x3:
		return "4/3"
	case Aspect16x9:
		return "16/9"
	default:
		return "1/1"
	}
}

func columnWidthValue(cw ColumnWidth) string {
	switch cw {
	case ColumnNarrow:
		return css.ColumnNarrow.Var()
	case ColumnMedium:
		return css.ColumnMedium.Var()
	case ColumnWide:
		return css.ColumnWide.Var()
	default:
		return css.ColumnNarrow.Var()
	}
}

func sizeValue(s Size) string {
	switch s {
	case Content:
		return "max-content"
	case Readable:
		return css.MaxWReadable.Var()
	case Compact:
		// A stack of controls is not prose: Readable's 65ch is the measure a
		// LINE of text wants, and a login card or a settings pane sized to it
		// reads as a form someone forgot to constrain. 24rem is the width a
		// single column of fields is legible at.
		//
		// min(), unlike every other step here, because this one is a cap and
		// not a share: the percentages resolve against the container and shrink
		// with it for free, while a bare 24rem would overflow a phone. The
		// element takes the container up to 24rem and stops.
		return "min(100%, 24rem)"
	case Third:
		return "33.33%"
	case Half:
		return "50%"
	case TwoThirds:
		return "66.66%"
	case Most:
		return "90%"
	case Full:
		return "100%"
	default:
		return "auto"
	}
}

func iconSizeValue(s IconSize) string {
	switch s {
	case IconSm:
		return "1em"
	case IconLg:
		return "2.5em"
	default:
		return "1.5em"
	}
}

func textSizeVar(ts TextSize) string {
	switch ts {
	case TextXs:
		return css.TextXs.Var()
	case TextSm:
		return css.TextSm.Var()
	case TextBase:
		return css.TextBase.Var()
	case TextLg:
		return css.TextLg.Var()
	case TextXl:
		return css.TextXl.Var()
	case Text2xl:
		return css.Text2xl.Var()
	default:
		return "inherit"
	}
}

func weightVar(w Weight) string {
	switch w {
	case WeightRegular:
		return css.FontWeightRegular.Var()
	case WeightMedium:
		return css.FontWeightMedium.Var()
	case WeightBold:
		return css.FontWeightBold.Var()
	default:
		return "inherit"
	}
}

func motionValue(m Motion) string {
	switch m {
	case MotionNone:
		return "none"
	case MotionFast:
		return "all " + css.DurationFast.Var() + " " + css.EaseInOut.Var()
	case MotionBase:
		return "all " + css.DurationBase.Var() + " " + css.EaseInOut.Var()
	case MotionSlow:
		return "all " + css.DurationSlow.Var() + " " + css.EaseInOut.Var()
	default:
		return "none"
	}
}

// turnValue maps a Turn to its CSS degrees. Quarter steps only — see Turn.
func turnValue(t Turn) string {
	switch t {
	case TurnQuarter:
		return "90deg"
	case TurnHalf:
		return "180deg"
	case TurnThreeQuarter:
		return "270deg"
	default:
		return "0deg"
	}
}

func railVar(rw RailWidth) string {
	switch rw {
	case RailNarrow:
		return css.RailNarrow.Var()
	case RailWide:
		return css.RailWide.Var()
	default:
		return css.RailNarrow.Var()
	}
}

// motionDurationVar es solo la duración de un Motion, sin propiedad ni easing.
func motionDurationVar(m Motion) string {
	switch m {
	case MotionFast:
		return css.DurationFast.Var()
	case MotionBase:
		return css.DurationBase.Var()
	case MotionSlow:
		return css.DurationSlow.Var()
	default:
		return "0s"
	}
}

func familyBase(s Surface) css.Token {
	switch s {
	case Panel, Inset, Highlight:
		return css.ColorSurface
	case Primary:
		return css.ColorPrimary
	case Accent:
		return css.ColorAccent
	case Secondary:
		return css.ColorSurface
	case Success:
		return css.ColorSuccess
	case Danger:
		return css.ColorDanger
	case Subtle:
		return css.ColorMuted
	case Page:
		// Page is the whitest surface: a white base lets Interactive(Page)
		// derive a hover/focus/press family from the page background itself.
		return css.ColorBackground
	default:
		return css.Token{}
	}
}
