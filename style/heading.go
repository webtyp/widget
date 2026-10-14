//go:build !wasm

package style

// Heading is the one recipe for a section title inside a screen's content —
// an accordion step, a card header, a dialog title, a closing panel. The
// screen title itself belongs to the platform header, not to this recipe.
// Size and weight are the design system's decision, never the component's.
func Heading() Option {
	return func(r *rule) {
		FontSize(TextLg)(r)
		FontWeight(WeightBold)(r)
	}
}
