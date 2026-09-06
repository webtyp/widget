# Specification — `webtyp/widget`

Strict functional requirements: exact public surface, exact scale mappings, exact
emitted output, exact failure conditions. Structure and reasoning are not repeated
here — see [ARCHITECTURE.md](ARCHITECTURE.md) and [DESIGN.md](DESIGN.md).

Every table below is a test assertion. An implementation is correct when it
matches these tables byte for byte.

---

## 1. Package `widget` (WASM-compatible)

```go
type Name string
type Part string
type Class string   // no public constructor

func (n Name) Root() Class
func (n Name) Class(p Part) Class
func (c Class) String() string
func (c Class) AsAttr() fmt.KeyValue

type Kind uint8
const (Region, Listbox, Menu, Dialog, Disclosure, Tabs, Toolbar, Grid, Combobox, Form, Alert)
func (k Kind) Role() fmt.KeyValue
func (k Kind) Layer() Layer        // stacking level of the pattern
func (k Kind) Allows(s State) bool

// Layer is the stacking level of a pattern. It is an enum, not a css.Token:
// this package travels inside the WASM binary and must not import css.
// widget/style maps it to the --z-* catalog.
type Layer uint8
const (LayerBase, LayerDropdown, LayerSticky, LayerModal, LayerToast, LayerTooltip)

type State uint8
const (Selected, Disabled, Locked, Invalid, Busy, Open, Current)

// StateAttr is the DOM projection of a State. Unexported fields; the only
// source is State.Attr(). Key()/Value() exist for dom and widget/style.
type StateAttr struct{ /* key, value */ }
func (a StateAttr) Key() string
func (a StateAttr) Value() string
func (s State) Attr() StateAttr
func (s State) Key() string   // = Attr().Key() — lets a State itself be passed to dom.BindState
func (s State) Value() string // = Attr().Value()

type Cue uint8
const (Hover, Focus, Press, Target)

type Widget interface {
    WidgetName() Name
    WidgetKind() Kind
}

type Selectable   interface{ Select(id string) }
type Dismissible  interface{ Dismiss() }
type Expandable   interface{ Expand(open bool) }
type Filterable   interface{ OnFilterChange(func(term string)) }
```

### 1.1 Class derivation

| Call | Result |
|---|---|
| `Name("targetlist").Root()` | `targetlist` |
| `Name("targetlist").Class("row")` | `targetlist__row` |

### 1.2 State attributes

| `State` | attribute |
|---|---|
| `Selected` | `data-selected="true"` |
| `Disabled` | `data-disabled="true"` |
| `Locked` | `data-locked="true"` |
| `Invalid` | `data-invalid="true"` |
| `Busy` | `data-busy="true"` |
| `Open` | `data-open="true"` |
| `Current` | `data-current="true"` |

### 1.3 `Kind` → role, stacking level, allowed states

| `Kind` | `Role()` | `Layer()` | states beyond the universal set |
|---|---|---|---|
| `Region` | `role="region"` | `LayerBase` | — |
| `Listbox` | `role="listbox"` | `LayerBase` | `Selected`, `Current` |
| `Menu` | `role="menu"` | `LayerDropdown` | `Open`, `Current` |
| `Dialog` | `role="dialog"` | `LayerModal` | `Open` |
| `Disclosure` | `role="group"` | `LayerDropdown` | `Open` |
| `Tabs` | `role="tablist"` | `LayerBase` | `Selected`, `Current` |
| `Toolbar` | `role="toolbar"` | `LayerBase` | — |
| `Grid` | `role="grid"` | `LayerBase` | `Selected`, `Current` |
| `Combobox` | `role="combobox"` | `LayerDropdown` | `Open`, `Selected`, `Invalid` |
| `Form` | `role="form"` | `LayerBase` | `Invalid`, `Busy` |
| `Alert` | `role="alert"` | `LayerToast` | `Open` |

Universal set, allowed for every `Kind`: `Disabled`, `Locked`, `Busy`.

`widget/style` maps `Layer` onto the catalog: `LayerBase`→`--z-base`,
`LayerDropdown`→`--z-dropdown`, `LayerSticky`→`--z-sticky`,
`LayerModal`→`--z-modal`, `LayerToast`→`--z-toast`,
`LayerTooltip`→`--z-tooltip`.

**Two stacking levels, one rule.** An out-of-flow element either claims the
overlay level its `Kind` owns (`Flyout`, `Drawer`, `Backdrop(Viewport)`,
`Docked(Viewport)` → `z-index:var(<Kind layer>)`) or — when it stays inside its
own box (`OnEdge`, `Backdrop(Parent)`, `Docked(Parent)`) — declares the local
level `z-index:1`: enough to order it against unpositioned siblings (Safari
otherwise paints a sibling form control over the legend), never a claim on a
layer it does not own. `stackingFor` is the only source of a `z-index` value.

