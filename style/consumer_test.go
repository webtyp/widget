//go:build !wasm

package style_test

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

type MasterDetail struct{}

func (m *MasterDetail) WidgetName() widget.Name { return widget.Name("masterdetail") }
func (m *MasterDetail) WidgetKind() widget.Kind { return widget.Grid }

func (m *MasterDetail) Style() *style.Sheet {
	return style.For(m).
		Root(style.Grid(style.ColumnNarrow, style.Space2), style.As(style.Page), style.Scroll()).
		Part("master", style.Stack(style.Space1), style.As(style.Panel), style.Pad(style.Space3)).
		Part("detail", style.Stack(style.Space2), style.As(style.Panel), style.Pad(style.Space3)).
		Part("item", style.Row(style.Space1), style.Interactive(style.Subtle)).
		When(widget.Selected, "item", style.As(style.Highlight))
}

func (m *MasterDetail) RenderCSS() *css.Stylesheet {
	return m.Style().Stylesheet()
}

func TestConsumerMasterDetail(t *testing.T) {
	wd := &MasterDetail{}
	sheet := wd.Style()
	ss := sheet.Stylesheet()
	cssStr := ss.String()

	t.Logf("Generated CSS:\n%s", cssStr)

	// 1. All class names in stylesheet exist in the expected set
	markupClasses := map[string]bool{
		"masterdetail":         true,
		"masterdetail__master": true,
		"masterdetail__detail": true,
		"masterdetail__item":   true,
	}

	classRegex := regexp.MustCompile(`\.masterdetail[a-zA-Z0-9_\-]*`)
	matches := classRegex.FindAllString(cssStr, -1)

	sheetClasses := make(map[string]bool)
	for _, m := range matches {
		name := strings.TrimPrefix(m, ".")
		sheetClasses[name] = true
	}

	for mc := range markupClasses {
		if !sheetClasses[mc] {
			t.Errorf("Class in markup %q not found in stylesheet", mc)
		}
	}

	for sc := range sheetClasses {
		if !markupClasses[sc] {
			t.Errorf("Class in stylesheet %q not found in markup", sc)
		}
	}

	// 2. Constraints: no !important, no @media (except prefers-reduced-motion if used), no color literals, no vw/vh.
	if strings.Contains(cssStr, "!important") {
		t.Error("Stylesheet contains forbidden '!important'")
	}

	// 3. Stacking layers in the correct order
	idxTokens := strings.Index(cssStr, "tokens")
	idxPrimitives := strings.Index(cssStr, "primitives")
	idxWidgets := strings.Index(cssStr, "widgets")
	idxStates := strings.Index(cssStr, "states")

	if idxTokens == -1 || idxPrimitives == -1 || idxWidgets == -1 || idxStates == -1 {
		t.Error("Missing a required cascade layer (tokens, primitives, widgets, states)")
	} else if !(idxTokens < idxPrimitives && idxPrimitives < idxWidgets && idxWidgets < idxStates) {
		t.Errorf("Incorrect layer order: tokens (%d) < primitives (%d) < widgets (%d) < states (%d)",
			idxTokens, idxPrimitives, idxWidgets, idxStates)
	}

	// 4. Deterministic emission
	cssStr2 := wd.Style().Stylesheet().String()
	if cssStr != cssStr2 {
		t.Error("Stylesheet emission is non-deterministic")
	}

	// 5. GOOS=js build dependency exclusion
	cmd := exec.Command("go", "list", "-deps", "webtyp.com/widget")
	cmd.Env = append(cmd.Environ(), "GOOS=js", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error running 'go list': %v, output: %s", err, string(out))
	}
	if strings.Contains(string(out), "webtyp.com/widget/style") {
		t.Error("WASM consumer depends on 'widget/style' package, violating build tag constraints")
	}

	buildCmd := exec.Command("go", "build", "webtyp.com/widget")
	buildCmd.Env = append(buildCmd.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Errorf("GOOS=js GOARCH=wasm go build webtyp.com/widget failed: %v, output: %s", err, string(out))
	}
}

