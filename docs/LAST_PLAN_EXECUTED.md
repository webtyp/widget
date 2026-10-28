---
PLAN: "feat(style): Paint, Ink, Line, Dashed and Typeface — SVG drawings coloured from the theme like any box"
EXECUTOR: local
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# Plan — `webtyp.com/widget/style`: paint SVG with the surface roles

## Why

`webtyp/erd` draws the database diagram as SVG and, with no recipe to colour an SVG shape, wrote its own
`<style>` with hex colours and a private `@media (prefers-color-scheme: dark)`: inside an app switched to
light the diagram stayed dark, and no token reached it. **widget-styling** says a look the DSL cannot express is
a missing recipe here, never CSS at the call site. SVG paints with `fill` and `stroke`; `As()` only sets
`background-color`, `color` and `border`, which an SVG shape ignores.

## Ecosystem search

`grep -rn 'fill\|stroke' widget/style` → only `Glyph(s)` (`except.go:78`): `color` = the family base colour and
`fill: currentColor`, for icons tinted with a brand colour. It cannot give a shape a surface background and an
outline, nor text the surface's text colour (Panel has no family). `css.Token` already resolves every role to a
static light value and an enhanced var (`surface.go:92`, `triplet`). Decision: **extend** `widget/style`.

## Design gate

1. **Prior art:** SVG's own two paints (`fill`, `stroke`); Tailwind `fill-*` / `stroke-*` utilities bound to the
   palette; Material's `tonal` roles applied to icons and charts. All colour SVG from the same roles as boxes.
   Here the role is a `Surface`, so the recipes take one, as `As()` does.
2. **Names:** `Paint(Panel)` — «paint the shape like a Panel» (fill = its background, stroke = its border);
   `Ink(Panel)` — «write in the Panel's ink» (text fill = its text colour); `Line(Danger)` — «draw the line in
   the Danger colour» (stroke only, no fill); `Dashed()` — the stroke is dashed.
3. **Ledger:** concepts +4 / −0; ways to colour SVG from the theme 1 (`Glyph`, icons) → 2 (`Glyph` for a
   brand-tinted icon, these for drawings) — different intents, documented on both; hand-written SVG CSS in the
   ecosystem 1 → 0 (erd, next plan).
4. **Belongs:** `widget/style` owns every look; the roles live in `css`.
5. **Deletes:** nothing here; it lets `erd/draw.go` delete its `<style>`.

## Stage 1 — `paint.go` (new, `//go:build !wasm`)

Options `Paint(s Surface)`, `Ink(s Surface)`, `Line(s Surface)`, `Dashed()`, each with a doc comment and an
example. Emission (static light value first, enhanced second, as `surfaceDecls` does):

- `Paint`: `fill` ← the surface background; `stroke` ← its border colour when the role has a border.
- `Ink`: `fill` ← the surface text colour; `stroke: none`.
- `Line`: `fill: none`; `stroke` ← the border colour when the role has one, otherwise its background (the
  colour that names a solid role: Danger, Primary).
- `Dashed`: `stroke-dasharray: var(--space-1) var(--space-1)`.

`rule` gains `hasPaint/paint`, `hasInk/ink`, `hasLine/line`, `dashed`; `Decls` appends `r.paintDecls()` right
after `surfaceDecls()`.

## Stage 2 — validation

`validateComposition`: on one part, `Paint`, `Ink` and `Line` all set `fill` → more than one is an error
(«Paint, Ink and Line all set fill; use one»). `Dashed()` without `Paint` or `Line` → error («Dashed needs a
stroke: Paint or Line»).

## Stage 3 — tests (`paint_test.go`)

Each recipe emits its paint from the role's tokens; two of them on one part fail `Validate()`; `Dashed()` alone
fails.

## Stage 3b — `Typeface()`

`font-family: var(--font-sans)`, emitted next to `FontSize`. A page inherits the font from `body`; a standalone
SVG (erd's checked-in diagram) has no body and fell back to the viewer's serif. Test: it emits the token.

## Stage 4 — docs

`README.md` recipe table: one row per option. Skill **widget-styling**: name the four in «Use the composed
recipe» and in the «No CSS text anywhere» section.

| Stage | File | Action |
|---|---|---|
| 1 | `style/paint.go`, `style/sheet.go`, `style/emit_decls.go` | four options |
| 2 | `style/validate_composition.go` | conflicts |
| 3 | `style/paint_test.go` | behaviour |
| 4 | `README.md`, skill | docs |
