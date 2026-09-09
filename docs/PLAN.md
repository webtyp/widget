---
PLAN: "feat(style): Button — the one recipe every button in the ecosystem uses"
EXECUTOR: local
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **This is a GATE.** `webtyp/components` cannot land its button migration
> until this ships. Orchestrator:
> [webtyp/docs/BUTTON_SYSTEM_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/BUTTON_SYSTEM_MASTER_PLAN.md).

# Plan — `style.Button(Surface)`: one place decides what a button is

## 1. The defect this closes

`webtyp.com/widget/style` exposes a vocabulary of **box recipes** —
`ControlBox()`, `IconBox()`, `ChipBox()`, `LogoBox()` — plus a paint recipe,
`Interactive(Surface)`. There is **no recipe for a button**. So every component
that renders one re-composes it by hand, and they do not agree:

| Component | Its recipe for a button |
|---|---|
| `components/actionbutton` | `Pad(Space2)` + `Round(RadiusSm)` + `As(Page)` + `Interactive(s)` |
| `components/scheduleeditor` | `ControlBox()` + `Interactive(s)` + `Round(RadiusSm)` + `KeepSize()` |
| `components/themetoggle` | `Interactive(Primary)` on the root |
| `components/usermenu` | `Interactive(Subtle)` on the trigger |

Measured in the running demo (`app-demo`, `#agenda`, viewport 888×588, via the
MCP browser tools):

- `.scheduleeditor__row-add` — the "Add row" button — renders **799.16 px wide**,
  the full width of its panel, as a gradient bar. It carries `KeepSize()`,
  which emits `flex-shrink: 0; flex-grow: 0`. Its parent is `Stack` — a flex
  **column** — where width is the **cross** axis. `flex-grow`/`flex-shrink`
  do not govern the cross axis; `align-items: stretch` does, and it is the
  default. **`KeepSize()` was never able to stop this**, and no other Option
  in the package can either: `grep -rn 'align-self' style/` returns **nothing**.
- `.scheduleeditor__row-remove` — "Remove row" — overflows its own box: the
  recipe pins `min-width` through `ControlBox()` and nothing adds inline
  padding for the label.

So the consumer cannot fix either defect from `components`. The vocabulary is
missing a member, and that is a defect in this library — not in its consumers.

**Not in scope, and explicitly not a defect:** `ControlBox()` sizing a native
checkbox to 44×50 is **correct and deliberate**. `css.ControlWidth`'s own
comment says so: *"The minimum width every interactive control shares — a
checkbox, a radio — so its tap target meets the touch-size floor on BOTH axes."*
Do **not** "fix" that, and do not touch `ControlBox()`, `ControlHeight` or
`ControlWidth`. A previous stage of this project accepted "0 tap targets under
44×44" as a shipped criterion; shrinking those would regress it.

## 2. Design gate

### 2.1 Prior art

Three established systems, and what each does with "what is a button":

- **Bootstrap** — `.btn` carries the box (inline-block, padding, line-height,
  border-radius, transition); `.btn-primary` only paints. Two classes, but the
  box is written **once** in `_buttons.scss`, and a consumer never restates
  padding.
- **Tailwind / DaisyUI** — `btn btn-primary`. DaisyUI exists precisely because
  hand-composing `px-4 py-2 rounded inline-flex items-center shrink-0` at every
  call site does not stay consistent. The component layer collapses it to one
  token.
- **Material UI** — `<Button variant="contained" color="primary">`; the recipe
  lives in the component and the single global adjustment point is
  `theme.components.MuiButton`. Ant Design and Chakra do the same through their
  theme's component recipes.

All three agree on the property we lack: **the button recipe lives in exactly
one place, and the consumer names it rather than rebuilding it.** None of them
asks each component to re-declare padding and height.

**Why we differ in form.** There is no global stylesheet and no utility classes
here: CSS is emitted per widget from its own `Sheet`, so the recipe cannot be a
shared class like `.btn` — a class emitted under `.actionbutton__primary` can
never paint `.scheduleeditor__row-add`. It has to be an **Option** that each
Sheet applies to its own Part. That is Bootstrap's split (box recipe + surface)
expressed as one typed call, which is also why the surface stays a parameter
instead of becoming three separate functions.

### 2.2 Novice-name test

`style.Button(style.Primary)` reads aloud as *"style: button, primary"*. A
junior with no context reads that call and knows what it makes. `Button` is the
word Bootstrap, MUI, Tailwind, Chakra and Ant all use for this; inventing
anything else costs a lookup for zero gain.

**Rejected alternative: `ButtonBox()`.** It keeps the `*Box` family's symmetry —
but the `*Box` members are boxes *only*, so `ButtonBox()` would still need
`Interactive(s)` beside it on every call, leaving two lines and the standing
possibility of writing one without the other. The whole point is that a button
cannot be assembled wrong. `Button(Surface)` is one call that cannot be
half-applied.

### 2.3 Complexity ledger

