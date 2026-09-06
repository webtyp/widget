# Architecture of `webtyp/widget`

Defines the **what** and **why** of visual contracts, anatomy, states, layout, and
the style system. Abstract structure only — exact values, signatures and emission
tables live in [SPECS.md](SPECS.md); the reasoning behind each choice and the
alternatives rejected live in [DESIGN.md](DESIGN.md).

---

## 1. What `webtyp/widget` is

The module governing the **structure** of visual components. It names the pieces a
component is made of, the states it can occupy, and how those pieces are laid
out — independent of data, transport, DOM, or free CSS text.

It exists to make one claim true: **someone who does not know design can build a
correct, accessible component without reading this library's source.** Every
constraint below serves that claim. Where a decision requires design judgement,
the library makes it; where it requires product judgement, the author makes it.

---

## 2. Position in the suite

Three modules divide one problem. The boundary is what keeps each of them small.

| Module | Owns | Never does |
|---|---|---|
| `webtyp/css` | **Values** — what a colour, space, duration or z-level *is*; light/dark switching; contrast guarantees | Know anything about components |
| `webtyp/widget` | **Decisions** — which token applies to which part, in which state | Invent a value |
| `webtyp/ssr` | **Delivery** — collect the sheets actually used, order and deduplicate them | Know what a widget is |

See [diagrams/BOUNDARIES.md](diagrams/BOUNDARIES.md).

