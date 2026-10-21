---
PLAN: "feat(style): Blink, Breathe and Frame motion recipes — CSS-only loops for illustrated widgets"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 2069284888834209850
PR: https://github.com/webtyp/widget/pull/10
REVIEWED_COMMIT: d849680913bf936e385213f0b2d260dacebf528e
---

> Read AGENTS.md at the repo root before any change.

# Plan — `widget/style`: `Blink()`, `Breathe()`, `Frame(index, count, cadence)`

`webtyp/avatar` animates an Inkscape illustration (layers rendered as `<use>` inside an `<svg>`).
Every repeating motion (eyes blinking, chest breathing, mouth syllables, "thinking" dots) must run
in CSS so the WASM side never runs a timer per frame. `Wave()` and `Pulse()` already exist in this
package for the same purpose (`style/except_visual.go:234-246`); this plan adds three siblings and
makes the "one animation per element" rule explicit.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Ecosystem search

- `widget/style` `Wave()`/`Pulse()`/`Pivot()` (v0.6.42): same family → **extend** here, same files.
- `widget/style` `AutoRotate()` (`flow.go:214`): closest to a flipbook (staggered layers), but it
  crossfades on a fixed cycle for an image carousel; a flipbook needs hard cuts at a per-frame
  cadence → not reused, not changed.
- `webtyp/css` keeps keyframe builders private (`css/dsl.go:86`); raw keyframe strings in this
  package already exist (`waveKeyframesCSS`, `emit_flowdecls.go:218`) → same approach.

## Design gate

1. **Prior art.** CSS sprite animation with `steps()` (one element, hard frame cuts — the web's
   flipbook); Adobe Character Animator's *Cycle Layers* behavior (cycles the layers of a group at a
   frame rate) and its auto-blink layer; Rive/Lottie frame timelines (need their own runtime). We
   keep the CSS approach: zero runtime, compositor-driven, `prefers-reduced-motion` handled here.
2. **Novice names.** `style.Blink()` (the part blinks), `style.Breathe()` (the part breathes),
   `style.Frame(1, 3, style.CadenceFast)` ("frame 1 of 3, fast cadence"). `Cadence` follows
   `Motion`'s `Fast/Slow` vocabulary (`scale.go:158`).
3. **Ledger.** Concepts +4 (`Blink`, `Breathe`, `Frame`, `Cadence`) / −0. Files to touch for a
   consumer: 1 (`css.go`). Ways to do the same thing: +0 (no flipbook existed).
4. **Where.** Motion recipes live in `widget/style` (skill widget-styling: every CSS rule of a
   widget comes from a recipe).
5. **Deletes.** Nothing here; downstream `webtyp/avatar` deletes its per-frame Go timers.

## The contract

```go
// scale.go — after Turn
// Cadence is how long each frame of a Frame() flipbook stays on screen.
type Cadence uint8

const (
	CadenceFast Cadence = iota // 150ms per frame — syllables of speech
	CadenceSlow                // 400ms per frame — dots of a "thinking" indicator
)

// except_visual.go — after Pulse()
func Blink() Option                                // the part closes and opens vertically every 4s
func Breathe() Option                              // the part rises slightly from its base every 4s
func Frame(index, count int, c Cadence) Option     // the part is one frame of a looping flipbook
```

Fields on `rule` (`sheet.go`, next to `hasWave`/`hasPulse`): `hasBlink bool`, `hasBreathe bool`,
`hasFrame bool`, `frameIndex int`, `frameCount int`, `frameCadence Cadence`.

### Exact CSS per option (declarations go where `hasWave` emits them, `emit_decls.go:130-141`)

| Option | Declarations, in this order |
|---|---|
| `Blink()` | `transform-box: fill-box;` `transform-origin: center;` `animation: tw-blink 4s ease-in-out infinite;` |
| `Breathe()` | `transform-box: fill-box;` `transform-origin: 50% 100%;` `animation: tw-breathe 4s ease-in-out infinite;` |
| `Frame(i, n, c)` | `animation-name: tw-frame-<n>;` `animation-duration: <n*ms>ms;` `animation-timing-function: step-end;` `animation-delay: <d>ms;` `animation-iteration-count: infinite;` |