| Ledger row | Δ |
|---|---|
| Concepts the developer must learn | **+1 / −4** → net **−3**. Learn `Button`; stop having to know that a button is `ControlBox` + `Interactive` + `Round` + `KeepSize` in that combination. |
| Files touched to restyle every button in the app | **1, was 9** → **−8** |
| Lines at the call site | **1, was 3–4** → **−2/−3** per site |
| Ways to do the same thing | **−1**. Two live recipes today (`actionbutton`'s `Pad`-based one and `scheduleeditor`'s `ControlBox`-based one); one after the migration. |
| Exported surface | **+1** |

### 2.4 Where it belongs

`webtyp.com/widget/style`, in the box-recipe family that already lives in
`except.go` next to `ControlBox`, `ChipBox`, `LogoBox`, `KeepSize`.

- **Not `webtyp.com/css`.** That package owns *values*, and the values already
  exist and are already global: `--control-height`, `--space-3`, `--radius-sm`.
  Nothing there is wrong. What is missing is the *recipe* that composes them.
- **Not `components/actionbutton`.** Its stylesheet is emitted under its own
  widget name, so it can only ever paint `.actionbutton__*`. `widget` sits below
  both `components` and `layout`, which is what makes it the only place a single
  definition can reach every button.

### 2.5 What this change deletes

Nothing in `widget` — this is genuinely new capability, and the plan says so
rather than pretending otherwise. The deletions land in `components` (the
hand-rolled per-component recipes) under its own plan; this library only makes
them deletable.

## 3. Stage 1 — the rule flag

**File: [`style/sheet.go`](../style/sheet.go).**

The `rule` struct declares the box flags together around line 42:

```go
	controlBox    bool
	logoBox       bool
	chipBox       bool
```

Add one line to that group, keeping the existing alignment:

```go
	buttonBox     bool
```

`Button` reuses the surface fields that `Interactive` already sets
(`hasSurface`, `surface`, `interactive`) — a second surface field would be a
second way to say the same thing.

That reuse is exactly why §6's diagnostic needs one more flag. Both options
write the same three fields, so after the fact a rule cannot tell whether the
author called `Interactive`, `Button`, or both — the second call just
overwrites the first. Add, in the same group:

```go
	hasInteractive bool
```

and have **`Interactive` alone** set it (`style/surface.go`), following the
`hasRound` / `hasGlyph` convention already in this struct: the `has…` flag
records that the author wrote the option, while the unprefixed field governs
emission. `Button` must NOT set it — that asymmetry is what makes
`buttonBox && hasInteractive` mean "both were called".

## 4. Stage 2 — the Option

**File: [`style/except.go`](../style/except.go).** Place it immediately after
`ControlBox()` so the box-recipe family stays together.

```go
// Button is the one recipe for a button: the shared control height, inline
// padding for its label, and a box that neither stretches to its container nor
// shrinks under pressure. s paints it — Primary, Secondary, Danger, Subtle —
// with the hover, focus and press treatments Interactive derives, and with that
// surface's default radius.
//
// It exists because the parts are not safely composable by hand. KeepSize()
// looks like the way to stop a button filling its panel and is not: it emits
// flex-shrink/flex-grow, which govern the MAIN axis, while a button inside a
// Stack is stretched across the CROSS axis by align-items: stretch. Every
// component that tried composing its own button got a different answer, and
// one of them shipped an 800px-wide "Add row" bar.
//
// Use it for anything the user presses: a button, a summary that acts as one.
// Not for an interactive surface that is not a button — a clickable table row
// or a calendar day keeps Interactive(), which paints without claiming a
// control's box.
func Button(s Surface) Option {
	return func(r *rule) {
		r.hasSurface, r.surface, r.interactive = true, s, true
		r.buttonBox = true
	}
}
```

## 5. Stage 3 — the emission

**File: [`style/emit_place.go`](../style/emit_place.go).**

`placementDecls` emits the self-alignment flags in a fixed order, and its doc
comment states that the sequence is part of the byte-identical contract. Insert
the new block **immediately after** the existing `controlBox` block (around
line 89) so the order stays deterministic:

```go
	if r.buttonBox {
		decls = append(decls, "min-height: "+css.ControlHeight.Var()+";")
		decls = append(decls, "padding-inline: "+css.Space3.Var()+";")
		decls = append(decls, "align-self: center;")
		decls = append(decls, "flex-shrink: 0;")
		decls = append(decls, "flex-grow: 0;")
	}
```

Then extend the `placementDecls` doc comment's list of self-alignment flags to
name `Button` alongside `ChipBox, ControlBox`.

Five declarations, and each one answers a measured defect:

| Declaration | Why |
|---|---|
| `min-height: var(--control-height)` | the shared rhythm; the same floor a list row and a form field stand on |
| `padding-inline: var(--space-3)` | the label stops touching the edges — the "Remove row" overflow |
| `align-self: center` | **the 799px fix**: a flex item with `align-self` other than `stretch` takes its content width on the cross axis |
| `flex-shrink: 0` / `flex-grow: 0` | it does not collapse or bloat on the main axis of a `Row` either |

Emit **no `display`**: a native `<button>` already centres its own label, and
declaring `inline-flex` here would race the flow layer's `display` when a rule
also carries `Row()` or `CenterContent()`.

