//go:build !wasm

package style_test

import (
	"regexp"
	"strings"
	"testing"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func TestDoubleDeclarationsAreParseTimeSafe(t *testing.T) {
	wd := &testWidget{name: "w", kind: widget.Dialog}
	sheet := style.For(wd).
		Root(style.As(style.Page), style.Backdrop(style.Parent), style.Veil()).
		Part("item1", style.Row(style.Space1), style.Interactive(style.Primary)).
		Part("item2", style.As(style.Panel)).
		Part("item3", style.As(style.Inset)).
		Part("item4", style.As(style.Secondary)).
		Part("item5", style.As(style.Highlight)).
		Part("item6", style.As(style.Accent)).
		Part("item7", style.As(style.Success)).
		Part("item8", style.As(style.Danger)).
		Part("item9", style.Interactive(style.Subtle)).
		Part("item10", style.Interactive(style.Inset))

	cssStr := sheet.Stylesheet().String()

	// Each {...} match is one rule's declaration body — [^{}]* can't cross
	// into a nested block, so this naturally lands on the innermost
	// (non-@layer-wrapper) blocks, exactly the plain declaration lists this
	// test needs to inspect.
	blockRe := regexp.MustCompile(`\{[^{}]*\}`)
	for _, block := range blockRe.FindAllString(cssStr, -1) {
		body := strings.TrimSuffix(strings.TrimPrefix(block, "{"), "}")
		seen := map[string]int{}
		for _, rawDecl := range strings.Split(body, ";") {
			decl := strings.TrimSpace(rawDecl)
			if decl == "" {
				continue
			}
			parts := strings.SplitN(decl, ":", 2)
			if len(parts) != 2 {
				continue
			}
			prop := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			seen[prop]++
			// Only the combination is broken: a plain var() with nothing
			// risky in it (e.g. "var(--color-primary,#1b5d8c)", where
			// --color-primary is never wrapped in light-dark()/color-mix())
			// is completely safe on any browser — var() itself needs no
			// protecting. light-dark()/color-mix() WITH a nested var() is
			// the failure mode this whole test exists to catch.
			risky := strings.Contains(val, "light-dark(") || strings.Contains(val, "color-mix(")
			if seen[prop] > 1 && risky && strings.Contains(val, "var(") {
				t.Errorf("declaration #%d of %q mixes a light-dark()/color-mix() call with var() — a browser that can't parse the outer function defers the WHOLE declaration to computed-value time and falls to the initial value instead of the earlier static declaration, discarding the Safari-legacy fallback: %q (rule: %q)",
					seen[prop], prop, val, block)
			}
		}
	}
}

func TestSurfaceCarriesShape(t *testing.T) {
	// As(Panel) alone emits radius; Round(RadiusNone) overrides it (closes D-6)
	wd := &testWidget{name: "w", kind: widget.Region}
	s1 := style.For(wd).Root(style.As(style.Panel)).Stylesheet().String()
	if !strings.Contains(s1, "border-radius: var(--radius-md") {
		t.Errorf("expected Panel alone to emit radius-md, got:\n%s", s1)
	}

	s2 := style.For(wd).Root(style.As(style.Panel), style.Round(style.RadiusNone)).Stylesheet().String()
	if strings.Contains(s2, "border-radius: var(--radius-md") {
		t.Errorf("expected Round(RadiusNone) to override Panel's default radius, got:\n%s", s2)
	}
}

func TestDerivedSurfaceAssertsNoBackgroundImage(t *testing.T) {
	// A family surface emits its `--x-image` companion so css.SetGradient can
	// reach it. A DERIVED surface (AccentWash/AccentInverse/AccentHover — no
	// family token) has no image of its own, but must still emit
	// `background-image: none` so a lower-layer family image cannot bleed
	// through: a nav item filled As(Primary) in @layer widgets and overridden
	// As(AccentInverse) in @layer states would otherwise keep the widgets-layer
	// `background-image: var(--color-primary-image, none)` — under SetGradient,
	// the gradient painting over the amber "current" fill.
	wd := &testWidget{name: "w", kind: widget.Region}

	for _, s := range []style.Surface{style.AccentWash, style.AccentInverse, style.AccentHover} {
		out := style.For(wd).Root(style.As(s)).Stylesheet().String()
		if !strings.Contains(out, "background-image: none;") {
			t.Errorf("As(%s) must emit `background-image: none;`, got:\n%s", s, out)
		}
		if strings.Contains(out, "background-image: var(") {
			t.Errorf("As(%s) is derived and must not emit a family background-image var, got:\n%s", s, out)
		}
	}

	fam := style.For(wd).Root(style.As(style.Primary)).Stylesheet().String()
	if !strings.Contains(fam, "background-image: var(--color-primary-image, none);") {
		t.Errorf("As(Primary) must still emit its family image companion, got:\n%s", fam)
	}
}