`ms` = 150 for `CadenceFast`, 400 for `CadenceSlow`. `d` = `-(((n - i) % n) * ms)`, so frame 0 has
`0ms` and is visible first. Worked example, `n = 3`, `CadenceFast`: duration `450ms`; delays: i=0 →
`0ms`, i=1 → `-300ms`, i=2 → `-150ms`.

### Keyframes (appended once per sheet, like `waveKeyframesCSS`, `emit.go:57-63`)

```css
@keyframes tw-blink {
  0%, 92%, 100% { transform: scaleY(1); }
  95% { transform: scaleY(0.1); }
}

@keyframes tw-breathe {
  0%, 100% { transform: scaleY(1); }
  50% { transform: scaleY(1.012); }
}

@keyframes tw-frame-3 {
  0% { opacity: 1; }
  33.333%, 100% { opacity: 0; }
}
```

`tw-frame-<n>` is emitted once for each distinct `n` used in the sheet, in ascending `n`. The
breakpoint is a literal from this table (a `switch`, never float formatting):

| n | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|
| breakpoint | `50%` | `33.333%` | `25%` | `20%` | `16.667%` |

Collect the distinct counts in a `[]int` scanned linearly (no `map`, AGENTS.md).

## Stage 1 — one predicate for "this rule animates"

1. `style/sheet.go` (or next to `rule`): add
   `func (r rule) animated() bool { return r.hasWave || r.hasPulse || r.hasBlink || r.hasBreathe || r.hasFrame }`.
2. `style/emit_motion.go:118-185`: replace every `X.hasWave || X.hasPulse` with `X.animated()`
   (the root, parts, stateRules, cueWithin, cueWithinHover, cueAcross, stateAcross, deviceRules,
   cueRules loops). Check: `grep -n "hasWave || .*hasPulse" style/emit_motion.go` → empty.
3. `style/emit.go:57-100`: replace `hasWaveOrPulse()` with `usedKeyframes()` returning
   `(wave, pulse, blink, breathe bool, frameCounts []int)` and emit, in this order: wave, pulse,
   blink, breathe, then each frame count ascending. Check: `grep -n "hasWaveOrPulse" style/` → empty.
4. `emit_decls.go:243` (the "rule is empty" predicate): add `!r.hasBlink && !r.hasBreathe && !r.hasFrame`.

## Stage 2 — options, declarations, keyframes

`except_visual.go`: the three options and doc comments (one line each, as `Wave`). `scale.go`:
`Cadence`. `emit_decls.go`: the declaration table above (Pivot keeps priority for the origin, as
today). `emit_flowdecls.go`: `blinkKeyframesCSS()`, `breatheKeyframesCSS()`,
`frameKeyframesCSS(n int) string` next to `waveKeyframesCSS()`.

## Stage 3 — reduced motion

Inside the existing `@media (prefers-reduced-motion: reduce)` block for animations
(`emit_motion.go:180-184`), every animated selector already gets `animation: none;` through
`animated()`. Add a second rule in the same media block for every `Frame` selector with
`frameIndex > 0`: `opacity: 0;` — otherwise, with animations off, all frames show at once and
overlap. Frame 0 stays visible.

```css
@media (prefers-reduced-motion: reduce) {
.avatar__mouth-a, .avatar__mouth-o, .avatar__mouth-rest {
  animation: none;
}
.avatar__mouth-o, .avatar__mouth-rest {
  opacity: 0;
}
}
```

## Stage 4 — validation (`validate_composition.go:30-60`)

One `animation` per element: count `hasWave, hasPulse, hasBlink, hasBreathe, hasFrame`; more than
one → error. Extend the transform and origin counts with `hasBlink`, `hasBreathe` (Frame touches
neither). New and changed messages, verbatim (update `style/validate_test.go:268-288`):

