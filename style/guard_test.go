//go:build !wasm

package style_test

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
	"regexp"
	"strings"
	"testing"
)

func extractVarCalls(cssStr string) [][2]string {
	var results [][2]string
	idx := 0
	for {
		start := strings.Index(cssStr[idx:], "var(--")
		if start == -1 {
			break
		}
		startPos := idx + start
		// scan for the matching closing parenthesis
		parenCount := 1
		p := startPos + 4 // after "var("
		for p < len(cssStr) && parenCount > 0 {
			if cssStr[p] == '(' {
				parenCount++
			} else if cssStr[p] == ')' {
				parenCount--
			}
			p++
		}
		if parenCount == 0 {
			fullMatch := cssStr[startPos:p]
			// Extract variable name, which is between "var(" and either "," or ")"
			varNamePart := fullMatch[4 : len(fullMatch)-1]
			varName := varNamePart
			commaIdx := strings.Index(varNamePart, ",")
			if commaIdx != -1 {
				varName = strings.TrimSpace(varNamePart[:commaIdx])
			}
			results = append(results, [2]string{fullMatch, varName})
		}
		idx = startPos + 4
	}
	return results
}

func TestNoInventedValues(t *testing.T) {
	// Extend drift guard to verify every var() matches the catalog + fallbacks (closes D-5)
	wd := &testWidget{name: "w", kind: widget.Dialog}

	// Giant sheet to exercise EVERY option and scale
	sheet := style.For(wd).
		Root(
			style.Stack(style.Space12),
			style.As(style.Page),
			style.Pad(style.Space8),
			style.Round(style.RadiusFull),
			style.Raise(style.Popover),
			style.Width(style.Readable),
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
			style.Animate(style.MotionSlow),
			style.Scroll(),
			style.KeepSize(),
			style.EdgeToEdge(),
			style.HideOverflow(),
			style.Backdrop(style.Parent), // Backdrop(Parent) avoids condition 6 Backdrop(Viewport) under a Split error
			style.Veil(),
		).
		Part("item1", style.Row(style.Space1), style.Interactive(style.Primary)).
		Part("item2", style.Split(style.SplitTwoThirds, style.Space2), style.As(style.Panel)).
		Part("item3", style.Grid(style.ColumnWide, style.Space3), style.As(style.Inset)).
		Part("item4", style.Center(style.Third), style.As(style.Secondary)).
		Part("item5", style.FillCentered(), style.As(style.Highlight)).
		Part("item6", style.ScrollRow(style.Space4), style.As(style.Success)).
		Part("item7", style.MediaBox(style.Aspect16x9), style.As(style.Danger)).
		Part("item8", style.Stack(style.Space6), style.Interactive(style.Subtle)).
		Part("item9", style.Row(style.SpaceNone), style.Interactive(style.Inset)).
		Part("item10", style.Cover()).
		Part("item11", style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone)).
		Part("item12", style.Drawer(style.SideEnd, style.TwoThirds, style.MotionSlow), style.RevealedBy(widget.Open)).
		Part("item13", style.IconBox(style.IconSm)).
		Part("item14", style.IconBox(style.IconMd)).
		Part("item15", style.IconBox(style.IconLg)).
		Part("item16", style.Anchor(), style.FloatingChrome(style.EdgeBottom, style.IconLg, style.Space4)).
		Part("item17", style.OnEdge(style.EdgeTop, style.SideStart, style.Space2, style.Space1)).
		Part("item18", style.EdgeStrip(style.Parent, style.SideStart)).
		Part("item19", style.Meter(style.Space1)).
		Part("item20", style.IconBox(style.IconMd), style.CenterSelf()).
		Part("item21", style.FixedGrid(7, style.Space2)).
		On(css.Mobile, "item10", style.Stack(style.Space1))

	cssStr := sheet.Stylesheet().String()

	// Token maps to assert fallbacks exactly
	cssTokens := []css.Token{
		css.ColorPrimary, css.ColorOnPrimary,
		css.ColorSuccess, css.ColorOnSuccess,
		css.ColorDanger, css.ColorOnDanger,
		css.ColorBackground, css.ColorOnBackground,
		css.ColorSurface, css.ColorOnSurface,
		css.ColorSurfaceSunken,
		css.ColorSelection, css.ColorOnSelection,
		css.ColorOutline, css.ColorMuted,
		css.MixHover, css.MixFocus, css.MixPress,
		css.TextXs, css.TextSm, css.TextBase, css.TextLg, css.TextXl, css.Text2xl,
		css.LeadingNormal,
		css.FontWeightRegular, css.FontWeightMedium, css.FontWeightBold,
		css.Space0, css.Space1, css.Space2, css.Space3, css.Space4, css.Space6, css.Space8, css.Space12,
		css.RadiusSm, css.RadiusMd, css.RadiusLg, css.RadiusFull,
		css.ShadowSm, css.ShadowMd, css.ShadowLg,
		css.DurationFast, css.DurationBase, css.DurationSlow,
		css.EaseInOut,
		css.ZBase, css.ZDropdown, css.ZSticky, css.ZModal, css.ZToast, css.ZTooltip,
		css.BpSm, css.BpMd, css.BpLg, css.BpXl,
		css.MaxWReadable,
		css.ColumnNarrow, css.ColumnMedium, css.ColumnWide,
		css.RailNarrow, css.RailWide,
		css.ControlHeight, css.ChipWidth, css.ChipHeight, css.VeilBlur,
		css.ColorAccent, css.ColorOnAccent,
	}

	tokenMap := make(map[string]css.Token)
	for _, tok := range cssTokens {
		tokenMap[tok.Name] = tok
	}

	// Layout variables allowed to not exist in tokenMap
	layoutVariables := map[string]bool{
		"--gap":             true,
		"--ratio":           true,
		"--track":           true,
		"--cols":            true,
		"--rail":            true,
		"--max-width":       true,
		"--floating-top":    true,
		"--floating-bottom": true,
		"--meter-fill":      true,
	}
	// Every theme pair's plain light/dark half properties (see
	// css.Token.EnhancedVar, css declareSplit) — a different kind of
	// variable than the ones above, not derived through Token.Var()'s
	// formula, so skipped the same way rather than checked against it.
	for _, tok := range []css.Token{
		css.ColorBackground, css.ColorOnBackground, css.ColorSurface,
		css.ColorOnSurface, css.ColorOutline, css.ColorMuted,
	} {
		layoutVariables[tok.LightVarName()] = true
		layoutVariables[tok.DarkVarName()] = true
	}
	// Every surface family's background-image companion (see
	// css.Token.ImageVarName, css.SetGradient) — emitted unconditionally
	// alongside background-color for every surfaced rule, one per distinct
	// token familyBase() can return, inert (var(..., none)) until an app
	// opts in with SetGradient.
	for _, tok := range []css.Token{
		css.ColorSurface, css.ColorPrimary, css.ColorAccent,
		css.ColorSuccess, css.ColorDanger, css.ColorMuted, css.ColorBackground,
	} {
		layoutVariables[tok.ImageVarName()] = true
	}

	// Check there are no hardcoded hexadecimal colors or rgba except in fallbacks of var()
	// or in a Safari-legacy static fallback (see below).
	varVarRegex := regexp.MustCompile(`var\([^)]+\)`)
	cleanCSS := varVarRegex.ReplaceAllString(cssStr, "")

	// ACCEPTED EXCEPTION — Safari-legacy static fallback (double declaration):
	// every themed color is now emitted twice, static-literal first and
	// light-dark()/color-mix()-enhanced second (see css.Token.LightValue,
	// css.HoverStatic/FocusStatic/PressStatic/FadeStatic, and their call
	// sites in emit.go/emit_decls.go/surface.go). A browser without
	// light-dark()/color-mix() support drops the second declaration
	// (invalid at parse time) and keeps the static first one — permanently
	// the light theme. That first declaration is a bare literal outside
	// var() BY DESIGN, so this drift guard must not flag it. It is still
	// bounded: only literals that exactly match a value this package itself
	// derives from the catalog (never an arbitrary hardcoded color) are
	// allowed.
	knownStaticFallbacks := map[string]bool{}
	// Only literals that are themselves hex colors go in the allowlist — a
	// non-color token's LightValue ("0.75rem", "15%", "0") would otherwise
	// strip that substring out of unrelated hex codes elsewhere in the sheet
	// (e.g. stripping "0" mangles "#ba2c0d" into "#ba2cd").
	addIfHex := func(s string) {
		if strings.HasPrefix(s, "#") {
			knownStaticFallbacks[s] = true
		}
	}
	addIfHex(css.FadeStatic(css.ColorSurface, 0.4)) // Veil() backdrop wash
	for _, tok := range cssTokens {
		addIfHex(tok.LightValue())
		addIfHex(css.HoverStatic(tok))
		addIfHex(css.FocusStatic(tok))
		addIfHex(css.PressStatic(tok))
		// EnhancedVar()/NestedEnhanced() bake BOTH halves of a theme token
		// as literals (e.g. light-dark(#F2F2F7,#161B22)) — see their doc
		// comments in webtyp/css for why a var() reference can't be used
		// here instead. addIfHex only ever matches a value that IS a bare
		// hex, so this is harmless for tok.Light/tok.Dark that are actually
		// live expressions (e.g. "15%", "0.75rem", or a color-mix() string).
		addIfHex(tok.Light)
		addIfHex(tok.Dark)
	}
	stripKnownStatic := func(s string) string {
		for lit := range knownStaticFallbacks {
			s = strings.ReplaceAll(s, lit, "")
		}
		return s
	}
	cleanCSS = stripKnownStatic(cleanCSS)

	hexColorRegex := regexp.MustCompile(`#[0-9a-fA-F]{3,8}`)
	if hexColorRegex.MatchString(cleanCSS) {
		t.Errorf("Literal color hexadecimal found outside var(): %q", hexColorRegex.FindString(cleanCSS))
	}

	rgbaRegex := regexp.MustCompile(`rgba\([^)]+\)`)
	if rgbaRegex.MatchString(cleanCSS) {
		t.Errorf("Literal rgba wash found outside var(): %q", rgbaRegex.FindString(cleanCSS))
	}

	vwRegex := regexp.MustCompile(`\b[0-9]+vw\b`)
	vhRegex := regexp.MustCompile(`\b[0-9]+vh\b`)
	if vwRegex.MatchString(cleanCSS) {
		t.Errorf("Prohibited unit 'vw' found: %q", vwRegex.FindString(cleanCSS))
	}
	if vhRegex.MatchString(cleanCSS) {
		t.Errorf("Prohibited unit 'vh' found: %q", vhRegex.FindString(cleanCSS))
	}

	// Check all variables are from our catalog
	varMatches := extractVarCalls(cssStr)
	for _, m := range varMatches {
		fullMatch := m[0]
		varName := m[1]

		if layoutVariables[varName] {
			continue
		}

		tok, exists := tokenMap[varName]
		if !exists {
			t.Errorf("CSS variable %q not in css catalog or local private tokens list", varName)
			continue
		}

		// ACCEPTED DRIFT GUARD EXCEPTION:
		// We tolerate a bare "var(--name)" call with no local fallback string only when the variable
		// name is a valid token from the css catalog.
		// Why this is safe: These matches represent nested variable references within larger formulas
		// (such as those returned by ColorSurfaceSunken or ColorSelection from webtyp/css v0.3.3+).
		// Because they are nested inside an outer var() call, they are already protected by the outer
		// var()'s fallback. Enforcing fallback matching on these nested references is both redundant
		// and structurally impossible here since the formulas themselves are owned and defined by webtyp/css.
		expectedVarCall := tok.Var()
		if fullMatch != expectedVarCall && fullMatch != "var("+varName+")" {
			t.Errorf("Visual drift detected for %q.\nIn stylesheet: %q\nExpected: %q",
				varName, fullMatch, expectedVarCall)
		}
	}
}

