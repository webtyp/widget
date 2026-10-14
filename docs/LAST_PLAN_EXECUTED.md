# Plan — `style.AppBar(st, m)`

## Problem
Chassis layouts like `platformd` need an in-flow mobile header bar that retracts smoothly off-screen when scrolling down, freeing screen real estate without resorting to floating chrome or abrupt `display: none` / fade-only reveals. Hand-composing negative margins, transitions, and state selectors at the call site violates widget-styling principles.

## Target
Introduce `AppBar(st widget.State, m Motion)` in `webtyp/widget/style`:
- In-flow row with block size `calc(var(--control-height) + 2 * var(--space-1))`
- Retracts by translating or setting negative `margin-block-start` when `st` is false
- Emits transition on `margin-block-start` with motion `m` and `@media (prefers-reduced-motion: reduce)` override
- Safe composition validation against conflicting positioning/sizing options
