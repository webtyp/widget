//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// A part whose only job is a primitive emits into @layer primitives, never into
// @layer widgets. Validate must consult both layers before calling it empty.
func TestPrimitiveOnlyPartIsNotEmpty(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	cases := []struct {
		name string
		opt  style.Option
		decl string
	}{
		{"Fill", style.Fill(), "flex-grow: 1;"},
		{"Scroll", style.Scroll(), "overflow-y: auto;"},
		{"KeepSize", style.KeepSize(), "flex-shrink: 0;"},
		{"EdgeToEdge", style.EdgeToEdge(), "margin: 0;"},
		{"HideOverflow", style.HideOverflow(), "overflow: hidden;"},
		{"FillCentered", style.FillCentered(), "place-items: center;"},
		{"Meter", style.Meter(style.Space1), "width: var(--meter-fill, 0%);"},
		{"CenterSelf", style.CenterSelf(), "margin-inline: auto;"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := style.For(w).Part("mount", c.opt)

			if errs := s.Validate(); len(errs) > 0 {
				t.Fatalf("%s alone must be a valid part, got: %v", c.name, errs)
			}

			out := s.Stylesheet().String()
			if !strings.Contains(out, c.decl) {
				t.Errorf("expected %q in emitted CSS, got:\n%s", c.decl, out)
			}
		})
	}
}

// The condition still catches the case it exists for: a part that declares nothing.
func TestPartWithoutOptionsEmitsNothing(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	errs := style.For(w).Part("mount").Validate()
	if len(errs) != 1 {
		t.Fatalf("expected exactly one error, got: %v", errs)
	}
	if got := errs[0].Error(); got != `sheet w: part "mount" emits nothing` {
		t.Errorf("unexpected message: %s", got)
	}
}

// Within() declares the part tree; both ends must exist and must differ.
func TestWithinChecksItsEnds(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}

	// A container that was never declared.
	errs := style.For(w).
		Part("panel", style.Flyout(style.SideEnd)).
		Within("ghost", "panel", style.Flyout(style.SideEnd)).
		Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), `Within() container "ghost" is not a declared part`) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the undeclared container to be diagnosed, got: %v", errs)
	}

	// A part declared inside itself.
	errs = style.For(w).Within("panel", "panel", style.Flyout(style.SideEnd)).Validate()
	found = false
	for _, e := range errs {
		if strings.Contains(e.Error(), "Within() declares the part inside itself") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the reflexive declaration to be diagnosed, got: %v", errs)
	}
}

// A containment cycle must not hang the chain walk.
func TestWithinCycleIsDiagnosed(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}

	s := style.For(w).
		Within("a", "b", style.Flyout(style.SideEnd)).
		Within("b", "a", style.Row(style.Space1))

	found := false
	for _, e := range s.Validate() {
		if strings.Contains(e.Error(), "containment cycle") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the containment cycle to be diagnosed, got: %v", s.Validate())
	}
}

// The full declaration of the broken shape — row(Anchor) > menu(Docked) >
// options(Flyout), nesting declared with Within at every level — is rejected
// with a precise error naming both the Flyout and the part that steals its
// containing block.
func TestDeclaredChainWithPositionedPartIsRejectedPrecisely(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Listbox}

	s := style.For(w).
		Within("menu", "options", style.Flyout(style.SideStart)).
		Within("row", "menu", style.Docked(style.Parent, style.EdgeTop, style.SideStart, style.Space1)).
		Part("row", style.Anchor(), style.Row(style.Space2)).
		Part("button", style.Round(style.RadiusSm))

	var joined string
	for _, e := range s.Validate() {
		joined += e.Error() + "\n"
	}
	for _, want := range []string{"options", "menu"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the diagnostic must name the parts involved; %q missing from:\n%s", want, joined)
		}
	}
	if !strings.Contains(joined, "positioned between the Flyout and its Anchor") {
		t.Errorf("expected the theft to be named, got:\n%s", joined)
	}
}

// A Flyout coexisting with a containing-block part, with NO Within declared,
// is ambiguous by construction and must be declared — this is the shape that
// shipped broken in targetlist.
func TestUncontainedFlyoutAlongsideContainingBlockIsAmbiguous(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Listbox}

	s := style.For(w).
		Part("row", style.Anchor(), style.Row(style.Space2)).
		Part("menu", style.Docked(style.Parent, style.EdgeTop, style.SideStart, style.Space1)).
		Part("options", style.Flyout(style.SideStart))

	var joined string
	for _, e := range s.Validate() {
		joined += e.Error() + "\n"
	}
	for _, want := range []string{"options", "menu", "Within()"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected the ambiguous composition to be diagnosed and the fix named; %q missing from:\n%s", want, joined)
		}
	}
}