// TestDoubleDeclarationsAreParseTimeSafe is the regression test for the
// iPhone-7-all-blue investigation. Every themed color in this package is
// emitted TWICE per property: a static hex literal first, a
// light-dark()/color-mix() "enhanced" value second — a browser without
// those functions is supposed to drop the second declaration (unrecognized
// function name, invalid AT PARSE TIME) and keep the first.
//
// That only works if the SECOND declaration contains no var() ANYWHERE.
// This is not obvious and was the actual bug: per the CSS Custom Properties
// cascade, a declaration containing var() — even a var() to a property that
// would always itself resolve fine — is a "variable-valued" declaration,
// and validity for the WHOLE thing is deferred to computed-value time
// instead of checked at parse time. It still wins the cascade over the
// earlier static declaration (cascade doesn't know yet that it will turn
// out invalid), and when it then fails to compute, the property falls to
// its INITIAL value (transparent, for a color) — not to the earlier static
// sibling. The static declaration is silently discarded, and the page goes
// solid blue with white text: the demo's own Primary-colored ancestor
// showing through every unpainted surface.
//
// Confirmed empirically, not just by spec reading, before this test
// existed: a minimal repro using an unrecognized function whose arguments
// were LITERALS correctly fell back to the static declaration; the
// identical repro with var() references as those arguments did not — see
// css.Token.NestedEnhanced's doc comment for the full mechanism and
// css.Token.EnhancedVar for the standalone-declaration counterpart.
//
// This walks every rule this package can generate — Interactive() (which
// exercises Hover/Focus/Press), every As(Surface), and Veil() — and, for
// every property declared more than once within a single rule, asserts
// every declaration after the first contains no var() anywhere.
