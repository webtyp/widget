package style_test

import (
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// Glyph(Subtle) and Glyph(Inactive) are muted TEXT. Subtle's family base is the neutral surface (a background
// colour, for its hover states), so taking the glyph from the family base painted a RUT in the surface grey on a
// white sheet: invisible until hover. Inactive had no family base at all and emitted an empty colour.
func TestGlyph_MutedSurfacesPaintMutedText(t *testing.T) {
	for _, s := range []style.Surface{style.Subtle, style.Inactive} {
		w := testWidget{name: "g", kind: widget.Region}
		got := style.For(w).Part("sub", style.Glyph(s)).Stylesheet().String()
		if !strings.Contains(got, "color: "+css.ColorMuted.EnhancedVar()+";") {
			t.Errorf("Glyph(%v): want muted text colour, got:\n%s", s, got)
		}
		if strings.Contains(got, "color: "+css.ColorSurface.EnhancedVar()+";") || strings.Contains(got, "color: ;") {
			t.Errorf("Glyph(%v): text painted with a background colour or nothing:\n%s", s, got)
		}
	}
}
