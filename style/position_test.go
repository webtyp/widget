//go:build !wasm

package style_test

import (
	"strings"
	"testing"
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestDockedPinsToTheAnchorCorner(t *testing.T) {
	w := testWidget{name: "cv", kind: widget.Disclosure}
	s := style.For(w).
		Part("aside", style.Anchor()).
		Part("action", style.Docked(style.Parent, style.EdgeBottom, style.SideEnd, style.Space4), style.CenterContent()).
		Stylesheet().String()

	for _, want := range []string{
		"position: relative;",
		"position: absolute;",
		"inset-block-end: var(--space-4,1rem);",
		"inset-inline-end: var(--space-4,1rem);",
		"justify-content: center;",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Docked/CenterContent to emit %q, got:\n%s", want, s)
		}
	}

	// A Parent dock stays out of the overlay layer: siblings doing the same
	// would tie there, and the last in the DOM would cover the panel the first
	// one opened. It still declares a LOCAL level (z-index: 1) — leaving it
	// `auto` let Safari paint an unpositioned sibling over the pinned control
	// — and that is the whole declaration it is allowed: any other level is a
	// claim on a layer it does not own.
	if !strings.Contains(s, "z-index: 1;") {
		t.Errorf("Docked(Parent) must declare its local stacking level, got:\n%s", s)
	}
	if strings.Contains(s, "z-index: var(--z-") {
		t.Errorf("Docked(Parent) must not claim an overlay stacking level, got:\n%s", s)
	}

	// A corner pin owns all four insets: the unpinned pair has to be reset, or
	// overriding another positioning option leaves the box over-constrained and
	// it collapses instead of moving.
	for _, want := range []string{"inset-block-start: auto;", "inset-inline-start: auto;"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Docked to reset the unpinned insets with %q, got:\n%s", want, s)
		}
	}

	// A Viewport dock is a real overlay and does claim it.
	v := style.For(w).Part("fab", style.Docked(style.Viewport, style.EdgeBottom, style.SideEnd, style.Space4)).Stylesheet().String()
	if !strings.Contains(v, "z-index: var(--z-dropdown,100);") {
		t.Errorf("expected Docked(Viewport) to claim the widget layer, got:\n%s", v)
	}
}

func TestOnEdgeStraddlesTheLine(t *testing.T) {
	// A fieldset legend rides the border it labels. For EdgeTop the chip's
	// CENTRE is put on the centre of that 1px border: ChipSeat(EdgeTop)
	// reserves 0.5·--chip-height for the content's border-box top, +0.5px
	// reaches the middle of the stroke, and translateY(-50%) centres the chip
	// by its own rendered height so it lands exactly on the line regardless of
	// height. Safe as a transform because ChipSeat keeps the whole chip inside
	// the container's box (nothing for scrollHeight to miss) and the z-index
	// already opens the stacking context. EdgeBottom keeps the older
	// "onEdgeBlock is the gap from the border" negative-margin form the list
	// badges — which DO poke out — depend on.
	w := testWidget{name: "tw-field", kind: widget.Form}
	s := style.For(w).
		Root(style.Anchor(), style.ChipSeat(style.EdgeTop)).
		Part("label", style.OnEdge(style.EdgeTop, style.SideStart, style.SpaceNone, style.Space4)).
		Part("badge", style.OnEdge(style.EdgeBottom, style.SideEnd, style.SpaceNone, style.Space3)).
		Stylesheet().String()

	for _, want := range []string{
		// EdgeTop: centre on the stroke, by the chip's own height.
		"min-height: " + css.ChipHeight.Var() + ";",
		"inset-block-start: calc(0.5 * " + css.ChipHeight.Var() + " + 0.5px);",
		"transform: translateY(-50%);",
		"inset-inline-start: var(--space-4,1rem);",
		// The container reserves the space the chip seats into.
		"padding-block-start: calc(0.5 * " + css.ChipHeight.Var() + ");",
		// EdgeBottom unchanged — negative margin, not a transform.
		"inset-block-end: 0;",
		"inset-inline-end: var(--space-3,0.75rem);",
		"margin-block-end: calc(-0.5 * " + css.ChipHeight.Var() + ");",
		"margin: 0;",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected OnEdge to emit %q, got:\n%s", want, s)
		}
	}
	// EdgeTop uses translateY; EdgeBottom must NOT — its badge pokes out of the
	// row, so it needs a layout-participating margin. Check the badge rule.
	badgeRule := s[strings.Index(s, ".tw-field__badge"):]
	if i := strings.Index(badgeRule, "}"); i >= 0 {
		badgeRule = badgeRule[:i]
	}
	if strings.Contains(badgeRule, "transform") {
		t.Errorf("EdgeBottom OnEdge must not emit transform (it pokes out; needs a margin), got:\n%s", badgeRule)
	}

	// A chip is content, so it must not reach the OVERLAY layer: level with the
	// real overlays it wins on DOM order, which puts it over a dropdown. That —
	// not "no z-index at all" — is the invariant. It does order itself against
	// its own siblings with a local z-index: leaving it `auto` let Safari paint
	// a sibling form control over the legend labelling it.
	if !strings.Contains(s, "z-index: 1;") {
		t.Errorf("expected OnEdge to order itself locally with z-index: 1, got:\n%s", s)
	}
	if strings.Contains(s, "z-index: var(--z-") {
		t.Errorf("OnEdge must not claim an overlay stacking level, got:\n%s", s)
	}
}

