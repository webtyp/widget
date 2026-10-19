---
PLAN: "feat(style): add motion recipes Pivot, Wave, and Pulse for animated SVG parts"
TAG: v0.0.86
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# Plan — `webtyp/widget` v0.0.86: Motion recipes for SVG parts in `widget/style`

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Design gate

### 1. Prior art
- `webtyp/widget/style/flow.go:214`: `AutoRotate()` emits `@keyframes tw-auto-rotate` for cross-fading slides.
- `webtyp/widget/style/except_visual.go:166`: `Animate(m Motion)` emits transitions.
- Interactive vector parts (such as avatar hands, waving poses, pulsation indicators) need continuous CSS transform animations with specific transform origins (`transform-origin: 50% 100%` for wrists, `center` for pulsing rays/rings).

### 2. Novice-name test
- `style.Wave()`: waves a part back and forth with a gentle rotation (2–4deg) anchored at bottom-center (`transform-origin: 50% 100%; transform-box: fill-box;`).
- `style.Pulse()`: breathes/pulses scale (0.95 -> 1.05) anchored at center (`transform-origin: center; transform-box: fill-box;`).
- `style.Pivot(origin string)`: sets `transform-box: fill-box; transform-origin: <origin>;` on a part.

### 3. Complexity ledger
```
Concepts the developer must learn   +3 (Wave, Pulse, Pivot) / 0
Files they must touch to do X       0
Lines at the call site              1 line: style.Part(PartHand, style.Wave())
Ways to do the same thing           +1 (closed motion recipes) / 0
```

### 4. Where does it belong
- In `webtyp.com/widget/style` (`except_visual.go` and `emit_flowdecls.go`).

### 5. What does this change delete?
- Deletes the need to write ad-hoc CSS `@keyframes` in consumer projects for interactive SVG parts.

---

## Stage 1 — Add recipes to `style`

In `style/except_visual.go`:
- Add `Wave() Option`: registers a gentle alternating rotation animation on the part.
- Add `Pulse() Option`: registers a subtle scaling pulse animation on the part.
- Add `Pivot(origin string) Option`: sets `transform-box: fill-box; transform-origin: <origin>;`.

In `style/emit_flowdecls.go` and `style/emit_primitives.go`:
- Emit `@keyframes tw-wave { 0%, 100% { transform: rotate(-3deg); } 50% { transform: rotate(3deg); } }`
- Emit `@keyframes tw-pulse { 0%, 100% { transform: scale(0.95); opacity: 0.6; } 50% { transform: scale(1.04); opacity: 1; } }`
- Automatically disabled under `@media (prefers-reduced-motion: reduce)` in `emit_motion.go`.

## Stage 2 — Validation and Tests

- Update `style/validate.go` to validate allowed contexts.
- Add unit tests in `style/tests/` asserting correct layer emission under `@layer widgets` and reduced-motion suppression.

## Stage 3 — Documentation

- Update `GUIDE.md` and `docs/SPECS.md` with `Wave`, `Pulse`, `Pivot`.

## Acceptance
- `gotest` passes completely.
- Drift guard and WASM boundary assertions pass.
