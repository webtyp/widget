//go:build !wasm

package style_test

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
	"strings"
	"testing"
)

func TestCoverEmitsViewportFrame(t *testing.T) {
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).Root(style.Cover()).Stylesheet().String()

	// A definite height, not a floor: a min-height leaves the frame auto-sized,
	// so a Fill() descendant resolves to nothing and HideOverflow() never clips.
	if !strings.Contains(s, "height: 100dvh;") {
		t.Errorf("expected Cover to emit height: 100dvh, got:\n%s", s)
	}
	if strings.Contains(s, "min-height: 100dvh;") {
		t.Errorf("Cover must not fall back to min-height, got:\n%s", s)
	}
	if strings.Contains(s, "100vh") {
		t.Errorf("100vh (without d) must not appear in Cover output:\n%s", s)
	}
	if !strings.Contains(s, "@layer primitives") {
		t.Errorf("expected Cover to emit in @layer primitives:\n%s", s)
	}
}

func TestCappedGivesAScrollRegionSomewhereToOverflow(t *testing.T) {
	// Scroll() emits height:100%, min-height:0, flex-grow:1 and overflow-y:auto
	// — all four RELATIVE, so all four are inert until an ancestor has a
	// definite block size. Without Capped() an out-of-flow panel has none: the
	// list grows to its content and the page scrolls instead of the list.
	w := testWidget{name: "picker", kind: widget.Combobox}
	s := style.For(w).
		Part("panel", style.Stack(style.Space1), style.Flyout(style.SideStart), style.Capped(style.ExtentMost)).
		Part("options", style.Scroll()).
		Stylesheet().String()

	if !strings.Contains(s, "max-block-size: 70dvh;") {
		t.Errorf("expected Capped(ExtentMost) to emit max-block-size: 70dvh, got:\n%s", s)
	}
	// dvh, not vh: vh is frozen at the tallest viewport state, so a panel sized
	// in vh hangs below the fold whenever the phone's toolbar is showing.
	if strings.Contains(s, "max-block-size: 70vh;") {
		t.Errorf("Capped must not fall back to vh, got:\n%s", s)
	}
	if !strings.Contains(s, "overflow-y: auto;") {
		t.Errorf("expected the Scroll() part to keep its overflow, got:\n%s", s)
	}
}

func TestCappedStepsAreDistinct(t *testing.T) {
	w := testWidget{name: "picker", kind: widget.Combobox}
	seen := make(map[string]style.Extent)
	for _, e := range []style.Extent{style.ExtentHalf, style.ExtentMost, style.ExtentFull} {
		s := style.For(w).Part("panel", style.Capped(e)).Stylesheet().String()
		if prev, dup := seen[s]; dup {
			t.Errorf("Extent %d and %d resolve to the same cap", prev, e)
		}
		seen[s] = e
	}
}

func TestDividersEmitOnABasePartRule(t *testing.T) {
	// primitiveDecls() is reached from the device, state, hover and across
	// paths — but the base-rule path groups its own primitives, and the two
	// drifted: Divider()/DividerBelow() on a plain Part() validated, compiled
	// and emitted no CSS at all. The only consumer in the ecosystem happened to
	// declare it inside On(css.Mobile, …), so nothing caught it.
	w := testWidget{name: "list", kind: widget.Region}
	s := style.For(w).
		Part("row", style.DividerBelow()).
		Part("rail", style.Divider(style.SideEnd)).
		Part("aside", style.Divider(style.SideStart)).
		Stylesheet().String()

	for _, want := range []string{
		"border-block-end: 1px solid",
		"border-inline-end: 1px solid",
		"border-inline-start: 1px solid",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected a base Part rule to emit %q, got:\n%s", want, s)
		}
	}
}