---

## 2. Package `widget/style` — scales

All values are `webtyp/css` token references. No literal may be emitted for any
of these.

### 2.1 `Space` — 8 steps, 8 distinct tokens

| Constant | Token | Value |
|---|---|---|
| `SpaceNone` | `--space-0` | `0` |
| `Space1` | `--space-1` | `0.25rem` |
| `Space2` | `--space-2` | `0.5rem` |
| `Space3` | `--space-3` | `0.75rem` |
| `Space4` | `--space-4` | `1rem` |
| `Space6` | `--space-6` | `1.5rem` |
| `Space8` | `--space-8` | `2rem` |
| `Space12` | `--space-12` | `3rem` |

**Invariant:** no two steps resolve to the same token.

### 2.2 Other scales

| Type | Constants | Tokens |
|---|---|---|---|
| `Radius` | `RadiusNone, RadiusSm, RadiusMd, RadiusLg, RadiusFull` | `0`, `--radius-sm/md/lg/full` |
| `TextSize` | `TextXs, TextSm, TextBase, TextLg, TextXl, Text2xl` | `--text-*` |
| `Weight` | `WeightRegular, WeightMedium, WeightBold` | `--font-weight-*` |
| `Elevation` | `Flat, Raised, Floating, Popover` | `none`, `--shadow-sm/md/lg` |
| `Motion` | `MotionNone, MotionFast, MotionBase, MotionSlow` | `none`, `--duration-*` + `--ease-in-out` |
| `Turn` | `TurnNone, TurnQuarter, TurnHalf, TurnThreeQuarter` | `0deg`, `90deg`, `180deg`, `270deg` — quarter-turn steps only; a free degree value is a generic hole with no intent |
| `ColumnWidth` | `ColumnNarrow, ColumnMedium, ColumnWide` | `--column-narrow/medium/wide` |
| shared boxes | `ControlBox()`, `ChipBox()` | `--control-height`, `--chip-width`, `--chip-height` |
| `Size` | `Content, Readable, Compact, Third, Half, TwoThirds, Most, Full` | `max-content`, `--max-w-readable`, `min(100%, 24rem)`, `33.33%`, `50%`, `66.66%`, `90%`, `100%` |
| `SplitRatio` | `SplitHalf, SplitTwoThirds, SplitThreeQuarters` | `1`, `2`, `3` — unitless, they feed `flex-grow` against a trailing `1` |
| `Aspect` | `AspectSquare, Aspect3x2, Aspect4x3, Aspect16x9` | `1/1`, `3/2`, `4/3`, `16/9` |
| `Scope` | `Parent, Viewport` | `position: absolute` / `fixed` |
| `Side` | `SideStart, SideEnd` | `inline-start`, `inline-end` |
| `RailWidth` | `RailNarrow, RailWide` | `--rail-narrow`, `--rail-wide` |
| `Edge` | `EdgeTop, EdgeBottom` | block-axis edge: `inset-block-start` / `inset-block-end` |

`Size` percentages and `Aspect` fractions are geometry, not theme, and are the
only literals the drift guard permits, **except** `100dvh` which is permitted
exclusively in the `Cover` primitive, the `0px` defaults of the floating
reservation (`--floating-top`/`--floating-bottom`), the `-0.5` factor of the
`OnEdge` straddle and the `24rem` cap of `Compact`. The `Turn` angles emit
`0deg`/`90deg`/`180deg`/`270deg` — rotation is the same class of pure geometry,
and every cheat a chevron needs is a quarter-turn step. `Compact` is the one `Size`
step that is a cap rather than a share — the percentages resolve against the
container and shrink with it for free, while a fixed width would overflow a
phone, so it emits `min(100%, …)` and stops. `ColumnWidth` requires `--column-*`, and `Readable` requires `--max-w-readable`
(today `--max-w-prose`); both are prerequisites of the release, because a scale
step is named after the token it emits and the two must not drift apart.

---

## 3. Surfaces

```go
type Surface uint8
const (
    Page, Panel, Inset, Primary, Secondary, Highlight, Accent, Success, Danger, Subtle, Inactive
)
```

