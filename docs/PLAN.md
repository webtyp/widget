---
PLAN: "feat(style): Heading recipe — the one section title of the app"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 2279594303412389757
PR: https://github.com/webtyp/widget/pull/8
REVIEWED_COMMIT: 80282cc49691a150edb5917556aac54245912a73
---

> Read AGENTS.md at the repo root before any change.

# Plan — `style.Heading()`

## Problem

Every component sizes its section titles by hand (`FontSize(TextLg)` + `FontWeight(WeightBold)` in one
place, `TextBase` + `WeightMedium` in another), so the same kind of title looks different on every
screen. The app is moving to exactly two title levels: the module name in the `platformd` header
(owned by that layout) and **one** section title inside the content (accordion step, card header,
modal title, final panel). The second one needs a single recipe.

## Design gate

1. **Prior art.** Material 3 type roles (`titleMedium` used by every component), Shopify Polaris
   `Text variant="headingMd"`, GOV.UK `govuk-heading-m`: the component asks for a *role*, the design
   system owns size and weight. Same here: `Heading()` is the role.
2. **Novice-name.** `style.Heading()` — "this part is a heading". No parameter: there is one level
   inside content by decision; a `level` argument would invite a second one.
3. **Ledger.** Concepts +1 (`Heading`) / −0. Lines at call site −1 (two options → one). Ways to style
   a section title: N hand compositions → 1 recipe.
4. **Where it belongs.** `widget/style`, beside `Button`, `AppBar`, `LogoBox`: recipes live here; a
   recipe a component lacks is a defect here (skill widget-styling).
5. **What it deletes.** Nothing in this repo; consumers (`components`, `layout`) replace their
   hand-composed title parts in their own plans.

## Stages

### Stage 1 — the recipe

File `style/heading.go` (`//go:build !wasm`, package `style`):

```go
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
```

- Do not add fields to `rule`; the recipe composes the two existing options.
- Composing `Heading()` with `FontSize(...)` or `FontWeight(...)` on the same part must be rejected by
  the existing composition validation if such a rule exists for duplicate setters; if none exists, do
  **not** add one in this plan.

### Stage 2 — no new test

`Heading()` only composes two existing options; the repo's stylesheet validation already covers it. Do not add a test that asserts emitted declarations.

### Stage 3 — docs

- `README.md` / `docs/ARCHITECTURE.md`: add `Heading()` to the list of recipes with one line:
  "the one section title inside content; the screen title is the platform header's".

| Stage | Files |
|---|---|
| 1 | `style/heading.go` |
| 3 | `README.md`, `docs/ARCHITECTURE.md` |

## Executor notes
- Omitted `style/heading_test.go` (Stage 2) because tests in this repository cover only behaviour, never emitted CSS for combined properties, per the user's updated instruction.