func TestDividerBetweenSkipsTheFirstChild(t *testing.T) {
	// A separator belongs to the pair of rows it comes between. DividerBelow on
	// each row gives N rules for N rows, and the last one has nothing after it
	// to separate — the list reads as cut off rather than ended.
	w := testWidget{name: "list", kind: widget.Region}
	s := style.For(w).
		Part("options", style.Stack(style.SpaceNone), style.DividerBetween()).
		Stylesheet().String()

	if !strings.Contains(s, ".list__options > * + *") {
		t.Errorf("expected DividerBetween to emit a `> * + *` child rule, got:\n%s", s)
	}
	if !strings.Contains(s, "border-block-start: 1px solid") {
		t.Errorf("expected the between-rule to draw on the block-start edge, got:\n%s", s)
	}
	// Drawing below would need "every child that has a FOLLOWING sibling",
	// which has no selector — the rule has to hang off the start edge.
	if strings.Contains(s, ".list__options > * + * {\n  border-block-end") {
		t.Errorf("DividerBetween must not draw on the block-end edge, got:\n%s", s)
	}
}

func TestDividerBetweenIsRejectedWhereItCannotEmit(t *testing.T) {
	w := testWidget{name: "list", kind: widget.Region}
	errs := style.For(w).
		Part("options", style.Stack(style.SpaceNone)).
		On(css.Mobile, "options", style.DividerBetween()).
		Validate()

	if len(errs) == 0 {
		t.Fatal("expected DividerBetween() inside On() to be rejected: the device path emits a flat declaration list and cannot write a child combinator, so it would silently emit nothing")
	}
	if !strings.Contains(errs[0].Error(), "DividerBetween()") {
		t.Errorf("expected the error to name the offending option, got: %v", errs[0])
	}
}

func TestIconBoxEmitsSquareThatCannotShrink(t *testing.T) {
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).
		Part("icon-sm", style.IconBox(style.IconSm)).
		Part("icon-md", style.IconBox(style.IconMd)).
		Part("icon-lg", style.IconBox(style.IconLg)).
		Stylesheet().String()

	for _, want := range []string{
		"width: 1em;", "height: 1em;",
		"width: 1.5em;", "height: 1.5em;",
		"width: 2.5em;", "height: 2.5em;",
		"flex-shrink: 0;",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected IconBox to emit %q, got:\n%s", want, s)
		}
	}
}

func TestGrowClaimsWidthWithoutHeight(t *testing.T) {
	// Fill() also emits height: 100%, which inside a Row resolves against the
	// row and stretches the part into a full-height block. Grow() must not.
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).Part("label", style.Grow()).Stylesheet().String()

	if !strings.Contains(s, "flex-grow: 1;") {
		t.Errorf("expected Grow to emit flex-grow: 1, got:\n%s", s)
	}
	if !strings.Contains(s, "min-width: 0;") {
		t.Errorf("expected Grow to emit min-width: 0, got:\n%s", s)
	}
	if strings.Contains(s, "height: 100%;") {
		t.Errorf("Grow must not claim height, got:\n%s", s)
	}
}

func TestStackGapSurvivesAFlowChild(t *testing.T) {
	// The separation must live on the container. A `> * + *` rule reading
	// var(--gap) resolves it against the CHILD, so a child that is itself a
	// flow container declares its own --gap and silently replaces the parent's
	// spacing — with SpaceNone it collapses it to nothing.
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).
		Part("aside", style.Stack(style.Space1)).
		Part("aside-content", style.Stack(style.SpaceNone)).
		Stylesheet().String()

	if !strings.Contains(s, "gap: var(--gap);") {
		t.Errorf("expected Stack to put the gap on the container, got:\n%s", s)
	}
	if strings.Contains(s, "> * + *") {
		t.Errorf("Stack must not space its children with a child-resolved var(--gap), got:\n%s", s)
	}
}

func TestFlyoutHangsUnderItsAnchor(t *testing.T) {
	// A dropdown left in the flow pushes everything below it down, so the list
	// jumps under the pointer that opened it.
	w := testWidget{name: "shell", kind: widget.Menu}
	s := style.For(w).
		Part("menu", style.Anchor()).
		Part("options", style.Flyout(style.SideEnd)).
		Stylesheet().String()

	for _, want := range []string{
		"position: relative;",
		"position: absolute;",
		"inset-block-start: 100%;",
		"inset-inline-end: 0;",
		"z-index: var(--z-dropdown,100);",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Anchor/Flyout to emit %q, got:\n%s", want, s)
		}
	}
}