`webtyp/css` is a hard dependency and is meant to stay one. This module emits
*references* (`var(--color-primary, …)`); something must declare those variables
and switch them for dark mode, and that is `css`. More importantly, `css` owns the
contrast guarantee across the palette — which is what makes a surface safe to hand
to an author who cannot evaluate contrast themselves. Rationale in
[DESIGN.md §1](DESIGN.md#1-why-webtypcss-stays).

---

## 3. Package boundaries

Two packages with opposing constraints.

**`webtyp.com/widget` (root, WASM-compatible).** Identity, anatomy,
states, browser cues, ARIA kinds. It **travels inside the WASM binary**, so it
carries zero style logic and zero emission.

**`webtyp.com/widget/style` (stylesheet engine, WASM-exempt).** Closed
scales, surfaces, flow primitives, and deterministic CSS emission. Excluded from
WebAssembly by `//go:build !wasm` so it can never reach the client binary. This is
enforced by test, not by convention.

---

## 4. Anatomy and state contract

- **`Name`** identifies a widget and prefixes every class it emits, so two widgets
  cannot collide even if they choose the same part name.
- **`Part`** is a named local slot of the anatomy (Open UI vocabulary): `"row"`,
  `"menu"`, `"header"`. Local to its widget; never prefixed by the author.
- **`Class`** is derived deterministically from a `Name` and a `Part`. It has **no
  public constructor**: the only way to obtain one is to derive it. This is what
  makes markup and stylesheet agree by construction rather than by discipline.
- **`State`** is a state the **widget owns**: written by Go, read by the sheet. It
  maps to a data attribute (`data-selected="true"`). A state is written **only**
  through `dom.BindState`/`BindStateFunc`/`SetState` — `State.Attr()` returns a
  `StateAttr` whose `Key()`/`Value()` exist for `dom` and `widget/style` and for
  nobody else. The accessors exist because the two halves have to meet somewhere,
  but the type is deliberately not an `fmt.KeyValue`: a string pair fits every
  attribute method `dom` exposes, including the ones that write the wrong value
  (`BindAttrBool("data-x", …)` landed as `data-x=""` while the sheet selects
  `data-x="true"` — a silent no-op that shipped once and is now a compile error).
- **`Cue`** is a state the **browser owns**. It can only be styled, never written
  from Go — which is why it is a separate type with no attribute method and no
  writer of any kind.

`State` and `Cue` are distinct types because confusing them is the most common way
to produce a stylesheet that cannot be driven from the application.

---

## 5. Kind is load-bearing

A widget declares its WAI-ARIA pattern once. The library derives from it:

- the **ARIA role** the markup must carry,
- the **stacking level** an overlay of that pattern belongs at,
- which **states are meaningful** for that pattern, used to reject nonsense.

This is the central mechanism for removing decisions from the author. Someone who
declares "this is a Dialog" should not additionally be asked to choose a role, a
z-index, and a set of valid states — those follow from the pattern, and getting
them wrong is exactly what a non-specialist cannot detect.

`Kind` is closed on purpose: a component that fits none of the patterns is almost
always two components.

---

## 6. The style system

### 6.1 Zero escape

The engine accepts no free strings, no viewport units, no arbitrary values, and
emits no `!important`. If a value is not on a closed scale, it cannot be
expressed. The sheet builder is the only public construction path; the rule
structures behind it are private, so there is no second way in.

### 6.2 Closed scales

Every visual property is chosen from a small, typed, semantic scale. Scales are
sized to the underlying token catalog — a scale never offers a step the token
system cannot distinguish, because a step that changes nothing teaches the author
that the library does not work.

### 6.3 Surfaces are complete decisions

A `Surface` is not a colour. It is a whole visual decision — background, text,
border and radius — resolved together, plus its own hover, focus and press
treatments. Padding is deliberately excluded: it varies with what a part
contains, so it stays an explicit choice. An author picks *what a thing is* (`Panel`, `Primary`, `Danger`), never
*what it looks like*, and cannot pair one surface's base with another surface's
hover.

### 6.4 Flow primitives

Layout is expressed with a closed set of primitives: `Stack`, `Row`, `Split`,
`Grid`, `Center`, `FillCentered`, `ScrollRow` and `MediaBox`. Height is never
declared directly; content flow determines it.

They are responsive through **intrinsic sizing**, not through queries: a
primitive reacts to its own available width, so a component behaves identically
in a sidebar and on a page. No emitted rule uses `@container` or `@media`, and
none needs a wrapper element to work — which keeps the engine from coupling to
DOM structure.

The consequence is that there is **no viewport-scoped mechanism at component
level at all**: no `vw`/`vh`, no media queries. Some decisions are legitimately
viewport-scoped — a mobile navigation pattern, a print layout — and the division
is deliberate:

> **A component responds to its own width. The application shell responds to the
> viewport.**

Shell-level decisions belong in application CSS, where `webtyp/css` already
publishes `--bp-sm` through `--bp-xl`. Admitting the escape at component level
would reintroduce exactly the bug intrinsic sizing avoids: a component that
reacts to the viewport is unusable inside a sidebar.

### 6.5 Visibility

Showing and hiding is a single declaration bound to a state, not a pair of
opposing rules the author must place correctly. The engine restores the display
mode the element's own flow requires, so revealing an element never destroys its
layout.

### 6.6 Validation

A sheet is checked before it is emitted, and a malformed sheet is treated as a
programming error rather than a runtime condition. Silent dead CSS — a rule for a
part that does not exist, a modifier with no effect — is the failure mode this
library most needs to prevent, because it is invisible in both the Go source and
the rendered page.

### 6.7 The part tree

CSS resolves a `Flyout`'s `inset-block-start:100%` against the nearest
**positioned** ancestor — which is not always the `Anchor()` the author meant.
A `Docked(Parent, …)` trigger between the two becomes the containing block and
the `Anchor()` is dead code, with nothing in the emitted CSS distinguishing the
broken shape from the documented one. That is a silent failure only the sheet
can see, because parts are held in a flat map with no record of nesting.

`Within(container, part, opts…)` closes the gap: it declares that `part`
renders inside `container` (applying `opts` exactly as `Part()` would) and the
sheet walks the declared chain when validating. Two failure shapes are loud by
default:

- a `Flyout()` with no declared nesting, coexisting with a containing-block
  part (`Docked(Parent, …)`, `OnEdge`, `Backdrop(Parent)`) — ambiguous by
  construction, the author is told to declare the nesting; and
- a positioned part on the declared chain *between* the `Flyout` and its
  `Anchor` — the theft is named, both parts; and
- a `Scroll()` part anywhere on the declared chain — overflow clips the panel
  by DOM ancestry, containing block or not (an `Anchor` above the scroller
  does not save it), so unlike the theft this walk runs to the end of the
  chain; rejected, both parts named.

A chain that ends at the declared container with no `Anchor` above it is the
author's explicit choice (a docked trigger that spans its anchor) and is
accepted. `Part()` remains the normal declaration; `Within()` only matters
where containment changes the result — which is exactly where the sheet refuses
to stay silent without it.

---

## 7. Emission guarantees

1. **Fixed cascade layers**, declared in a stable order: `tokens, primitives,
   widgets, states`. This removes specificity conflicts without `!important`.
2. **Stable output.** Emitting twice yields byte-identical bytes, so version
   control diffs stay meaningful.
3. **Flat specificity.** Every rule is `.class`, `.class[data-state]` or
   `.class:pseudo`. Nothing couples to DOM structure.
4. **No invented values.** Every emitted value is a token reference. This is
   enforced by a drift guard that compares each `var()` against the catalog,
   fallback included.
5. **No unreachable selectors.** The engine emits only selectors that some markup
   can actually carry.

---

## 8. Contract with `webtyp/ssr`

`ssr` compiles a generated program that instantiates a component as a **zero
value** and calls a provider method matched **by name**. For CSS that method is
`RenderCSS()`.

Two obligations follow for every component author, and they are architectural
rather than stylistic:

1. A style builder **must not read fields**. It runs on `&T{}`.
2. A style builder **must be pure and deterministic**.

Sheets are concatenated across components, so anything a single sheet repeats is
repeated once per component in the shipped stylesheet. That is why this module
minimises per-sheet preamble and leaves cross-sheet deduplication to `ssr`, which
is the only layer that can see all sheets at once.

### 8.1 Appearance that depends on component data

A sheet is static per component *type*, so a component whose appearance depends
on its configuration — a table with a variable column count, a chart with N
series — cannot express that in the sheet itself.

The sanctioned route is a custom property written by the component onto the
element, consumed by the sheet through `var()`:

```
markup   <div class="chart" style="--series: 7">
sheet    .chart { grid-template-columns: repeat(var(--series), 1fr) }
```

The sheet stays static and cacheable; the variation stays in the markup, where
the data already is. Parameterising sheets instead would multiply the emitted CSS
by the number of configurations and defeat `ssr`'s content-hash cache.

This is the one place the "no free strings" guarantee is not enforced: that
`style` attribute is written by the component and nothing checks it. Confining it
to **numeric custom properties** keeps the hole small, and it is a hole by
design rather than by oversight.

Distinct from the hole: custom properties **the sheet itself** writes and
consumes — `--gap`, `--rail`, `--max-width`, and the `FloatingChrome` strip
(`--floating-top`/`--floating-bottom`, written on the host, consumed by any
`Scroll()` descendant through `var(…, 0px)`, inherited across widget
boundaries). These are statically emitted, token-derived, drift-guarded, and
carry their own fallbacks at every consumption site, so they are governed by
the same rules as any other declaration — `FloatingChrome` is not an exception
to "no free strings", it is a declaration like any other that happens to be a
custom property.

---

## Related documents

- [SPECS.md](SPECS.md) — exact API surface, scale mappings, and emission tables.
- [DESIGN.md](DESIGN.md) — why each decision was made, and what was rejected.
- [TRADEOFFS.md](TRADEOFFS.md) — what this structure costs, and the limitations it accepts.
- [MIGRATION.md](MIGRATION.md) — moving from the pre-release API.
- [diagrams/BOUNDARIES.md](diagrams/BOUNDARIES.md) — module boundaries.
