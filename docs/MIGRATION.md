# Migration — closed-API release

Mapping from the published API to the one specified in [SPECS.md](SPECS.md). One
breaking release; there is no compatibility period and no aliases. Reasoning for
that choice is in [DESIGN.md §11](DESIGN.md#11-why-one-breaking-release).

---

## 1. Prerequisite: `webtyp/css`

This release depends on a `css` release first. Nothing here compiles until it
lands.

| Addition to `css` | Why |
|---|---|
| `--color-<family>-hover|focus|press` for every surface family | These tokens have been replaced by dynamic, programmatic state derivations (`css.Hover()`, `css.Focus()`, and `css.Press()`) to ensure contrast-tested safety. |
| A real token behind the `Subtle` interaction states | They are literal `rgba(0,0,0,0.05)` washes today, invisible on a dark surface |
| `--column-narrow`, `--column-medium`, `--column-wide` | `Grid` column minimums are hardcoded `rem` values with no token |
| `--max-w-readable`, replacing `--max-w-prose` | `Size.Readable` must mirror the token it emits; leaving the token named `prose` reintroduces the translation step |
| `--color-focus-ring` | The focus ring currently reuses `--color-primary`, which is invisible on the `Primary` surface itself |

---

## 2. Renames

Reasoning for the naming pass is in [DESIGN.md §12](DESIGN.md#12-naming). It is
grouped below by *why* the name changed, because the four groups carry very
different risk: the first two are mechanical, the third is a correctness fix, and
the fourth changes which constant you should be reaching for.

### 2.1 Truncations — names now mirror the token they emit

| Before | After | Emits |
|---|---|---|
| `Opt` | `Option` | — |
| `Space0 … Space12` | `SpaceNone, Space1, Space2, Space3, Space4, Space6, Space8, Space12` | `--space-N` |
| `Track` | `ColumnWidth` | — |
| `TrackSm, TrackMd, TrackLg` | `ColumnNarrow, ColumnMedium, ColumnWide` | `--column-narrow/medium/wide` |

`Radius*` and `Text*` are unchanged: they already mirror `--radius-*` and
`--text-*`. If those token names are themselves unreadable, that is a
`webtyp/css` question — renaming only on the Go side would reintroduce the
translation step this rule removes.

**Spacing is not a 1:1 remap.** The old scale advertised 13 steps and resolved to
6 values, so adjacent constants were identical. Pick by intent:

| Before | Resolved to | Now |
|---|---|---|
| `Space0` | `0` | `SpaceNone` |
| `Space1` | `--space-1` | `Space1` |
| `Space2` | `--space-2` | `Space2` |
| `Space3` | `--space-3` | `Space3` |
| `Space4` | `--space-4` | `Space4` |
| `Space5`, `Space6` | `--space-6` | `Space6` |
| `Space7`, `Space8` | `--space-8` | `Space8` |
| `Space9` … `Space12` | `--space-12` | `Space12` |

The gaps are deliberate: there is no `Space5` because there is no `--space-5`.

### 2.2 Specialist vocabulary — replaced with what the thing does

| Before | After | Why the old name was opaque |
|---|---|---|
| `Reel(gap)` | `ScrollRow(gap)` | *Every Layout* term |
| `Cover()` | `FillCentered()` | *Every Layout* term |
| `Frame(r)` | `MediaBox(a)` | *Every Layout* term; its child rules target `img`/`video` |
| `Scrim()` | `Veil()` | stage-lighting term |
| `Size.Prose` | `Size.Readable` | typography term for a readable line length |
| `Surface.Sunken` | `Surface.Inset` | design term |
| `Flush()` | `EdgeToEdge()` | design term |
| `Clip()` | `HideOverflow()` | states the mechanism instead of a metaphor |
| `Of(name)` | `For(widget)` | a preposition stating no relation |
| `On(surface)` | `As(surface)` | a preposition stating no relation |

`Of` → `For` and `On` → `As` are the weakest entries; `On` is also the most-typed
function in the API, so this is the largest single source of churn. See
[DESIGN.md §12.7](DESIGN.md#127-the-two-closest-calls).

### 2.3 Names that read as something else — mandatory

| Before | After | Why |
|---|---|---|
| `Fixed()` | `KeepSize()` | It emits `flex-shrink:0; flex-grow:0`, but `position:fixed` means something entirely different. A reader who knows CSS reads the old name **exactly wrong**. |

Also normalised for grammar, so every option reads as an instruction:
`Scrolls()` → `Scroll()`.

### 2.4 Surfaces — renamed for role, not appearance

| Before | After | Why |
|---|---|---|
| `Selected` | `Highlight` | collided with `widget.Selected`, a `State` |
| `Disabled` | `Inactive` | collided with `widget.Disabled`; also mirrors `--color-disabled` |
| `Muted` | `Subtle` | `Muted` and `Dimmed` both meant "faded"; nothing said which was which |
| `Accent` | `Primary` | it resolves to `--color-primary`; the old name added a translation step |

The collision this ends produced lines like
`When(widget.Selected, "item", style.On(style.Selected))`, where the two
`Selected`s were different types meaning different things.

### 2.5 Other renames

| Before | After |
|---|---|
| `Ratio` in `Split` | `SplitRatio` — `SplitHalf, SplitTwoThirds, SplitThreeQuarters` |
| `Ratio` in `Frame` | `Aspect` — `AspectSquare, Aspect3x2, Aspect4x3, Aspect16x9` |
| `Elevation.Overlay` | `Elevation.Popover` |
| `Text(TextSm)` | `FontSize(TextSm)` |

**`Frame` ratios did not mean what they said:**

| Before | Emitted | Now |
|---|---|---|
| `Frame(RatioHalf)` | `aspect-ratio: 1/1` (a square) | `MediaBox(AspectSquare)` |
| `Frame(RatioTwoThirds)` | `aspect-ratio: 3/2` | `MediaBox(Aspect3x2)` |
| `Frame(RatioThreeQuarters)` | `aspect-ratio: 4/3` | `MediaBox(Aspect4x3)` |

If the intent was a wide media box, `Aspect16x9` is new and probably what was
wanted.

---

## 3. Collapsed calls

| Before | After |
|---|---|
| `On(X)` + `Cue(Hover, p, On(XHover))` + `Cue(Focus, …)` + `Cue(Press, …)` | `Interactive(X)` |
| `On(Panel)` + `Round(RadiusMd)` | `As(Panel)` |
| `Hidden()` + `When(st, p, Shown())` | `RevealedBy(st)` |
| `Center()` + `Width(s)` | `Center(s)` |
| `Center()` alone | `Center(Readable)` |

`Round` and `Raise` still exist as overrides for the case that genuinely differs
from the surface default. `Pad` is **not** folded into surfaces and stays an
explicit choice on every part that needs it — padding follows from what a part
contains, not from what surface it wears.

---

## 4. Deletions

| Removed | Replacement |
|---|---|
| `Hidden()`, `Shown()` | `RevealedBy(state)` |
| `Above()` | derived from `Kind` — see below |
| The 27 `*Hover` / `*Focus` / `*Press` surface constants | private; use `Interactive()` |
| The 18 exported `Color*Hover/Focus/Press` tokens | moved to `webtyp/css` |
| `Rule`, `Triplet`, and all exported `Sheet` fields | `For()` + options only |

### 4.1 Stacking

`Backdrop(Scope)` no longer emits a hardcoded `z-index: 100`, and `Above()` is
gone. The level comes from the widget's `Kind`.

Consumers that relied on the old integers should check their result: a `Dialog`
backdrop previously sat at `100` — the same level as a dropdown, and *below* a
sticky element at `--z-sticky: 200`. Anything that compensated for a sticky
element covering a modal can drop the workaround.

### 4.2 Direct `Rule` construction

Code of this shape no longer compiles, by design:

```go
sh.PartRules["p"] = style.Rule{HasFlow: true, FlowType: "wobble", HasPad: true, Pad: style.Space2}
```

Rebuild it with `For()` and options.

---

## 5. Behaviour changes without an API change

These compile unchanged and behave differently. Check them.

| Change | Effect |
|---|---|
| Revealing an element restores its flow's `display` | A `Row` or `Grid` that previously became a block when opened now stays a row or grid. **Any CSS written to compensate must be removed.** |
| Focus cue emits `:focus-visible` | The focus ring no longer appears on mouse click. Custom rules that hid it are now unnecessary and may suppress keyboard focus. |
| An invalid sheet panics | A rule pointing at a misspelled part used to emit dead CSS silently; it now fails loudly at emission. Expect to find pre-existing typos. |
| Surfaces carry a radius | Parts that relied on a surface having none now have one. Use `Round(RadiusNone)` where square is intended. Padding is unaffected — it was never folded in. |
| `Split` becomes responsive | It never collapsed before: `container-type` was set on the same selector the `@container` rule targeted, so the rule never applied. It now stacks below ~40rem of its own width. **Layouts tuned around a split that never collapsed will change.** |
| `.fl-*` / `.exc-*` selectors are no longer emitted | Unreachable before — `Class` has no public constructor — so no markup can be affected. Any hand-written CSS targeting them stops matching. |

---

## 6. Full example

Before:

```go
style.Of(m.WidgetName()).
    Root(style.Grid(style.TrackSm, style.Space2), style.On(style.Page), style.Scrolls()).
    Part("master", style.Stack(style.Space1), style.On(style.Panel), style.Round(style.RadiusMd)).
    Part("detail", style.Stack(style.Space2), style.On(style.Panel), style.Round(style.RadiusMd)).
    Part("item",   style.Row(style.Space1),   style.On(style.Muted)).
    When(widget.Selected, "item", style.On(style.Selected)).
    Cue(widget.Hover,     "item", style.On(style.MutedHover))
```

After — 13 options down to 12, and every remaining name says what it does:

```go
style.For(m).
    Root(style.Grid(style.ColumnNarrow, style.Space2), style.As(style.Page), style.Scroll()).
    Part("master", style.Stack(style.Space1), style.As(style.Panel), style.Pad(style.Space3)).
    Part("detail", style.Stack(style.Space2), style.As(style.Panel), style.Pad(style.Space3)).
    Part("item",   style.Row(style.Space1),   style.Interactive(style.Subtle)).
    When(widget.Selected, "item", style.As(style.Highlight))
```

---

## 7. Known consumers

| Module | Impact |
|---|---|
| `webtyp/form` | `NameField` and its parts are unchanged. Any `RenderCSS()` needs rewriting. |
| `webtyp/components/fieldset` | Same. This is the global form skin, so it exercises surfaces and states heavily. |
| `webtyp/layout` | Verify before starting whether it overlaps the flow primitives here. If it does, resolve the overlap in this release rather than after it. |

---

## 8. Upgrading to v0.5.0 — one typed way to write a state

v0.5.0 is breaking in exactly one place: `State.Attr()` stopped returning an
`fmt.KeyValue` and returns a `StateAttr` instead, and `webtyp/dom` gained the
only sanctioned writers. Everything else is additive. Requirement: upgrade to the
matching `dom` release (`BindState`/`BindStateFunc`/`SetState`) **before** this
release — nothing here compiles against an older `dom`.

### The mapping

Every hand-wired state write collapses to one typed call. The six-line idiom is
gone; `attr.Key`/`attr.Value` no longer exist because there is no pair to pull
apart.

| Before | After |
|---|---|
| `attr := widget.Open.Attr()`<br/>`el.BindAttrFunc(attr.Key, func() string {`<br/>&nbsp;&nbsp;`if on { return attr.Value }`<br/>&nbsp;&nbsp;`return ""`<br/>`})` | `el.BindStateFunc(widget.Open, func() bool { return on })` |
| `el.BindAttrFunc(st.Attr().Key, func() string { … })` | `el.BindStateFunc(st, func() bool { … })` |
| `el.BindAttrBool("data-selected", sig)` | `el.BindState(widget.Selected, sig)` — **this is the bug fix** |
| `el.Attr(st.Attr().Key, st.Attr().Value)` | `el.SetState(st)` (or `BindState`, if reactive) |
| `form`'s package-level `attrInvalid`/`attrLocked` vars (`states.go`, deleted) | `BindStateFunc(widget.Invalid, …)`, `BindStateFunc(widget.Locked, …)` |

### What still compiles is now wrong on purpose

`Key()`/`Value()` exist so `dom` (writing) and `widget/style` (selecting) can meet.
Any other call site that reads them to hand-wire a state is now the wrong path and
it costs more to write than the right one. Reviews should `grep -rn 'BindAttrBool("data-'`
and find nothing.

### The overlap that stays

`widget.Disabled` is `data-disabled="true"`; the HTML `disabled` attribute is
written by `BindAttrBool("disabled", …)` and is a different thing. They are not
confusable anymore: one accepts a `StateAttr`, the other a bare string.

### New primitives (additive)

| What | API |
|---|---|
| Viewport-height flex column | `Cover()` |
| Fixed rail beside fluid content | `Sidebar(side, width, gap)` |
| Edge-anchored overlay panel | `Drawer(side, size)` |
| Device-scoped override | `On(css.Device, "part", …)` |
| Device-only part | `OnlyOn(css.Device, "part", …)` |

### New Sheet methods

`StateAttrs()` returns every `[data-*="true"]` pair the emitted CSS selects on.
Any consumer using `RevealedBy()` should add a test that renders their markup and
asserts every returned pair appears in it.

### Before/after: application shell

Before (hypothetical — the old API had no cover or sidebar):

```go
style.For(app).
    Root(style.Stack(style.SpaceNone), style.As(style.Page)).
    Part("header", style.KeepSize(), style.As(style.Panel)).
    Part("body", style.Fill()).
    Part("rail", style.Width(style.Content), style.As(style.Panel)).
    Part("content", style.Fill())
```

After:

```go
style.For(app).
    Root(style.Cover(), style.As(style.Page)).
    Part("header", style.KeepSize(), style.As(style.Panel)).
    Part("body", style.Fill()).
    Part("rail", style.KeepSize(), style.As(style.Panel)).
    Part("content", style.Fill())
```

Mobile nav rail that becomes a drawer:

```go
style.For(nav).
    Root(style.As(style.Page)).
    Part("rail", style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone), style.As(style.Panel)).
    On(css.Mobile, "rail", style.Drawer(style.SideEnd, style.TwoThirds), style.RevealedBy(widget.Open))
```

---

## 9. Upgrading to v0.6.0 — `SlideDeck`, declared stacking, ring state borders, floating chrome, the part tree

v0.6.0 removes one flow primitive and changes several emission behaviours: every
out-of-flow element declares its stacking level, the `OnEdge` straddle stops
using `transform`, state borders are painted as shadow rings, a new
`FloatingChrome` contract reserves the edge strip of a scroll container, and
the sheet learns the part tree so a `Flyout` hanging from the wrong containing
block is diagnosed instead of shipped silently.

### 9.1 `Deck(gap)` → `SlideDeck(m Motion)`

`Deck` was a horizontal scroll-snap strip; a nested `MasterDetail` in the same
shell chained the scroll and the application changed section on its own. The
replacement stacks its children as absolute layers and shows the one carrying
`widget.Current`; there is no scroller.

| Before | After |
|---|---|
| `Part("stage", style.Deck(style.SpaceNone), …)` | `Part("stage", style.SlideDeck(style.MotionBase), …)` |

The markup changes too, because the mechanism changed:

| Before | After |
|---|---|
| `ScrollIntoView` on the target panel | write `widget.Current` on the panel that should be on screen (one `BindState(widget.Current, …)` per panel) |

All panels stay mounted; nothing is unmounted on change. The `On()`/`OnlyOn()`
device-scoped forms emit the same three rules inside their query.

Reasoning: [DESIGN.md §19](DESIGN.md#19-why-deck-was-replaced-by-slidedeck).

### 9.2 State rules repaint governed borders as a shadow ring

A state rule (`When`, `Cue`, `CueWithin`) that carries a bordered surface
(`Panel`, `Inset`) emits `box-shadow: 0 0 0 1px <color>` — static and enhanced
halves as a double declaration, composed with `Raise()`'s elevation in one
declaration (ring first) when the state rule also raises — instead of
`border: 1px solid …`. The box no longer grows by 2px when the pointer enters
it. Base rules are unchanged.

**If a state rule's border was load-bearing** — e.g. CSS that measured or aligned
against the grown box — it stops growing, which is the intended correction.

The ring is not an `outline` (as 9.2 in earlier drafts specified): outlines
paint at the end of the stacking context, over the element's positioned
descendants, and Safari < 16.4 ignores `border-radius` on them.

Reasoning: [DESIGN.md §18](DESIGN.md#18-why-a-state-never-changes-the-box-size).

### 9.3 Out-of-flow elements declare their stacking level

`Backdrop(Parent)` and `Docked(Parent)` now emit `z-index: 1` (previously no
`z-index` at all). They declare the **local** level — enough to order the element
against its unpositioned siblings — while still never claiming the overlay layer
their `Kind` owns. CSS that relied on these two primitives being `z-index: auto`
for `z-index: 0`-layering tricks must switch to explicit ordering.

`OnEdge` also emits `z-index: 1` and drops its `transform: translateY(±50%)` in
favour of `margin-block-{start|end}: calc(-0.5 * var(--chip-height,1.25rem))`:
a transform is invisible to `scrollHeight` by spec, so no ancestor could reserve
the space the chip really occupies, and it creates an implicit stacking context.
The straddle is now exact because the chip's height is the shared `--chip-height`
token. CSS that selected on the transform (e.g. `transform: translateY(-50%)`
overrides) must target the margin instead.

Reasoning: [SPECS.md §1.3](SPECS.md#13-kind--role-stacking-level-allowed-states).

### 9.4 `FloatingChrome(edge, size, gap)` — the scroll reservation contract

A floating element that occupies a strip along its edge (a FAB, a miniplayer)
declares it with `FloatingChrome(EdgeBottom, size, gap)`; the emitted sheet sets
`--floating-bottom: calc(<size> + 2 * <gap>)` on its box. Every `Scroll()` region
— in this widget or a descendant one — now reserves that strip by padding itself
with `padding-block-start/end: var(--floating-top|bottom, 0px)`. Without
`FloatingChrome` the reservation is `0px` and nothing changes, so the new
`Scroll()` padding is behaviourally invisible except where the new option is
used.

`size` is an `IconSize`, not a `Size`: floating chrome pinned to a screen edge
is by construction a small icon-only control. A `Size`'s percentages would
resolve against the scroll region's own inline size and `max-content` is not a
`<length>`, so neither can compile into the calc. Pass the same `IconBox(...)`
step the glyph already has — `FloatingChrome(EdgeBottom, IconLg, Space4)`
emits `--floating-bottom: calc(2.5em + 2 * <gap>)`. **If a sheet passed a
`Size` constant** (e.g. `Readable`), it no longer compiles: replace it with the
control's `IconSize` step.

**If a `Scroll()` region's last items were load-bearing** — e.g. a list whose
final row was expected flush against the container's end — the new `0px`-default
padding keeps the geometry identical.

### 9.5 The prerequisite `--chip-height` token

`OnEdge`'s straddle consumes `--chip-height` from `webtyp/css` (released as
`css.ChipHeight` in `css` v0.4.10). The `css` dependency in `go.mod` must be at
least that version.

### 9.6 `Within(container, part, opts)` — the sheet learns the part tree

A `Flyout` resolves its `inset-block-start:100%` against the nearest
**positioned** ancestor, not against the `Anchor()` the author declared. A
`Docked(Parent, …)` part between the two becomes the containing block, the
`Anchor()` is dead code, and the dropdown hangs from the wrong box — the shape
that shipped broken in `targetlist`. The library now refuses to stay silent
about it:

- A `Flyout` part with **no declared nesting**, coexisting with a
  containing-block part (`Docked(Parent, …)`, `OnEdge`, `Backdrop(Parent)`), is
  **ambiguous**: `Validate()` reports it and tells the author to declare the
  nesting with `Within(container, part, opts)` — which applies `opts` to the
  part exactly as `Part()` would and adds the containment relation.
- With the nesting declared, a positioned part **between** the `Flyout` and its
  `Anchor()` in the declared chain is rejected with a precise error naming both
  parts. A chain that ends at the declared container with no `Anchor` above it
  (a docked trigger that spans its anchor) is the author's explicit choice and
  stays valid.

**If a sheet combines `Anchor()` + `Docked(Parent, …)` + `Flyout()`** (the
`targetlist` shape), `Validate()` now fails: declare the containment with
`Within("menu", "options", style.Flyout(...))`, then resolve the theft by
un-positioning the part in between, moving the `Flyout` out of the docked part,
or making the docked trigger span its anchor. The resolution chosen for
`webtyp/components` ships in its own release.

### 9.7 A `Flyout` inside a `Scroll()` region is rejected

An overlay anchored inside a scrolling list is **clipped by the scroller** —
overflow clips every descendant that escapes the padding box, containing block
or not. The shape measured 10px visible of a 84.8px panel in `targetlist`
before that component moved its options to an accordion. The library now
rejects it: a `Flyout()` whose declared chain (`Within()`) passes through a
`Scroll()` part fails `Validate()` with both parts named.

**If a sheet declares `Part(list, Scroll())` + `Within(list, "panel",
style.Flyout(...))`**, `Validate()` now fails. Choose deliberately: an
accordion in flow inside the row (what `targetlist` adopted), or move the panel
out of the scroller (what `usermenu` does today). A `Scroll()` part that is
NOT on the Flyout's declared chain is unaffected.

### 9.8 `Interactive(s)` no longer paints — it declares the family only

`Interactive(s)` used to write the resting surface AND derive the states from
it, racing `As()` for the same field: the last call won silently. It now
writes only the interaction family; `As()` alone paints the resting look.

| Before | After |
|---|---|
| `Interactive(X)` (resting X + states from X) | `As(X)` + `Interactive(X)` |
| `Interactive(X)` + `As(Y)` (resting Y by accident of order, family lost) | `As(Y)` + `Interactive(X)` — resting Y, family X, both kept |
| `Button(s)` | unchanged — it paints `s` and derives from `s` |

**Every `Interactive(X)` without an `As` on the same rule loses its resting
background.** The compiler does not catch it — the part simply goes
transparent — so each migrated rule needs a visual check, not only a green
suite. `Subtle` resting text additionally repaints to on-surface inside the
states (AA 9.68:1, was 1.42:1); `familyBase(Subtle)` is the neutral surface
token, never the muted text token.

---

## Related documents

- [SPECS.md](SPECS.md) — the target API in full.
- [DESIGN.md](DESIGN.md) — why each change was made.
- [ARCHITECTURE.md](ARCHITECTURE.md) — the structure it produces.