func TestDrawerAnchorsToEdge(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Root(style.Drawer(style.SideEnd, style.TwoThirds, style.MotionSlow), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(s, "position: fixed;") {
		t.Errorf("expected Drawer to emit position: fixed:\n%s", s)
	}
	if !strings.Contains(s, "inset-block: 0;") {
		t.Errorf("expected Drawer to emit inset-block: 0:\n%s", s)
	}
	if !strings.Contains(s, "inset-inline-end: 0;") {
		t.Errorf("expected Drawer(SideEnd) to emit inset-inline-end: 0:\n%s", s)
	}
	if !strings.Contains(s, "width: 66.66%;") {
		t.Errorf("expected Drawer(TwoThirds) to emit width: 66.66%%:\n%s", s)
	}
	if !strings.Contains(s, "z-index: var(--z-base") {
		t.Errorf("expected Drawer on Region to emit z-index matching layer:\n%s", s)
	}

	// The drawer rests parked on a transform (never display:none), so the
	// RevealedBy can transition it in AND out — the exit is choreographed.
	if strings.Contains(s, "display: none;") {
		t.Errorf("a Drawer must not use display:none — it parks on a transform so it can transition:\n%s", s)
	}
	if !strings.Contains(s, "transform: translateX(100%);") {
		t.Errorf("expected Drawer(SideEnd) to park at translateX(100%%):\n%s", s)
	}
	if !strings.Contains(s, "visibility: hidden;") {
		t.Errorf("expected the parked drawer to be visibility: hidden:\n%s", s)
	}
	if !strings.Contains(s, "transition: transform var(--duration-slow") {
		t.Errorf("expected the parked drawer to carry the slide transition:\n%s", s)
	}
	// The RevealedBy state is the "arrived" slide — same transition, so close
	// animates too — never a bare display flip.
	if !strings.Contains(s, "[data-open=\"true\"] {\n  transform: translateX(0);") {
		t.Errorf("expected RevealedBy to emit the arrived slide state, not display:\n%s", s)
	}
	if strings.Contains(s, "[data-open=\"true\"] {\n  display:") {
		t.Errorf("a revealed Drawer must not restore via display — it slides:\n%s", s)
	}
	// prefers-reduced-motion silences both the parked and the arrived rule.
	if !strings.Contains(s, "@media (prefers-reduced-motion: reduce) {") {
		t.Errorf("expected an animated Drawer to register for reduced-motion:\n%s", s)
	}
}

// TestDrawerMotionNoneParksWithoutAnimation is MotionNone's contract: the
// panel still parks on a transform (so RevealedBy stays a class swap, not a
// display flip) but carries no transition.

func TestDrawerMotionNoneParksWithoutAnimation(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Root(style.Drawer(style.SideStart, style.TwoThirds, style.MotionNone), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(s, "transform: translateX(-100%);") {
		t.Errorf("expected Drawer(SideStart) to park at translateX(-100%%):\n%s", s)
	}
	if strings.Contains(s, "transition: transform") {
		t.Errorf("MotionNone Drawer must emit no transition:\n%s", s)
	}
	if strings.Contains(s, "display: none;") {
		t.Errorf("even a MotionNone Drawer parks on a transform, not display:none:\n%s", s)
	}
}

func TestHideSwitchesAPartOffForOneDevice(t *testing.T) {
	// OnlyOn hides by default and reveals per device. Hide() is the other
	// direction: keep the base styling, switch the part off on one device.
	w := testWidget{name: "pd", kind: widget.Region}
	s := style.For(w).
		Part("header", style.Row(style.Space2), style.As(style.Panel)).
		On(css.Mobile, "header", style.Hide()).
		Stylesheet().String()

	if !strings.Contains(s, "@media (max-width: 639.98px)") {
		t.Errorf("expected the rule to be device-scoped, got:\n%s", s)
	}
	if !strings.Contains(s, "display: none;") {
		t.Errorf("expected Hide to emit display:none, got:\n%s", s)
	}
	if !strings.Contains(s, "background-color") {
		t.Errorf("the base styling must survive, got:\n%s", s)
	}
}

func TestOnEmitsDeviceMedia(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Part("menu", style.Row(style.Space1)).
		On(css.Mobile, "menu", style.Stack(style.Space2)).
		Stylesheet().String()

	if !strings.Contains(s, "@media (max-width: 639.98px)") {
		t.Errorf("expected On(css.Mobile) to emit @media (max-width: 639.98px):\n%s", s)
	}
}

func TestOnlyOnHidesByDefault(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		OnlyOn(css.Mobile, "hamburger", style.Row(style.Space1)).
		Stylesheet().String()

	if !strings.Contains(s, ".w__hamburger {\n  display: none;\n}") {
		t.Errorf("expected OnlyOn part to emit display: none in widgets layer:\n%s", s)
	}
	if !strings.Contains(s, "@media (max-width: 639.98px)") {
		t.Errorf("expected OnlyOn to emit media block:\n%s", s)
	}
}

func TestOnlyOnPartPassesValidate(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	sheet := style.For(w).
		OnlyOn(css.Mobile, "hamburger", style.Row(style.Space1))

	if errs := sheet.Validate(); len(errs) > 0 {
		t.Errorf("expected OnlyOn part to pass Validate, got: %v", errs)
	}
}

func TestOnRevealedByStaysInsideMedia(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure} // Disclosure allows Open
	s := style.For(w).
		Part("panel", style.RevealedBy(widget.Open)).
		On(css.Mobile, "panel", style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(s, "[data-open=\"true\"]") {
		t.Errorf("expected state rule for Open to be emitted:\n%s", s)
	}
}

func TestNoLayerStatementInsideMedia(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Part("menu", style.Row(style.Space1)).
		On(css.Mobile, "menu", style.Stack(style.Space2)).
		Stylesheet().String()

	idxMedia := strings.Index(s, "@media")
	if idxMedia >= 0 {
		after := s[idxMedia:]
		// Ensure @layer tokens, primitives, widgets, states does NOT appear after @media
		if strings.Contains(after, "@layer tokens,") {
			t.Errorf("found @layer statement inside @media block:\n%s", after)
		}
	}
}

func TestValidateDrawerWithoutRevealedBy(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	sheet := style.For(w).
		Part("panel", style.Drawer(style.SideEnd, style.TwoThirds, style.MotionNone))

	errs := sheet.Validate()
	if len(errs) == 0 {
		t.Fatal("expected Validate to reject Drawer without RevealedBy")
	}
	if !strings.Contains(errs[0].Error(), "Drawer() without RevealedBy()") {
		t.Errorf("unexpected error message: %v", errs[0])
	}
}

// TestPrimarySurface_GradientHookIsInertByDefaultAndLiveWhenSet is the
// consumer-shaped proof for css.SetGradient (webtyp/css) and the
// background-image line it depends on (style/emit_decls.go): a Primary
// surface always emits the hook, it costs nothing when no app opts in, and
// when an app's own Theme() call sets it, the SAME rule's output carries it
// through — the two packages actually compose, not just each in isolation.

// TestFloatMiddlePinsToTheViewportMiddleEdge is the emission proof for a
// vertically-centered floating control: fixed, top at half the viewport
// (SizeHalf, never a literal), pulled back by half its own height, off the
// inline edge by one Space step, at the widget kind's layer.
func TestFloatMiddlePinsToTheViewportMiddleEdge(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}
	s := style.For(w).
		Part("fab", style.FloatMiddle(style.SideEnd, style.Space4)).
		Stylesheet().String()

	for _, want := range []string{
		"position: fixed;",
		"top: 50%;",
		"transform: translateY(-50%);",
		"inset-inline-end: var(--space-4,1rem);",
		"inset-inline-start: auto;",
		"inset-block-end: auto;",
		"z-index: var(--z-dropdown,100);",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected FloatMiddle to emit %q, got:\n%s", want, s)
		}
	}

	m := style.For(w).
		Part("fab", style.FloatMiddle(style.SideStart, style.Space2)).
		Stylesheet().String()
	for _, want := range []string{
		"inset-inline-start: var(--space-2,0.5rem);",
		"inset-inline-end: auto;",
	} {
		if !strings.Contains(m, want) {
			t.Errorf("expected FloatMiddle(SideStart) to emit %q, got:\n%s", want, m)
		}
	}
}

func TestFloatMiddleRejectsTransformOwners(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}
	// Each sheet carries exactly one transform owner besides FloatMiddle;
	// the Drawer case pairs its mandatory RevealedBy so the combo rule is
	// the only possible error source.
	rotate := style.For(w).Part("fab", style.FloatMiddle(style.SideEnd, style.Space4), style.Rotate(style.TurnHalf))
	if errs := rotate.Validate(); len(errs) == 0 {
		t.Error("expected Validate to reject FloatMiddle + Rotate (both own transform)")
	}
	drawer := style.For(w).Part("fab",
		style.FloatMiddle(style.SideEnd, style.Space4),
		style.Drawer(style.SideEnd, style.TwoThirds, style.MotionNone),
		style.RevealedBy(widget.Open))
	if errs := drawer.Validate(); len(errs) == 0 {
		t.Error("expected Validate to reject FloatMiddle + Drawer (both own transform/position)")
	}
}