A surface resolves background, text, border and **radius** together. It does
**not** resolve padding — reasoning in
[DESIGN.md §5](DESIGN.md#5-why-a-surface-carries-shape).

| Surface | Background | Text | Border | Radius |
|---|---|---|---|---|
| `Page` | `--color-background` | `--color-on-surface` | — | none |
| `Panel` | `--color-surface` | `--color-on-surface` | `1px solid --color-outline` | `RadiusMd` |
| `Inset` | `--color-surface-sunken` | `--color-on-surface` | `1px solid --color-outline` | `RadiusSm` |
| `Primary` | `--color-primary` | `--color-on-primary` | — | `RadiusSm` |
| `Secondary` | `--color-surface` | `--color-on-surface` | — | `RadiusSm` |
| `Highlight` | `--color-selection` | `--color-on-selection` | — | `RadiusSm` |
| `Accent` | `--color-accent` | `--color-on-accent` | — | `RadiusSm` |
| `Success` | `--color-success` | `--color-on-success` | — | `RadiusSm` |
| `Danger` | `--color-danger` | `--color-on-danger` | — | `RadiusSm` |
| `DangerWash` | `--color-danger-wash` | `--color-on-surface` | — | `RadiusSm` |
| `Subtle` | `transparent` | `--color-muted` | — | none |
| `Inactive` | `--color-surface` | `--color-muted` | — | `RadiusSm` |

Explicit `Round()` or `Raise()` on the same rule overrides the surface default.
`Pad()` is always explicit — there is no default to override.

**A state never changes the box size.** A state rule (`When`, `Cue`, `CueWithin`)
repaints a bordered surface as a shadow ring — `box-shadow: 0 0 0 1px <border color>`,
the border's width a hair's breadth outside the box — instead of `border:`. The ring
takes no layout space, so the element does not grow under the pointer that entered
it. The static and enhanced halves are emitted as a double declaration, exactly like
every other themed color; when the state rule also carries `Raise()`, the ring and
the elevation compose into ONE declaration, ring first. The ring is not an
`outline`: outlines paint at the END of the stacking context (a Locked border
crossed over a legend riding the same line) and Safari < 16.4 ignores
`border-radius` on them. Base rules (`Root`, `Part`, `On`, `OnlyOn`) keep `border:`.

### 3.1 Interaction families

`Interactive(s)` emits the base surface plus three state rules. The interactive states
are derived programmatically from the base token of the surface family using functions
`css.Hover(base)`, `css.Focus(base)`, and `css.Press(base)` defined in `webtyp/css`.

| Cue | Selector suffix | Change |
|---|---|---|
| hover | `:hover` | background → `css.Hover(base)` |
| focus | `:focus-visible` | background → `css.Focus(base)` |
| press | `:active` | background → `css.Press(base)` |

### 3.2 Which surfaces accept `Interactive`

| Surface | `Interactive` | Why |
|---|---|---|
| `Panel`, `Inset`, `Primary`, `Secondary`, `Highlight`, `Success`, `Danger`, `Subtle` | yes | can be a control or a selectable row |
| `Page` | no | the page background is not a control |
| `Inactive` | no | it *is* the non-interactive state; deriving a hover from it is a contradiction |

`Interactive(Page)` and `Interactive(Inactive)` are reported by `Validate()`
(§6.1). This is the eight-family set the private `--color-<family>-*` tokens must
cover — 24 tokens, not 27.

**Invariant:** it is not expressible to combine one family's base with another
family's interaction state.

---

## 4. Flow primitives

| Option | Emits |
|---|---|---|
| `Stack(gap)` | `display:flex; flex-direction:column; gap:var(--gap); min-height:0` — the gap lives on the container, never on a `> * + *` rule that would resolve `var(--gap)` against the child |
| `Row(gap)` | `display:flex; flex-wrap:wrap; gap:var(--gap); align-items:center` |
| `Split(r, gap)` | `display:flex; flex-wrap:wrap; gap:var(--gap)`, and `> * { flex-grow:1; flex-basis:calc((40rem - 100%) * 999) }` plus `> :first-child { flex-grow:var(--ratio) }` — stacks below ~40rem of **its own** width, no query and no wrapper element |
| `Grid(min, gap)` | `display:grid; gap:var(--gap); grid-template-columns:repeat(auto-fit, minmax(min(var(--column),100%),1fr))` |
| `FixedGrid(cols, gap)` | `display:grid; gap:var(--gap); grid-template-columns:repeat(var(--cols), minmax(0,1fr))` — unlike `Grid`, the column count never reflows. `--cols` is a value, not a literal `repeat(N,1fr)`: a stylesheet builder works on a zero-value receiver and cannot read an instance field, so a count only known at runtime (a month strip sized by an instance's field) is set the same way any other per-instance value crosses into an otherwise-static sheet — the host overrides `--cols` inline on the element |
| `Center(max)` | `margin-inline:auto; width:100%; max-width:var(--max-width)` |
| `FillCentered()` | `display:grid; place-items:center; min-height:100%; width:100%` |
| `ScrollRow(gap)` | `display:flex; gap:var(--gap); overflow-x:auto; scroll-snap-type:x mandatory; scroll-behavior:smooth`, and `> * { scroll-snap-align:start; flex:0 0 auto }` |
| `MediaBox(a)` | `aspect-ratio:var(--ratio); overflow:hidden; display:flex; justify-content:center; align-items:center`, and `> img, > video { width:100%; height:100%; object-fit:cover }` |
| `Cover()` | `display:flex; flex-direction:column; height:100dvh` |
| `MasterDetail(detail)` | `display:flex; flex-direction:row; flex-wrap:nowrap; direction:rtl; gap:0; overflow-x:auto; overflow-y:hidden; scroll-snap-type:x mandatory; scroll-behavior:smooth`, and `> * { flex:0 0 auto }` plus `> :nth-child(1) { direction:ltr; flex:0 0 <detail>; scroll-snap-align:end; order:2 }` plus `> :nth-child(2) { direction:ltr; flex:0 0 100%; scroll-snap-align:start; order:1 }` |
| `SlideDeck(m)` | `position:relative; overflow:hidden`, and `> * { position:absolute; inset:0; transform:translateX(-100%); visibility:hidden; transition:transform <dur> var(--ease-in-out), visibility 0s linear <dur> }` plus `> *[data-current="true"] { transform:translateX(0); visibility:visible; transition:transform <dur> var(--ease-in-out), visibility 0s }` — `<dur>` is the `Motion`'s duration token; `MotionNone` switches without animation |
| `AutoRotate()` | `position:relative; overflow:hidden`, and `> * { position:absolute; inset:0; opacity:0; animation:tw-auto-rotate 30s infinite }` plus `> :first-child { opacity:1 }` plus `> :nth-child(2)` through `> :nth-child(AutoRotateLayers)` each with a staggered `animation-delay`, plus the shared `@keyframes tw-auto-rotate` rule and an `@media (prefers-reduced-motion: reduce) { > * { animation:none } }` block — cross-fades children with no state and no JS. `AutoRotateLayers` (currently 6) is a fixed constant, not a parameter: like `FixedGrid`'s `--cols`, a stylesheet builder runs on a zero-value receiver and cannot read how many children a real instance renders, but unlike `--cols` this cannot be solved with an inline custom-property override either — the count changes keyframe *selector* percentages (e.g. `100/N%`), and `calc()`/`var()` are not valid there. A caller with fewer than `AutoRotateLayers` real children must tile them across all slots or unused slots go dark on their turn |
| `Sidebar(side, width, gap)` | `display:flex; flex-wrap:wrap; gap:var(--gap)`, and `> :first-child` / `> :last-child` rail/content split based on `side` |

No emitted selector may begin with `.fl-` or `.exc-`.

### 4.1 `Split` uses no container query, deliberately

The published `Split` sets `container-type: inline-size` on the same selector the
`@container` rule targets. **An element is never its own query container**, so the
rule never applies. Measured in Chromium at a 320px viewport, well below the
40rem breakpoint:

```
self-query   (published)  grid-template-columns: 213.328px 106.656px   two columns
ancestor-query (correct)  grid-template-columns: 320px                 collapsed
```

`Split` therefore has no responsive behaviour at all today. The correct
query-based form needs a wrapper element to act as the container, which would
couple the engine to DOM structure — something §7.1 forbids.

The specified form uses intrinsic sizing instead, and needs neither. Measured
across viewports:

| viewport | result | ratio |
|---|---|---|
| 320px | stacked | — |
| 500px | stacked | — |
| 800px | side by side | 2.00 : 1 |
| 1200px | side by side | 2.00 : 1 |

**Invariant:** the emitted sheet contains no `@container` rule and no
`container-type` declaration. Nothing else in the API produces one; `Grid` is
intrinsically responsive through `auto-fit`/`minmax`.

**Correction on record.** An earlier revision claimed `container-type` makes an
element a containing block for `position: fixed` descendants, and specified a
validation rule against `Backdrop(Viewport)` inside `Split`. That is false, and
was measured false: a fixed child inside a `container-type` element covers the
full viewport (320×600 of a 320×600 viewport), identical to no containment. Only
full `contain: layout` contains it (320×0). The rule was removed rather than
shipped.

---

## 5. Remaining options

```go
func As(s Surface) Option
func Interactive(s Surface) Option
func RevealedBy(st widget.State) Option
func Pad(Space) Option
func PadEdge(Edge, Space) Option
func Round(Radius) Option
func Raise(Elevation) Option
func Width(Size) Option
func IconBox(IconSize) Option
func FontSize(TextSize) Option
func FontWeight(Weight) Option
func Animate(Motion) Option
func Rotate(Turn) Option
func Fill() Option
func Grow() Option
func PushEnd() Option
func Glyph(Surface) Option
func ControlBox() Option
func ChipBox() Option
func Hide() Option
func CenterContent() Option
func CenterSelf() Option
func StartContent() Option
func Anchor() Option
func Docked(scope Scope, edge Edge, side Side, gap Space) Option
func EdgeStrip(scope Scope, side Side) Option
func Meter(thickness Space) Option
func OnEdge(edge Edge, side Side, block Space, inline Space) Option
func FloatingChrome(edge Edge, size IconSize, gap Space) Option
func Flyout(side Side) Option
func Scroll() Option
func KeepSize() Option
func EdgeToEdge() Option
func HideOverflow() Option
func Backdrop(Scope) Option
func Veil() Option
func Cover() Option
func Sidebar(side Side, width RailWidth, gap Space) Option
func SlideDeck(m Motion) Option
func AutoRotate() Option
func MasterDetail(detail Size) Option
func Drawer(side Side, size Size) Option
func FloatMiddle(side Side, gap Space) Option
```

| Option | Emits |
|---|---|
| `Fill()` | `height:100%; min-height:0; flex-grow:1` |
| `Grow()` | `flex-grow:1; min-width:0` — the Row counterpart of `Fill()`, no height claim |
| `PushEnd()` | `margin-inline-start:auto` — sends the item to the trailing edge of its line |
| `Scroll()` | `overflow-y:auto` plus everything `Fill()` emits, plus `padding-block-start:var(--floating-top,0px); padding-block-end:var(--floating-bottom,0px)` — reserves the strip declared by any `FloatingChrome()` ancestor |
| `KeepSize()` | `flex-shrink:0; flex-grow:0` |
| `EdgeToEdge()` | `margin:0; border-radius:0` |
| `HideOverflow()` | `overflow:hidden` |
| `PadEdge(e, s)` | `padding-block-{start\|end}:var(--space-N)` |
| `IconBox(s)` | `width:<1em\|1.5em\|2.5em>; height:<same>; flex-shrink:0` |
| `Backdrop(Parent)` | `position:absolute; inset:0; z-index:1` — the local level only: on the overlay layer it would tie with the panel it is supposed to sit behind |
| `Backdrop(Viewport)` | `position:fixed; inset:0; z-index:var(<Kind layer>)` |
| `Veil()` | `background-color: color-mix(in srgb, var(--color-surface,<fallback>) 60%, transparent)` |
| `Glyph(s)` | `color:<surface base>; fill:currentColor` — tints the content, paints no background |
| `ControlBox()` | `min-height:var(--control-height)` |
| `ChipBox()` | `width:var(--chip-width); overflow:hidden` |
| `Hide()` | `display:none` — for use inside `On()`, the inverse of `OnlyOn` |
| `CenterContent()` | `display:flex; align-items:center; justify-content:center` |
| `CenterSelf()` | `margin-inline:auto` — centers the part itself within the space its container gives it; the counterpart of `CenterContent()`, which centers what a part contains. Needed alongside `IconBox()`/`Width()` inside a wider flex/grid track: an item with an explicit size does not stretch to fill the track and defaults to the leading edge |
| `StartContent()` | `display:flex; align-items:center; justify-content:flex-start` |
| `Docked(scope, edge, side, gap)` | `position:{absolute\|fixed}; margin:0; inset-block-{start\|end}:<gap>; inset-inline-{start\|end}:<gap>`; `z-index:var(<Kind layer>)` for `Viewport`, `z-index:1` for `Parent` — a `Parent` dock pins the element to the corner of the nearest **positioned** ancestor (the `Anchor()` only if nothing positioned sits between) and is a control inside its own box: it declares its local level against its unpositioned siblings, but must not tie with real overlays. A `Parent` dock is itself a containing block: a `Flyout()` inside it hangs from THIS box, not from whatever `Anchor()` sits above — `Validate()` rejects that theft (see `Within()`) |
| `OnEdge(edge, side, block, inline)` | `position:absolute; margin:0; inset-block-{start\|end}:<block>; inset-inline-{start\|end}:<inline>; margin-block-{start\|end}:calc(-0.5 * var(--chip-height,1.25rem)); z-index:1` — the straddle is half the shared `--chip-height` token, never a `transform` (transforms are invisible to `scrollHeight` by spec, so no ancestor could reserve the space the chip really occupies, and they create an implicit stacking context); `z-index:1` orders the chip against its own siblings without claiming the overlay layer. Like every positioned part, it is a containing block for a `Flyout()` descendant |
| `FloatingChrome(edge, size, gap)` | declares the strip this element occupies along its edge as an inherited custom property — `--floating-{top\|bottom}:calc(<size> + 2 * <gap>)` — consumed by every `Scroll()` descendant via `var(--floating-{top\|bottom},0px)`. The seam between a floating action button and the scroll container behind it: the reservation crosses widget boundaries because the property is inherited, and no declaration without `FloatingChrome` (the `0px` default applies everywhere). `size` is an `IconSize`, not a `Size`: floating chrome pinned to a screen edge is by construction a small icon-only control (a FAB, a hamburger); a `Size`'s percentages would resolve against the scroll region's own inline size and `max-content` is not a `<length>` at all |
| `Anchor()` | `position:relative` — makes the element a positioning reference. The containing block of a `Flyout()` is the nearest **positioned** ancestor, so this only becomes the reference if nothing positioned sits between the two; `Validate()` rejects the interposition (see `Within()`) |
| `Flyout(side)` | `position:absolute; inset-block-start:100%; inset-inline-{start\|end}:0; z-index:var(<Kind layer>)` — the `100%` resolves against the nearest positioned ancestor; a positioned part between the element and its `Anchor()` steals the containing block, which the sheet detects through the declared part tree (`Within()`); a `Scroll()` part anywhere on the declared chain clips the panel, which the sheet also rejects (`Within()`) |
| `Drawer(side, size)` | `position:fixed; inset-block:0; inset-inline-{start|end}:0; width:var(<size>); z-index:var(<Kind layer>)` |
| `FloatMiddle(side, gap)` | `position:fixed; top:50%; transform:translateY(-50%); inset-block-end:auto; inset-inline-{start\|end}:var(--space-N); inset-inline-{start\|end}:auto` on the other side; `z-index:var(<Kind layer>)` — pins the element to the vertical middle of one viewport edge. The `50%` comes from `Size.Half`, never a literal; the centring transform is why it cannot combine with `Rotate()`/`Drawer()` (`Validate()` rejects both) |
| `EdgeStrip(scope, side)` | `position:{absolute\|fixed}; inset-block:0; inset-inline-{start\|end}:0`; `z-index:var(<Kind layer>)` for `Viewport`, `z-index:1` for `Parent` — spans the FULL block axis of its `Scope` plus one inline edge, and deliberately emits no `width:` — the property that differentiates it from `Drawer()`, which forces one and hard-requires a paired `RevealedBy()`. `EdgeStrip` carries no such requirement: it is permanent chrome (a calendar's prev/next overlay strip), not a toggled panel. Like `Docked`/`OnEdge`, a `Parent`-scoped `EdgeStrip` is itself a containing block for a `Flyout()` descendant |
| `Meter(thickness)` | `height:var(<thickness>); width:var(--meter-fill,0%)` — a thin bar whose fill fraction is a per-instance runtime value, not a scale step: the stylesheet declares only the SHAPE (that the length axis feeds `width`, with a `0%` fallback); the host sets ONLY the bare `--meter-fill:N%;` value inline, never a property name or selector |
| `Animate(m)` | `transition: all var(--duration-*) var(--ease-in-out)` — except on a rule that also carries `RevealedBy()`, where it becomes the animated reveal of §5.2 instead |
| `Rotate(t)` | `transform: rotate(<Turn degrees>)` — emits `0deg`/`90deg`/`180deg`/`270deg` by the `Turn` step. Pair with a state rule (`When`/`WhenWithin`) so the rotation IS the state, and with `Animate()` on the base rule so the turn is a transition. Not combinable with `OnEdge()`/`Drawer()`/`FloatMiddle()`: all already own the element's `transform`, and a second declaration would silently replace the first — `Validate()` rejects the combination |

`Veil()` must emit the token **with its catalog fallback**. Every rule carrying
`Animate` is repeated under `@media (prefers-reduced-motion: reduce)` with
`transition: none`.

### 5.1 `RevealedBy` display resolution

Base rule emits `display:none`. The state rule emits:

| Flow declared on the same part | `display` |
|---|---|
| `Stack`, `Row`, `Split`, `Sidebar`, `ScrollRow`, `MediaBox`, `Cover`, `MasterDetail` | `flex` |
| `Grid`, `FillCentered` | `grid` |
| `Center`, or no flow | `block` |

**Invariant:** `display: revert-layer` is never emitted — it resolves to the base
`display:none` and leaves the element hidden.

### 5.2 Animated reveal (`RevealedBy` + `Animate` on the same rule)

Pairing `Animate(m)` with `RevealedBy(st)` on one rule upgrades the instant
display swap into a choreographed fade — entry and exit alike, over the motion
duration. `MotionNone`, or no `Animate` at all, keeps the instant swap of §5.1.
No new option exists for this on purpose: the motion the author already
declares gains meaning for the reveal, the same way `Animate()` on a base rule
turns a `Rotate()` state change into a transition.

Emitted output for `Part("p", Row(Space2), Animate(MotionBase),
RevealedBy(Open))` on widget `w`:

- Base rule (wherever the base is emitted — part, root, or device block):
  `display: none;` (as in §5.1) plus `opacity: 0;` plus
  `transition: all var(--duration-base,250ms) var(--ease-in-out,…) , display var(--duration-base,250ms) allow-discrete;`
  (`all` keeps every other transition the rule already had; listed first so
  the discrete display timing wins for `display`.)
- Shown rule (`.w__p[data-open="true"]`, wherever §5.1 puts it):
  `display: flex;` (as in §5.1) plus `opacity: 1;`.
- Entry from-state, top-level (inside the same `@media` for a device reveal):
  `@starting-style { .w__p[data-open="true"] { opacity: 0; } }`.

Opacity is the only animated property on purpose: a slide would need a length
the closed scales do not have. Browsers without `@starting-style` ignore the
block and the reveal is the same instant swap §5.1 always emitted. Under
`prefers-reduced-motion` the rule is already covered by the `transition: none`
repetition, so the swap is instant there too.

---

## 6. Sheet API

```go
func For(w widget.Widget) *Sheet   // needs Kind, not just Name — see §1.3, §5, §6.1
func (s *Sheet) Root(opts ...Option) *Sheet
func (s *Sheet) Part(p widget.Part, opts ...Option) *Sheet
func (s *Sheet) Within(container, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) CueWithin(c widget.Cue, container, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) CueWithinHover(c widget.Cue, container, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) On(d css.Device, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) OnlyOn(d css.Device, p widget.Part, opts ...Option) *Sheet
func (s *Sheet) Parts() []widget.Part   // declared parts, sorted — for component tests
func (s *Sheet) StateAttrs() []fmt.KeyValue
func (s *Sheet) Validate() []error
func (s *Sheet) Stylesheet() *css.Stylesheet   // panics if Validate() is non-empty
```

`Rule`, `Triplet` and all `Sheet` fields are unexported. `FlowType` is an enum,
not a string. An empty `Part` argument to `When`/`Cue` targets the root.

`Within(container, p, opts…)` applies `opts` to `p` exactly as `Part()` would
and additionally records that `p` renders **inside** `container` — the part
tree, which the sheet walks to reason about positioning (who is whose
containing block). It reads like the DOM: `Within("menu", "options",
Flyout(SideStart))` is "options, inside menu". `Part()` remains the normal
declaration; `Within()` is only needed where containment changes the result —
a `Flyout()` hanging from an `Anchor()` while a positioned part sits between
them, or a `Flyout()` whose chain passes through a `Scroll()` region — and in
exactly those cases `Validate()` (§6.1) rejects the sheet until the nesting is
declared (or the composition corrected).

### 6.1 Validation conditions

`Validate()` returns **all** matching problems, never just the first.

| Condition | Message shape |
|---|---|
| `When`/`Cue` names a part never declared with `Part()` | `sheet <name>: rule for undeclared part "<part>"` |
| A declared part produces no declarations | `sheet <name>: part "<part>" emits nothing` |
| `Veil()` without `Backdrop()` on the same rule | `sheet <name>: Veil() requires Backdrop()` |
| `When` uses a state `Kind.Allows` rejects | `sheet <name>: state <state> is not meaningful for kind <kind>` |
| `Interactive()` on `Page` or `Inactive` (§3.2) | `sheet <name>: surface <surface> has no interaction states` |
| `On()` names a part never declared with `Part()` | `sheet <name>: device rule for undeclared part "<part>"` |
| An `On()` rule emits nothing | `sheet <name>: device rule for part "<part>" on <device> emits nothing` |
| `Drawer()` and `Backdrop()` on the same rule | `sheet <name>: part "<part>": Drawer() and Backdrop() both set position; use one` |
| `Drawer()` and `Width()` on the same rule | `sheet <name>: part "<part>": Drawer() already sets width; remove Width()` |
| `Drawer()` without `RevealedBy()` in any rule for that part | `sheet <name>: part "<part>": Drawer() without RevealedBy(); the panel would be permanently visible` |
| `OnlyOn` called twice for the same part with different devices | `sheet <name>: part "<part>" declared OnlyOn for both <device1> and <device2>` |
| `Within()` names a container never declared with `Part()` | `sheet <name>: Within() container "<part>" is not a declared part` |
| `Within()` declares a part inside itself | `sheet <name>: part "<part>": Within() declares the part inside itself` |
| `Within()` containment is cyclic | `sheet <name>: part "<part>": Within() containment cycle` |
| A `Flyout()` part is not nested inside anything while a containing-block part exists (`Docked(Parent, …)`, `OnEdge`, `Backdrop(Parent)`) | `sheet <name>: part "<flyout>" is a Flyout and part "<stealer>" is a containing block, but the sheet cannot know which contains which — declare the nesting with Within()` |
| A positioned part sits between a `Flyout()` and its `Anchor()` in the declared chain | `sheet <name>: part "<flyout>": part "<thief>" is positioned between the Flyout and its Anchor and steals the containing block the Flyout hangs from; make it not positioned or re-anchor the Flyout` |
| The declared chain of a `Flyout()` passes through a `Scroll()` part | `sheet <name>: part "<flyout>": part "<scroller>" is a Scroll() region and clips the Flyout inside it; move the panel out of the scroller or use a flow accordion` |

Every message names the sheet and the part. The panic in `Stylesheet()` is the
only signal the author gets, and it surfaces from inside `ssr`'s generated
program — far from the source that caused it.

---

## 7. Emission structure

Exact order of the emitted document:

```
@layer tokens, primitives, widgets, states;

@layer primitives { … }    omitted entirely when empty
@layer widgets    { … }    omitted entirely when empty
@layer states     { … }    omitted entirely when empty

@media (hover: hover) {              only when some rule carries CueWithinHover
  @layer states { … }
}

@media (max-width: 639.98px) {     one block per device that has rules
  @layer widgets { … }
  @layer states  { … }
}
@media (min-width: 640px) and (max-width: 1023.98px) { … }
@media (min-width: 1024px) { … }

@media (prefers-reduced-motion: reduce) { … }   only if some rule carries Animate
```

Within each layer: root rule first, then parts in ascending name order; state
rules ordered by state value then part name; cue rules ordered by cue value then
part name; cue-within rules (both `CueWithin` and `CueWithinHover`) ordered by
cue, container, then part name. Declarations within a rule are sorted; selectors
within a rule are sorted.

`CueWithinHover` emits the same descendant selector as `CueWithin`, but inside
`@media (hover: hover)`: the rule only applies when the primary pointer can
actually hover. Reach for it whenever a hover reveal would misfire on touch —
a tap fires `:hover` and synthetic mouse events, but not the hover capability.

Device blocks appear after `@layer states` and before `prefers-reduced-motion`,
ordered by device ascending (Mobile, Tablet, Desktop), parts ascending within
each. Query strings come from `css.Device.Query()` — never built here.

### 7.1 Device block emission rules

1. The query string is `d.Query()` from `css`. Never build a query string here.
2. Device rules emit into `@layer widgets`, not `@layer primitives`.
3. A flow primitive used inside `On()` emits its full declaration set inline in
   that widgets block, including its child-combinator rules. It does not
   contribute to the top-level primitive buckets.
4. `RevealedBy()` inside `On()` puts `display: none` in the media `@layer widgets`
   block and the `[data-state="true"] { display: <flow> }` rule in the media
   `@layer states` block.
5. Omit an empty `@layer` block entirely, exactly as the top level already does.
6. No `@layer <list>;` statement is emitted inside `@media`.

### 7.2 Global invariants

1. Two emissions of the same sheet are byte-identical.
2. Every `var()` matches a `webtyp/css` token **including its fallback**.
3. No `#`, no `rgba(`, no `!important`, no `vw`/`vh` anywhere in the output.
4. No selector begins with `.fl-` or `.exc-`.
5. No empty `@layer` block.
6. Every selector is `.class`, `.class[data-*="true"]` or `.class:pseudo`.

---

## 8. SSR provider contract

```go
func (w *T) RenderCSS() *css.Stylesheet
```

Called on a zero value (`&T{}`) by `webtyp/ssr`. It must not read fields, and it
must be deterministic. See [ARCHITECTURE.md §8](ARCHITECTURE.md#8-contract-with-webtypssr).

---

## Related documents

- [ARCHITECTURE.md](ARCHITECTURE.md) — structure and constraints.
- [DESIGN.md](DESIGN.md) — why these values and not others.
- [MIGRATION.md](MIGRATION.md) — mapping from the published API.