- `sheet %s: part %q: Wave, Pulse, Blink, Breathe and Frame all set animation; use only one`
- `sheet %s: part %q: Rotate, Wave, Pulse, Blink and Breathe all set transform; use only one`
- `sheet %s: part %q: Wave, Pulse, Blink, Breathe and Pivot all set transform-origin; use only one`
- `sheet %s: part %q: Rotate/Wave/Pulse/Blink/Breathe cannot combine with OnEdge/Drawer/FloatMiddle — all own transform`
- `sheet %s: part %q: Frame count must be 2 to 6, got %d`
- `sheet %s: part %q: Frame index must be 0 to %d, got %d` (the bound printed is `count-1`)

## Stage 5 — tests (`style/motion_test.go`, extend `TestMotion_WavePulsePivot`'s style)

| Setup | Expect |
|---|---|
| `For(avatar).Part("eyes", Blink())` | `.avatar__eyes` has the three Blink declarations; output contains `@keyframes tw-blink {` once |
| `Part("body", Breathe())` | Breathe declarations; `@keyframes tw-breathe {` |
| `Part("a", Frame(0,3,CadenceFast)).Part("o", Frame(1,3,CadenceFast)).Part("rest", Frame(2,3,CadenceFast))` | delays `0ms`, `-300ms`, `-150ms`; duration `450ms`; exactly one `@keyframes tw-frame-3 {` with `33.333%, 100% { opacity: 0; }` |
| `Frame(0,2,CadenceSlow)` and `Frame(0,4,CadenceSlow)` in one sheet | `tw-frame-2` before `tw-frame-4`; durations `800ms`, `1600ms` |
| the three Frame parts above | reduced-motion block holds `animation: none;` for all three and `opacity: 0;` for `o` and `rest` only |
| `Part("x", Wave(), Frame(0,2,CadenceFast))` | `Validate()` returns the "all set animation" error |
| `Part("x", Blink(), Rotate(TurnHalf))` | the "all set transform" error |
| `Frame(0,7,CadenceFast)` / `Frame(3,3,CadenceFast)` | the two Frame range errors, verbatim |
| same sheet built twice | identical output (determinism, as `TestMotion_Determinism`) |

Add `Blink()`, `Breathe()` and one `Frame` part to the giant sheet of `TestNoInventedValues`
(`style/guard_test.go:55`) so the guard covers them.

## Stage 6 — docs

`docs/SPECS.md`: signatures next to `Wave()` (line 336) and three rows after `Pulse()` (line 408)
in the same format, plus a `Cadence` row with its two values. `GUIDE.md`: where `Wave()` is shown,
one example of a three-frame flipbook (the mouth parts above). Do not link `docs/PLAN.md`.

## Delivery check

```bash
GOFLAGS=-mod=mod GOPRIVATE=webtyp.com/* go build ./...
gotest
grep -n "hasWave || .*hasPulse" style/emit_motion.go   # empty
grep -rn "hasWaveOrPulse" style/                       # empty
```

A PR that does not compile, or whose tests could not run, is a failed delivery. Never pause to
ask permission or report progress: when the check passes, commit, push and open the PR. Stuck for
30 minutes → open the PR anyway, listing what fails.

## Stages

| # | What | Files |
|---|---|---|
| 1 | `animated()` predicate, `usedKeyframes()` | `style/sheet.go`, `style/emit_motion.go`, `style/emit.go`, `style/emit_decls.go` |
| 2 | options, `Cadence`, declarations, keyframes | `style/except_visual.go`, `style/scale.go`, `style/emit_decls.go`, `style/emit_flowdecls.go` |
| 3 | reduced motion for frames | `style/emit_motion.go` |
| 4 | validation | `style/validate_composition.go`, `style/validate_test.go` |
| 5 | tests | `style/motion_test.go`, `style/guard_test.go` |
| 6 | docs | `docs/SPECS.md`, `GUIDE.md` |
