//go:build !wasm

package style

// Space is the spacing scale: 8 steps mirroring --space-N.
type Space uint8

const (
	SpaceNone Space = iota
	Space1
	Space2
	Space3
	Space4
	Space6
	Space8
	Space12
)

// Radius is the border radius scale.
type Radius uint8

const (
	RadiusNone Radius = iota
	RadiusSm
	RadiusMd
	RadiusLg
	RadiusFull
)

// TextSize is the typography size scale.
type TextSize uint8

const (
	TextXs TextSize = iota
	TextSm
	TextBase
	TextLg
	TextXl
	Text2xl
)

// Weight is the font weight scale.
type Weight uint8

const (
	WeightRegular Weight = iota
	WeightMedium
	WeightBold
)

// Edge is a block-axis edge of a box.
type Edge uint8

const (
	EdgeTop Edge = iota
	EdgeBottom
)

// IconSize is the square size scale for icon-sized parts. The steps are
// relative to the inherited font size, so an icon tracks the text it sits with.
type IconSize uint8

const (
	IconSm IconSize = iota // inline with a line of text
	IconMd                 // a control's icon: button, field affix
	IconLg                 // a navigation rail or toolbar icon
)

// Elevation is the shadow elevation scale.
type Elevation uint8

const (
	Flat Elevation = iota
	Raised
	Floating
	Popover
)

// SplitRatio is the flex-grow ratio for Split partitions.
type SplitRatio uint8

const (
	SplitHalf SplitRatio = iota
	SplitTwoThirds
	SplitThreeQuarters
)

// Aspect is the aspect ratio for MediaBox containers.
type Aspect uint8

const (
	AspectSquare Aspect = iota
	Aspect3x2
	Aspect4x3
	Aspect16x9
)

// ColumnWidth represents minimum column width for grids.
type ColumnWidth uint8

const (
	ColumnNarrow ColumnWidth = iota
	ColumnMedium
	ColumnWide
)

// Side names which edge a Sidebar's rail or a Drawer's panel is anchored to.
// Logical, not physical: it follows writing direction.
type Side uint8

const (
	SideStart Side = iota // inline-start — left in LTR
	SideEnd               // inline-end   — right in LTR
)

// RailWidth is the closed scale for a Sidebar's fixed column.
type RailWidth uint8

const (
	RailNarrow RailWidth = iota // icon only
	RailWide                    // icon plus label
)

// Size is the relative size measurement.
type Size uint8

const (
	Content  Size = iota // adjusts to its content
	Readable             // readable line length
	Compact              // a single column of controls: a login card, a settings pane
	Third
	Half
	TwoThirds
	Most // 90% — leaves a sliver of what sits behind it
	Full // 100% of the container
)

// Extent is the share of the VIEWPORT's block axis that Capped() lets an
// element occupy. Deliberately a viewport share and not a container share:
// the elements that need a ceiling are the ones taken out of the flow —
// a Flyout, a Drawer — whose containing block is an ancestor box that says
// nothing about how much screen is left.
//
// The steps are dvh, not vh: on a phone the browser chrome shrinks and grows
// as the user scrolls, and vh is frozen at the LARGEST of those two states,
// so a panel sized in vh hangs below the fold exactly when the toolbar is
// showing — the one moment the user is looking at it.
type Extent uint8

const (
	ExtentHalf Extent = iota // 50dvh — leaves the page behind clearly readable
	ExtentMost               // 70dvh — the default for a picker: long list, page still visible
	ExtentFull               // 100dvh — the panel is the screen
)

// Motion is the transition scale. Duration is owned by CSS;
// here we only select the level.
type Motion uint8

const (
	MotionNone Motion = iota // no transition
	MotionFast               // immediate highlight: hover, focus
	MotionBase               // state change
	MotionSlow               // panel/overlay transition
)

// Turn is the closed set of rotations the DSL supports — quarter-turn steps,
// which is every rotation a chevron, caret or disclosure arrow needs. A free
// degree value is deliberately not accepted: it would be a generic hole with
// no intent, and no part in the ecosystem has ever needed one.
type Turn uint8

const (
	TurnNone         Turn = iota // 0deg — the resting position
	TurnQuarter                  // 90deg
	TurnHalf                     // 180deg — a chevron flipped to point the other way
	TurnThreeQuarter             // 270deg
)

// Cadence is how long each frame of a Frame() flipbook stays on screen.
type Cadence uint8

const (
	CadenceFast Cadence = iota // 150ms per frame — syllables of speech
	CadenceSlow                // 400ms per frame — dots of a "thinking" indicator
)