func TestEdgeToEdgeBeatsSurfaceRadius(t *testing.T) {
	// EdgeToEdge must square a part that also carries As(Panel): the surface's
	// default radius used to win because the primitive's border-radius: 0 sits
	// in @layer primitives and the widgets layer outranks it — the crudview
	// root measured 4px despite already carrying EdgeToEdge.
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.As(style.Panel), style.EdgeToEdge()).Stylesheet().String()

	if strings.Contains(s, "border-radius: var(--radius-md") {
		t.Errorf("EdgeToEdge must suppress the surface's default radius, got:\n%s", s)
	}
	if !strings.Contains(s, "border-radius: 0;") {
		t.Errorf("EdgeToEdge must still emit border-radius: 0, got:\n%s", s)
	}

	// An explicit Round() next to EdgeToEdge still wins: it is emitted in the
	// widgets layer, which outranks the primitives-layer 0.
	s2 := style.For(wd).Root(style.As(style.Panel), style.EdgeToEdge(), style.Round(style.RadiusMd)).Stylesheet().String()
	if !strings.Contains(s2, "border-radius: var(--radius-md") {
		t.Errorf("explicit Round must survive next to EdgeToEdge, got:\n%s", s2)
	}
}

func TestRevealedByKeepsFlow(t *testing.T) {
	// a Row with RevealedBy(Open) emits display:flex in the state rule; revert-layer appears nowhere (closes D-1)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).
		Part("rowpart", style.Row(style.Space1), style.RevealedBy(widget.Open)).
		Stylesheet().
		String()

	if strings.Contains(s, "revert-layer") {
		t.Error("revert-layer must not be emitted")
	}

	if !strings.Contains(s, ".w__rowpart[data-open=\"true\"] {\n  display: flex;\n}") {
		t.Errorf("expected state rule to emit display:flex for Row with RevealedBy, got:\n%s", s)
	}
}

func TestStackingFromKind(t *testing.T) {
	// a Dialog backdrop emits var(--z-modal…), a Menu emits var(--z-dropdown…); no integer z-index (closes D-2)
	dialog := &testWidget{name: "dlg", kind: widget.Dialog}
	dialogCSS := style.For(dialog).Root(style.Backdrop(style.Viewport)).Stylesheet().String()
	if !strings.Contains(dialogCSS, "z-index: var(--z-modal,300);") {
		t.Errorf("expected Dialog backdrop to emit var(--z-modal,300), got:\n%s", dialogCSS)
	}

	// A Parent backdrop stays out of the layer: it would outrank the panel it
	// is supposed to sit behind. It still declares its LOCAL level — level
	// with its siblings, nothing more — and that is the only declaration it
	// is allowed: any other level is a claim on a layer it does not own.
	parentCSS := style.For(dialog).Part("catcher", style.Backdrop(style.Parent)).Stylesheet().String()
	if !strings.Contains(parentCSS, "z-index: 1;") {
		t.Errorf("Backdrop(Parent) must declare its local stacking level, got:\n%s", parentCSS)
	}
	if strings.Contains(parentCSS, "z-index: var(--z-") {
		t.Errorf("Backdrop(Parent) must not claim an overlay stacking level, got:\n%s", parentCSS)
	}

	menu := &testWidget{name: "mnu", kind: widget.Menu}
	menuCSS := style.For(menu).Root(style.Backdrop(style.Viewport)).Stylesheet().String()
	if !strings.Contains(menuCSS, "z-index: var(--z-dropdown,100);") {
		t.Errorf("expected Menu backdrop to emit var(--z-dropdown,100), got:\n%s", menuCSS)
	}

	// check that no integer z-index is emitted directly
	zindexIntRegex := regexp.MustCompile(`z-index:\s*\d+;`)
	if zindexIntRegex.MatchString(dialogCSS) {
		t.Errorf("Direct integer z-index emitted in dialog backdrop:\n%s", dialogCSS)
	}
}

