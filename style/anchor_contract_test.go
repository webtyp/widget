//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// The two shapes below are the real ones shipped in webtyp/components today.
// One works and one is silently broken, and NOTHING in the DSL distinguishes
// them — that is the defect these tests pin down.
//
//	usermenu   (works)    Root Anchor()  >  PartPanel  Flyout(SideEnd)
//	targetlist (broken)   PartRow Anchor()  >  PartMenu Docked(Parent,…)  >  PartOptions Flyout(SideStart)
//
// Flyout emits `position: absolute; inset-block-start: 100%`, and CSS resolves
// that 100% against the nearest POSITIONED ancestor. In usermenu that ancestor
// is the Anchor, exactly as documented. In targetlist the Docked <details> in
// between is also `position: absolute`, so IT becomes the containing block: the
// dropdown hangs from the 24px trigger rather than the 50px row, and paints
// over the row's own label. Measured in the browser at 1440x900 before these
// tests existed: options.top = 142.0 against row.bottom = 163.2 — 21.2px inside
// its own row — with options.offsetParent === .targetlist__menu.
//
// Docked()'s doc says "Parent pins it to the corner of the nearest Anchor" and
// Flyout()'s says "the nearest Anchor() ancestor is what it hangs from". Both
// sentences are false whenever anything between them is positioned, which the
// author cannot see from the call site. The consumer already carries the
// memorised workaround in a code comment ("No Anchor() here — Docked already
// makes this a containing block, and the two fight over position"), and a rule
// the author must remember is a hole in the harness, not a consumer problem.

// TestAnchorAndFlyoutWithNothingBetween is the CONTROL, and it is green today:
// usermenu's shape must keep working. Any diagnostic added for the broken shape
// that also fires here has over-reached.
func TestAnchorAndFlyoutWithNothingBetween(t *testing.T) {
	w := testWidget{name: "usermenushape", kind: widget.Menu}

	s := style.For(w).
		Root(style.Anchor()).
		Part("trigger", style.Row(style.Space2), style.Pad(style.Space1)).
		Part("panel", style.Stack(style.SpaceNone), style.Flyout(style.SideEnd))

	if errs := s.Validate(); len(errs) > 0 {
		t.Fatalf("Anchor > Flyout with nothing between is the shape the DSL "+
			"documents and must stay valid, got: %v", errs)
	}

	out := s.Stylesheet().String()
	if !strings.Contains(out, "position: relative;") {
		t.Error("the Anchor must emit position: relative")
	}
	if !strings.Contains(out, "inset-block-start: 100%;") {
		t.Error("the Flyout must hang from its containing block's bottom")
	}
}

// TestDockedPartBetweenAnchorAndFlyoutIsDiagnosed is RED until widget/docs/PLAN.md
// lands. It reproduces targetlist's composition through the public API exactly
// as a consumer writes it, and asserts the library says something about it.
//
// Today Validate() returns nothing at all. checkPosition already catches
// Anchor+Docked on ONE rule — the easy version of this mistake — but parts are
// held in a flat map with no record of which part renders inside which, so the
// across-parts version, the one that actually shipped broken, is structurally
// invisible to it. Closing this needs the sheet to know the part tree; the plan
// covers what that costs.
//
// The assertion is on the diagnostic, not on emitted CSS, on purpose: whichever
// way the plan resolves the composition (a spanning dock, a restructured DOM, a
// Flyout that names its anchor), the thing that must never happen again is a
// consumer writing this and getting silence.
func TestDockedPartBetweenAnchorAndFlyoutIsDiagnosed(t *testing.T) {
	w := testWidget{name: "targetlistshape", kind: widget.Listbox}

	s := style.For(w).
		Part("row", style.Anchor(), style.Row(style.Space2), style.Pad(style.Space3)).
		Part("menu", style.KeepSize(),
			style.Docked(style.Parent, style.EdgeTop, style.SideStart, style.Space1)).
		Part("button", style.Round(style.RadiusSm), style.Pad(style.Space1)).
		Part("options", style.Stack(style.SpaceNone), style.Flyout(style.SideStart))

	errs := s.Validate()
	if len(errs) == 0 {
		t.Fatal("a sheet that declares an Anchor, a Docked(Parent, …) part and a " +
			"Flyout part gives the author no way to say which contains which, and " +
			"no way to see that the Docked part will steal the Flyout's containing " +
			"block. Validate() must not stay silent about it: either the sheet " +
			"learns the part tree and rejects the theft precisely, or this " +
			"composition is reported as ambiguous and the author is made to " +
			"declare the nesting. See widget/docs/PLAN.md.")
	}

	var joined string
	for _, e := range errs {
		joined += e.Error() + "\n"
	}
	for _, want := range []string{"options", "menu"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the diagnostic must name the parts involved; %q missing from:\n%s",
				want, joined)
		}
	}
}