Emit **no `border-radius`**: `emit_surface.go:68` already gives a surface its
`defaultRadius()` when the rule carries no explicit `Round()`, and
`Primary`/`Secondary`/`Danger`/`Subtle` all resolve to `RadiusSm`. Adding one
here would be the second way to say it.

### 5.1 The exists check

**File: [`style/emit_decls.go`](../style/emit_decls.go), line ~227.** A
predicate lists every flag that makes a rule non-empty, as one long
conjunction of `!r.…` terms including `!r.controlBox && !r.logoBox &&
!r.chipBox`. Add `&& !r.buttonBox` to it. Miss this and a Part whose only
Option is `Button()` is treated as empty and emits nothing.

## 6. Stage 4 — composition validation

**File: [`style/validate_composition.go`](../style/validate_composition.go).**

Two combinations are contradictions, and this package's job is to make them
loud rather than let the last declaration win silently. Add a check alongside
the existing `checkPosition`, following its shape exactly (same
`fmt.Errf` style, same `sheet %s: part %q: …` prefix):

- `Button()` with `Interactive()` — both set the surface; the second call
  silently overwrites the first's argument. Message, verbatim:

  `sheet %s: part %q: Button already paints an interactive surface; drop Interactive`

- `Button()` with `ControlBox()` — both set `min-height` from the same token;
  the duplicate declaration is dead weight that reads as if it did something.
  Message, verbatim:

  `sheet %s: part %q: Button already carries the control height; drop ControlBox`

`Button()` with `KeepSize()` is **not** an error — it is merely redundant, and
this plan does not add a diagnostic for redundancy.

## 7. Stage 5 — the tests

**File: [`style/button_test.go`](../style/button_test.go)** (new).

Package `style`, standard `testing` only, following the shape of the existing
`interactive_test.go` and `decls_test.go`. Four cases, and the first one is the
consumer-shaped proof the publication rule demands — it goes through a real
`Sheet` for a real widget and asserts on emitted CSS text, not on struct fields:

1. **`TestButton_DoesNotStretchInAStack`** — build a sheet with a `Stack` parent
   part and a child part carrying `Button(Primary)`; assert the emitted CSS for
   that part contains `align-self: center;`. This is the regression guard for
   the 799px bar: without it the button fills its panel and no test notices.
2. **`TestButton_CarriesControlHeightAndInlinePadding`** — assert the emitted
   rule contains `min-height: var(--control-height);` and
   `padding-inline: var(--space-3);`.
3. **`TestButton_PaintsLikeInteractive`** — a part with `Button(Primary)` and a
   part with `Interactive(Primary)` emit the same surface, hover, focus and
   press declarations. Compare the surface-derived substrings; do not assert
   the two rules are byte-identical, because only one carries the box.
4. **`TestButton_RejectsRedundantComposition`** — `Button(Primary)` +
   `Interactive(Secondary)` on one part, and `Button(Primary)` + `ControlBox()`
   on another, each make `Sheet.Validate()` return the exact message from §6.

Do not add a fifth test asserting the emitted byte order; `parsesafe_test.go`
and the existing determinism tests already own that contract.

## 8. Constraints

- `gotest` — never `go test`.
- No `TODO`, no `FIXME`, no deprecated path, no commented-out block. Before
  closing: `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' style/` and
  confirm every hit predates this change.
- No standard library: this package already uses `webtyp.com/fmt`; keep it.
- Touch only the six files this plan names. Do **not** migrate any consumer —
  `components` has its own plan, and this repo has no consumers of `Button` yet
  by design.
- Do **not** modify `ControlBox`, `ChipBox`, `IconBox`, `LogoBox`, `KeepSize`,
  `Interactive`, or any token in `webtyp.com/css`. They are correct.

## 9. Acceptance criteria

| # | Check | Expected |
|---|-------|----------|
| 1 | `grep -n 'func Button(s Surface) Option' style/except.go` | one hit |
| 2 | `grep -n 'buttonBox' style/sheet.go style/except.go style/emit_place.go style/emit_decls.go` | four files, all present |
| 3 | `grep -c 'align-self: ' style/emit_place.go` | **1** — the only `align-self` declaration in the package |
| 4 | `gotest ./...` | green, including the four new cases |
| 5 | `go vet ./... && gofmt -l .` | clean |
| 6 | `grep -rn 'TODO\|FIXME\|Deprecated' --include='*.go' style/` | no new hits |
| 7 | `grep -rn 'ControlHeight\|ControlWidth' style/emit_place.go` | `ControlBox`'s block unchanged; `Button`'s adds only `ControlHeight` |

## 10. Stages table

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | Rule flag | `style/sheet.go` | `buttonBox bool` in the box-flag group |
| 2 | The Option | `style/except.go` | `Button(Surface)` exported, documented |
| 3 | Emission | `style/emit_place.go`, `style/emit_decls.go` | five declarations emitted; exists-check updated |
| 4 | Validation | `style/validate_composition.go` | both contradictions rejected with the verbatim messages |
| 5 | Tests | `style/button_test.go` | four cases, `gotest` green |