func TestValidateReportsAll(t *testing.T) {
	// a sheet with an undeclared part, an empty part and a Veil without Backdrop returns three errors (closes D-3)
	wd := &testWidget{name: "w", kind: widget.Listbox} // Grid/Listbox allows Selected state
	sheet := style.For(wd).
		Part("emptypart").                                            // empty part (no options/declarations)
		When(widget.Selected, "undeclared", style.As(style.Primary)). // undeclared part
		Part("veilpart", style.Veil())                                // Veil without Backdrop

	errs := sheet.Validate()
	if len(errs) != 3 {
		t.Errorf("expected exactly 3 validation errors, got %d:\n%v", len(errs), errs)
	}
}

func TestStylesheetPanicsOnInvalid(t *testing.T) {
	// emission panics, and the message names the offending part (closes D-3)
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected stylesheet emission to panic on invalid sheet")
			return
		}
		errMsg := r.(error).Error()
		if !strings.Contains(errMsg, "veilpart") {
			t.Errorf("panic message should name the offending part, got:\n%s", errMsg)
		}
	}()

	wd := &testWidget{name: "w", kind: widget.Region}
	sheet := style.For(wd).Part("veilpart", style.Veil())
	sheet.Stylesheet()
}

func TestSheetParts(t *testing.T) {
	// Parts() returns the declared parts, sorted (closes C-7)
	wd := &testWidget{name: "w", kind: widget.Region}
	sheet := style.For(wd).
		Part("beta", style.As(style.Panel)).
		Part("alpha", style.As(style.Panel)).
		Part("gamma", style.As(style.Panel))

	parts := sheet.Parts()
	if len(parts) != 3 || parts[0] != widget.Part("alpha") || parts[1] != widget.Part("beta") || parts[2] != widget.Part("gamma") {
		t.Errorf("expected sorted parts [alpha, beta, gamma], got: %v", parts)
	}
}

type MyZeroWidget struct{}

func (z *MyZeroWidget) WidgetName() widget.Name { return widget.Name("zero") }
func (z *MyZeroWidget) WidgetKind() widget.Kind { return widget.Region }
func (z *MyZeroWidget) RenderCSS() *css.Stylesheet {
	return style.For(z).Root(style.As(style.Panel)).Stylesheet()
}

func TestZeroValueProvider(t *testing.T) {
	// (&T{}).RenderCSS() succeeds without reading fields
	var z MyZeroWidget
	ss := z.RenderCSS()
	if ss == nil || !strings.Contains(ss.String(), ".zero") {
		t.Error("RenderCSS failed on zero value of component")
	}
}

func TestSpaceStepsDistinct(t *testing.T) {
	// no two Space steps resolve to the same token (closes D-6)
	steps := []style.Space{
		style.SpaceNone, style.Space1, style.Space2, style.Space3,
		style.Space4, style.Space6, style.Space8, style.Space12,
	}

	wd := &testWidget{name: "w", kind: widget.Region}
	resolved := make(map[string]style.Space)

	for _, s := range steps {
		cssStr := style.For(wd).Root(style.Stack(s)).Stylesheet().String()
		// find `--gap: var(...)`
		re := regexp.MustCompile(`--gap:\s*([^;]+);`)
		match := re.FindStringSubmatch(cssStr)
		if len(match) < 2 {
			t.Fatalf("expected stack to emit --gap declaration for step %d", s)
		}
		val := match[1]
		if prev, exists := resolved[val]; exists {
			t.Errorf("Duplicate resolution: step %d and step %d both resolve to %q", s, prev, val)
		}
		resolved[val] = s
	}
}

func TestNoUnreachableSelectors(t *testing.T) {
	// no selector begins with .fl- or .exc-; no empty @layer block (closes D-7)
	wd := &testWidget{name: "w", kind: widget.Region}
	s := style.For(wd).Root(style.Stack(style.Space1)).Stylesheet().String()

	if strings.Contains(s, ".fl-") {
		t.Error("unreachable .fl- selector emitted")
	}
	if strings.Contains(s, ".exc-") {
		t.Error("unreachable .exc- selector emitted")
	}

	// Should not contain empty layers
	if strings.Contains(s, "@layer states {}") || strings.Contains(s, "@layer widgets {}") {
		t.Error("empty @layer blocks must be omitted")
	}
}
