---
PLAN: "feat(style)!: Grid takes a column cap — reflow by container, never past maxCols"
EXECUTOR: local
REVIEWER: none
---

> Master: `/home/cesar/.claude/plans/si-la-ui-se-drifting-treasure.md` (track B, gate B) · 2026-10-08.
> The previous PLAN.md (`style.Button`) shipped; it lives in git history.

# Plan — `style.Grid(maxCols, min, gap)`

## Problem

`form/css.go` lays fields out with `FixedGrid(2)` and drops to one column only through a
**viewport** media query. The login card is narrow on every screen, so on a desktop its only field
takes half the card. Forms must keep two columns wherever their container is wide enough. `Grid`
already reflows by container (auto-fit), but has no cap: a wide CRUD form would grow to 3–4 columns.

## API gate

1. **Prior art.** The CSS Grid "RAM" pattern (`repeat(auto-fit, minmax())`, Every Layout's Grid),
   capped as `minmax(max(min, 100%/N), 1fr)`. Bootstrap `row-cols-md-2` and Tailwind
   `md:grid-cols-2` decide by **viewport** — the defect itself. Tailwind v4 `@container` decides by
   container but needs a new concept (container queries). auto-fit is already container-relative
   (the `%` is the grid's own width), so no `container-type` is needed.
2. **Name.** Still `Grid`: the `Grid` (reflows) / `FixedGrid` (fixed) pair exists. The cap comes
   first, where `FixedGrid(cols, gap)` puts its count; `min ColumnWidth` is what says it reflows.
   `maxCols < 1` is a `Validate()` error, never "no cap".
3. **Ledger.** Concepts 0 · ways to build a reflowing grid 0 net (the only one changes; no
   `FluidGrid` beside it) · call site +1 argument · form loses its mobile media query (−1 rule).
4. **Where.** `widget/style` — a missing recipe is a defect here, not in a consumer.
5. **Deletes.** The uncapped grid; `form/css.go`'s `On(css.Mobile, PartForm, …)`.

Breaking: three consumers (`components/statgrid/css.go:14`, `layout/landing/css.go:44` and `:73`)
are updated in the same wave; each gets the cap it shows today at 1440 px.

## Steps

1. Red tests (`style/flow_test.go`): `Grid(2, ColumnMedium, Space2)` declares `--cols: 2;` and the
   capped track; `Grid(0, …)` fails `Validate()`; the cap survives `On()`.
2. `flow.go`: new signature, `flowCols` set. One unexported constant holds the track expression,
   used by `emit_primitives.go`, `emit_flowdecls.go`, `emit_device.go` (today the literal is
   written three times). `emit_decls.go` declares `--cols` for both grids.
3. `validate.go`: `Grid`/`FixedGrid` with `cols < 1` → error.
4. Docs: `docs/SPECS.md` table, `GUIDE.md`, `README.md`, `docs/MIGRATION.md`.
5. `gotest`; `gopush --no-cascade` (consumers are updated by hand right after).
