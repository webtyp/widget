//go:build !wasm

package style

import (
	"webtyp.com/css"
	"webtyp.com/widget"
)

// Option is a visual option that configures a rule.
type Option func(*rule)

// rule contains all accumulated visual properties for an element.
type rule struct {
	hasFlow    bool
	flowType   flowType
	flowGap    Space
	flowRatio  SplitRatio
	flowAspect Aspect
	flowWidth  ColumnWidth
	flowSide   Side
	flowRail   RailWidth
	flowDetail Size
	flowMotion Motion
	flowCols   int

	hasDrawer    bool
	drawerSide   Side
	drawerSize   Size
	drawerMotion Motion

	hasEdgeStrip   bool
	edgeStripScope Scope
	edgeStripSide  Side

	hasMeter       bool
	meterThickness Space
	centerSelf     bool

	centerContent  bool
	startContent   bool
	controlBox     bool
	iconCap        bool
	logoBox        bool
	chipBox        bool
	buttonBox      bool
	visuallyHidden bool
	hasInteractive bool
	capitalize     bool

	hasGlyph bool
	glyph    Surface

	hasDocked   bool
	dockedScope Scope
	dockedEdge  Edge
	dockedSide  Side
	dockedGap   Space

	hasFloatMiddle  bool
	floatMiddleSide Side
	floatMiddleGap  Space

	hasOnEdge    bool
	onEdgeEdge   Edge
	onEdgeSide   Side
	onEdgeBlock  Space
	onEdgeInline Space

	// hasChipSeat reserves half a chip-height of block padding on ONE edge, the
	// space an OnEdge chip on a child straddles into. Pairs with OnEdge on the
	// child: the chip sits flush against this container's padding-box edge and
	// its centre lands exactly on the content's edge line.
	hasChipSeat  bool
	chipSeatEdge Edge

	hasFloatingChrome  bool
	floatingChromeEdge Edge
	floatingChromeSize IconSize
	floatingChromeGap  Space

	hasAnchor  bool
	foreground bool
	hasFlyout  bool
	flyoutSide Side

	hasDivider        bool
	dividerSide       Side
	hasDividerBelow   bool
	hasDividerBetween bool

	hasSurface  bool
	surface     Surface
	interactive bool
	// interactiveFamily is the family the hover/focus/press states derive
	// from. It is a separate decision from the resting surface: As() writes
	// only surface, Interactive() writes only interactiveFamily. One intent
	// ("resting look Y, interaction family X") is exactly one path —
	// As(Y) + Interactive(X) — instead of two options racing for one field.
	interactiveFamily Surface

	// overlay marca una regla de ESTADO (When/Cue/CueWithin). Un estado se pinta
	// encima de la caja base: no puede cambiar su tamaño, porque el elemento
	// crecería justo cuando el puntero está encima y perdería el propio hover que
	// lo activó. Es lo que hace que As() emita un anillo de box-shadow (ver
	// boxShadowDecls) en vez de border.
	overlay bool

	hasPad bool
	pad    Space

	hasPadEdge   bool
	padEdge      Edge
	padEdgeSpace Space

	hasPadInline bool
	padInline    Space

	hasRound bool
	round    Radius

	hasGradientAngle bool
	gradientAngle    string

	hasRaise bool
	raise    Elevation

	hasSize bool
	size    Size

	hasCapped bool
	capped    Extent

	hasIcon bool
	icon    IconSize

	fill         bool
	grow         bool
	pushEnd      bool
	scroll       bool
	keepSize     bool
	edgeToEdge   bool
	hideOverflow bool

	hasScrollGutter bool
	scrollGutter    Space

	hasTextSize bool
	textSize    TextSize
	hasWeight   bool
	weight      Weight

	hasMotion bool
	motion    Motion

	hasRotate bool
	rotate    Turn

	hasBackdrop   bool
	backdropScope Scope
	hasVeil       bool
	revealedBy    widget.State
	hasRevealed   bool

	hidden bool
	shown  bool
}

type stateKey struct {
	state widget.State
	part  widget.Part
}

type cueKey struct {
	cue  widget.Cue
	part widget.Part
}

// cueWithinKey addresses a part through an ancestor's cue. Every other rule in
// this package is one selector on one element; this is the single exception,
// and it exists because a container that reveals its own children on hover — a
// nav rail that expands — has no other expression.
type cueWithinKey struct {
	cue       widget.Cue
	container widget.Part
	part      widget.Part
}

// stateWithinKey addresses a part through an ancestor's STATE — the written
// counterpart of cueWithinKey. It exists because a state is written onto the
// element that owns it, which is not always the element it should repaint: a
// field's read-only gate belongs to the field, but what has to look read-only
// is the control inside it.
type stateWithinKey struct {
	state     widget.State
	container widget.Part
	part      widget.Part
}

// cueAcrossKey addresses a part through a cue on some REGION, with no assumed
// DOM relationship between them: it is checked from the root with :has(), so
// the region and the part may sit anywhere. The escape hatch for the cases
// cueWithinKey (descendant) cannot reach — floating chrome that yields while
// the module content region has focus within it, wherever each lives in the
// tree.
type cueAcrossKey struct {
	cue    widget.Cue
	region widget.Part
	part   widget.Part
}

// stateAcrossKey is the written-state counterpart of cueAcrossKey: it fires
// while the region CONTAINS an element carrying the state
// (`.n:has(.n__region [data-x="true"]) .n__part`), for a state a module sets
// deep inside the content region — a record being edited — that the same
// floating chrome must also yield to.
type stateAcrossKey struct {
	state  widget.State
	region widget.Part
	part   widget.Part
}

type deviceKey struct {
	device css.Device
	part   widget.Part
}

// Sheet represents a scoped stylesheet for a widget.
type Sheet struct {
	widget         widget.Widget
	rootRule       rule
	partRules      map[widget.Part]rule
	stateRules     map[stateKey]rule
	stateWithin    map[stateWithinKey]rule
	cueRules       map[cueKey]rule
	cueWithin      map[cueWithinKey]rule
	cueWithinHover map[cueWithinKey]rule
	cueAcross      map[cueAcrossKey]rule
	stateAcross    map[stateAcrossKey]rule
	deviceRules    map[deviceKey]rule

	// within records the part tree: which part renders inside which. The
	// sheet needs it to reason about positioning (who is whose containing
	// block), because a Flyout's inset resolves against the nearest
	// POSITIONED ancestor, which is not always the Anchor the author meant.
	within map[widget.Part]widget.Part
}

// For opens the styling block for a widget.
func For(w widget.Widget) *Sheet {
	return &Sheet{
		widget:         w,
		partRules:      make(map[widget.Part]rule),
		stateRules:     make(map[stateKey]rule),
		stateWithin:    make(map[stateWithinKey]rule),
		cueRules:       make(map[cueKey]rule),
		cueWithin:      make(map[cueWithinKey]rule),
		cueWithinHover: make(map[cueWithinKey]rule),
		cueAcross:      make(map[cueAcrossKey]rule),
		stateAcross:    make(map[stateAcrossKey]rule),
		deviceRules:    make(map[deviceKey]rule),
		within:         make(map[widget.Part]widget.Part),
	}
}
