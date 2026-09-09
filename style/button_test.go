//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

const partAdd = widget.Part("add")

// TestButton_DoesNotStretchInAStack is the regression guard for the defect that
// motivated the option: a button inside a Stack (flex column) was stretched to
// its panel's full width — 799px measured in the running demo — because
// align-items: stretch governs the cross axis and KeepSize()'s flex-shrink /
// flex-grow do not. Any align-self other than stretch takes content width.
func TestButton_DoesNotStretchInAStack(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Root(style.Stack(style.Space2)).
		Part(partAdd, style.Button(style.Primary)).
		Stylesheet().String()

	if !strings.Contains(s, "align-self: center;") {
		t.Errorf("Button must not stretch across a Stack's cross axis; no align-self emitted:\n%s", s)
	}
}

// TestButton_CarriesControlHeightAndInlinePadding: the box half of the recipe.
// The height is the shared rhythm a list row and a form field also stand on;
// the inline padding is what stopped "Remove row" overflowing its own box.
func TestButton_CarriesControlHeightAndInlinePadding(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Part(partAdd, style.Button(style.Primary)).
		Stylesheet().String()

	for _, want := range []string{
		"min-height: " + css.ControlHeight.Var() + ";",
		"padding-inline: " + css.Space3.Var() + ";",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("Button missing %q:\n%s", want, s)
		}
	}
}

// TestButton_PaintsLikeInteractive: Button's surface half is Interactive's, not
// a second paint path. The two rules are not byte-identical — only one carries
// the box — so this compares the surface-derived declarations.
func TestButton_PaintsLikeInteractive(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}
	btn := style.For(wd).Part(partAdd, style.Button(style.Primary)).Stylesheet().String()
	inter := style.For(wd).Part(partAdd, style.Interactive(style.Primary)).Stylesheet().String()

	for name, want := range map[string]string{
		"hover": css.Hover(css.ColorPrimary),
		"focus": css.Focus(css.ColorPrimary),
		"press": css.Press(css.ColorPrimary),
	} {
		if !strings.Contains(inter, want) {
			t.Fatalf("test premise broken: Interactive(Primary) emits no %s derivation", name)
		}
		if !strings.Contains(btn, want) {
			t.Errorf("Button(Primary) must derive %s like Interactive(Primary): missing %s", name, want)
		}
	}
}

// TestButton_RejectsRedundantComposition: both contradictions are loud. Saying
// Interactive beside Button silently overwrites the surface Button chose, and a
// second ControlBox repeats the same min-height token while reading as if it
// added something.
func TestButton_RejectsRedundantComposition(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Region}

	for _, tc := range []struct {
		name string
		opts []style.Option
		want string
	}{
		{
			name: "Button+Interactive",
			opts: []style.Option{style.Button(style.Primary), style.Interactive(style.Secondary)},
			want: "Button already paints an interactive surface; drop Interactive",
		},
		{
			name: "Button+ControlBox",
			opts: []style.Option{style.Button(style.Primary), style.ControlBox()},
			want: "Button already carries the control height; drop ControlBox",
		},
	} {
		errs := style.For(wd).Part(partAdd, tc.opts...).Validate()
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), tc.want) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: expected an error containing %q, got %v", tc.name, tc.want, errs)
		}
	}
}