// The other half of the Scroll/Flyout seam: a Flyout whose declared chain
// passes through a Scroll() region is clipped by it — the targetlist shape,
// list is a scroller and options is a flyout inside it, which measured 10px
// visible of 84.8. Validate must reject it instead of emitting CSS that shows
// a panel at half height.
func TestFlyoutInsideScrollRegionIsRejected(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Listbox}

	s := style.For(w).
		Part("list", style.Scroll()).
		Within("list", "options", style.Flyout(style.SideStart))

	errs := s.Validate()
	if len(errs) != 1 {
		t.Fatalf("expected exactly one error, got: %v", errs)
	}
	for _, want := range []string{"options", "list", "Scroll()"} {
		if !strings.Contains(errs[0].Error(), want) {
			t.Errorf("the diagnostic must name the parts involved; %q missing from:\n%s", want, errs[0])
		}
	}
}

// The clip is by DOM ancestry, so an Anchor below the scroller does not save
// the Flyout — the walk goes past it, to the end of the declared chain. But a
// scroller that is NOT on the Flyout's chain changes nothing and must not be
// reported: only the chain decides.
func TestFlyoutScrollChainWalkIsCompleteAndExact(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Listbox}

	// Anchor between the Flyout and the scroller: still clipped.
	s := style.For(w).
		Part("list", style.Scroll()).
		Within("list", "row", style.Row(style.Space2)).
		Part("row", style.Anchor()).
		Within("row", "options", style.Flyout(style.SideStart))

	found := false
	for _, e := range s.Validate() {
		if strings.Contains(e.Error(), "Scroll() region") {
			found = true
		}
	}
	if !found {
		t.Errorf("an Anchor below the scroller must not stop the chain walk, got: %v", s.Validate())
	}

	// A scroller on a different branch of the tree: not on the chain, ignored.
	if errs := style.For(w).
		Part("row", style.Anchor(), style.Row(style.Space2)).
		Within("row", "options", style.Flyout(style.SideStart)).
		Part("pane", style.Scroll()).
		Validate(); len(errs) > 0 {
		t.Errorf("a Scroll() part outside the Flyout's chain must not trip the check, got: %v", errs)
	}
}

// The resolutions the plan leaves open for the components stage must NOT
// trip the new rules: declaring the nesting silences the ambiguous error,
// and a chain that ends at the declared container without an Anchor above it
// is the author's explicit choice.
func TestDeclaredContainmentIsNotOverRejected(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Listbox}

	// The trigger in flow: Anchor above, nothing positioned between — the
	// documented shape, now with the nesting declared.
	if errs := style.For(w).
		Part("row", style.Anchor(), style.Row(style.Space2)).
		Within("menu", "options", style.Flyout(style.SideStart)).
		Part("menu", style.KeepSize()).
		Validate(); len(errs) > 0 {
		t.Errorf("Anchor > plain part > Flyout with declared nesting must stay valid, got: %v", errs)
	}

	// The docked trigger that spans its anchor: the declared container IS
	// the positioned terminal, no Anchor above it in the chain.
	if errs := style.For(w).
		Part("row", style.Anchor(), style.Row(style.Space2)).
		Within("menu", "options", style.Flyout(style.SideStart)).
		Part("menu", style.Docked(style.Parent, style.EdgeTop, style.SideStart, style.Space1)).
		Validate(); len(errs) > 0 {
		t.Errorf("a declared docked trigger (the plan's 2b shape) must stay valid, got: %v", errs)
	}

	// A docked part that is NOT between the Flyout and its Anchor changes
	// nothing: the sheet only reports what the declared chain shows.
	if errs := style.For(w).
		Part("row", style.Anchor(), style.Row(style.Space2)).
		Within("row", "options", style.Flyout(style.SideStart)).
		Part("fab", style.Docked(style.Parent, style.EdgeBottom, style.SideEnd, style.Space4)).
		Validate(); len(errs) > 0 {
		t.Errorf("an unrelated docked part must not trip the chain check, got: %v", errs)
	}
}