func TestDeviceRuleKeepsItsPrimitives(t *testing.T) {
	// The layout flags are grouped into shared selectors on the main path. A
	// device rule has nothing to group with, so they have to be emitted on its
	// own selector or an option passed to On()/OnlyOn() vanishes.
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).
		OnlyOn(css.Mobile, "trigger", style.Row(style.Space1), style.PushEnd(), style.KeepSize()).
		Stylesheet().String()

	for _, want := range []string{"margin-inline-start: auto;", "flex-shrink: 0;"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected the device rule to keep %q, got:\n%s", want, s)
		}
	}
}

func TestMasterDetailRestsOnTheMaster(t *testing.T) {
	// The master has to be what shows on arrival, and it has to get there
	// without a scroll nudge: the component contract has no mount hook. RTL puts
	// the strip's start edge on the right, which is where scroll position 0
	// already is, and order:1 sends the master — second in the DOM — there.
	w := testWidget{name: "cv", kind: widget.Region}
	s := style.For(w).
		On(css.Mobile, "", style.MasterDetail(style.Most)).
		Stylesheet().String()

	for _, want := range []string{
		"direction: rtl;",
		"flex-wrap: nowrap;",
		"flex: 0 0 auto;",
		"scroll-snap-type: x mandatory;",
		".cv > :nth-child(2)",
		"order: 1;",
		"scroll-snap-align: start;",
		".cv > :nth-child(1)",
		"order: 2;",
		"scroll-snap-align: end;",
		"flex: 0 0 90%;",
		"direction: ltr;",
		".cv > *",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected MasterDetail to emit %q, got:\n%s", want, s)
		}
	}
	if !strings.Contains(s, "@media") {
		t.Errorf("expected the rule to stay inside its device query, got:\n%s", s)
	}
}

func TestPushEndMovesFreeSpaceInFront(t *testing.T) {
	// Once flex-wrap drops an item onto a line of its own, nothing else on that
	// line can push it: the free space has to go in front of it.
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).Part("badge", style.PushEnd()).Stylesheet().String()

	if !strings.Contains(s, "margin-inline-start: auto;") {
		t.Errorf("expected PushEnd to emit margin-inline-start: auto, got:\n%s", s)
	}
}

func TestIconBoxStepsAreDistinct(t *testing.T) {
	w := testWidget{name: "shell", kind: widget.Region}
	seen := make(map[string]style.IconSize)
	for _, sz := range []style.IconSize{style.IconSm, style.IconMd, style.IconLg} {
		s := style.For(w).Part("icon", style.IconBox(sz)).Stylesheet().String()
		if prev, dup := seen[s]; dup {
			t.Errorf("IconSize %d and %d resolve to the same box", prev, sz)
		}
		seen[s] = sz
	}
}

func TestSidebarEndPutsRailLast(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Root(style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone)).Stylesheet().String()

	if !strings.Contains(s, "> :last-child") {
		t.Errorf("expected Sidebar(SideEnd) to emit > :last-child rail rule:\n%s", s)
	}
	if !strings.Contains(s, "flex-grow: 999;") {
		t.Errorf("expected Sidebar to emit flex-grow: 999 for content:\n%s", s)
	}
}

func TestSidebarStartMirrors(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Root(style.Sidebar(style.SideStart, style.RailNarrow, style.SpaceNone)).Stylesheet().String()

	if !strings.Contains(s, "> :first-child") {
		t.Errorf("expected Sidebar(SideStart) to emit > :first-child as rail:\n%s", s)
	}
}

func TestSidebarRailUsesToken(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Root(style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone)).Stylesheet().String()

	expected := css.RailNarrow.Var()
	if !strings.Contains(s, "--rail: "+expected) {
		t.Errorf("expected --rail value to be %q, got:\n%s", expected, s)
	}
}

func TestEmissionDeterministic(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	sheet := style.For(w).
		Root(style.Cover()).
		Part("sidebar", style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone)).
		Part("drawer", style.Drawer(style.SideEnd, style.TwoThirds, style.MotionSlow), style.RevealedBy(widget.Open)).
		On(css.Mobile, "drawer", style.Stack(style.Space2))

	css1 := sheet.Stylesheet().String()
	css2 := sheet.Stylesheet().String()
	if css1 != css2 {
		t.Errorf("Stylesheet emission is non-deterministic")
	}
}
