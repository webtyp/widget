# AGENTS.md — webtyp/widget

Constraints for agents changing this library. Read before touching any file.

## Mission

Visual component contracts for the ecosystem: identity, anatomy, states and ARIA kinds on
one side; a closed styling API that emits deterministic CSS on the other.

The goal that decides every argument here: **someone who does not know design builds a
correct, accessible widget without reading this library's source.** An option that
requires design judgement to use correctly is a defect, even if it compiles.

Consumers: `webtyp/components`, `webtyp/layout` (their `//go:build !wasm` `css.go`
files), `webtyp/ssr` (extracts the stylesheets). Depends on `webtyp/css` for every
value and on `webtyp/fmt` for everything else.

---

## 1. Two packages with opposing constraints

**`widget` (root) travels inside the WASM binary.** Identity, parts, states, cues, kinds.
Zero style logic, zero emission, and it imports only `webtyp/fmt` — never
`webtyp/css`.

**`widget/style` never reaches the client.** Every file carries `//go:build !wasm`. Scales,
surfaces, flow primitives and CSS emission live here.

The boundary is enforced by test, not by convention: under `GOOS=js` the dependency graph
of `webtyp.com/widget` must not contain `widget/style`. Do not add an import that
crosses it, and do not "temporarily" move a symbol to the wrong side — the whole zero
client cost argument rests on this line.

## 2. Never invent a value

Every emitted value is a `webtyp/css` token reference, fallback included. A drift guard
compares each emitted `var()` against the catalog and fails on any mismatch.

- Do not write a `var(--token,#hex)` by hand; call `css.<Token>.Var()`. A hand-written
  fallback silently goes stale the day the token changes.
- Do not define a `color-mix()` formula here. Deriving a colour is `webtyp/css`'s job —
  a formula duplicated across consumers is exactly the drift the token catalog exists to
  eliminate. If a step seems to need a new formula, the css side is incomplete: fix it
  there and release it, then come back.
- The only literals the guard permits are geometry: `Size` percentages and `Aspect`
  fractions, enumerated in `docs/SPECS.md` §2.2 and nowhere else.

## 3. Zero escape

No free strings, no `vw`/`vh`, no arbitrary values, no `!important`. If a value is not on a
closed scale it cannot be expressed — that is the feature, not a gap to fill.

The sheet builder is the only public construction path; the rule structures behind it stay
private so there is no second way in. Before adding an option, check that the same result
is not already reachable: **one way to do a thing**, and a new option must earn its place
against the existing decision table in `GUIDE.md`.

The one sanctioned hole is specified in `docs/ARCHITECTURE.md` §8.1: a component may write
a **numeric** custom property onto its element and consume it through `var()`. Keep it
numeric and keep it rare; it is a hole by design, not permission to widen.

## 4. Closed scales, complete surfaces

A scale is sized to what the token catalog can actually distinguish. A step that renders
identically to its neighbour teaches the author that the library is broken — if two steps
collapse to one token, the scale is wrong, not the catalog.

A `Surface` is a whole visual decision — background, text, border, radius, plus its own
hover/focus/press treatments — resolved together. An author picks *what a thing is*
(`Panel`, `Primary`, `Danger`), never what it looks like, and cannot pair one family's base
with another family's hover. Padding stays explicit: it depends on what a part contains.

## 5. Emission guarantees (all of them are tested)

1. **Deterministic**: two emissions are byte-identical.
2. **Fixed cascade layers** in a stable order: `tokens, primitives, widgets, states` —
   this is what removes specificity conflicts without `!important`.
3. **Flat specificity**: every selector is `.class`, `.class[data-*="true"]` or
   `.class:pseudo`. Nothing couples to DOM structure.
4. **No invented values** (§2), **no unreachable selectors**, no empty `@layer`.

Changing the emitted output changes every consumer's stylesheet **without a compile
error**. That class of change is the most expensive to defer and the most dangerous to
ship quietly: it belongs in a release with its migration note, never in a drive-by commit.

## 6. Mistakes must be loud

`Validate()` returns **all** problems, not the first, and each message names the sheet and
the part. The panic from `Stylesheet()` surfaces from inside `ssr`'s generated program,
far from the source that caused it — a vague message there costs an author an afternoon.

Anything the compiler cannot reject must be caught by validation. Silent CSS that matches
nothing is the failure mode this library exists to remove.

## 7. The SSR contract binds every style builder

`ssr` instantiates the component as a **zero value** and calls the provider by name.
Therefore a style builder:

- **must not read fields** — it runs on `&T{}`;
- **must be pure and deterministic**.

Sheets are concatenated across components, so whatever one sheet repeats is repeated once
per component in the shipped CSS. Keep per-sheet preamble minimal; cross-sheet
deduplication belongs to `ssr`, the only layer that sees all sheets at once.

---

## Adding or changing an option

1. Check `GUIDE.md`'s decision table first: if the need is already covered, the answer is
   documentation, not a new option.
2. Types and scales in `widget/style`, identity in `widget` — see §1 before choosing a
   file.
3. Emission goes through the existing rule structures; no new public escape hatch.
4. Add the validation condition that makes its misuse loud (§6).
5. Update `docs/SPECS.md` (the exact emitted output is part of the spec) and `GUIDE.md`.
6. `gotest`.

Removing or renaming a public identifier is a breaking change for `components` and
`layout`: the replacement mapping ships in `docs/MIGRATION.md` in the same release. No
aliases, no deprecation shims — a single release, with the mapping written down.

## Testing

```bash
go install webtyp.com/devflow/cmd/gotest@latest   # external agents have no global gotest
gotest
```

`gotest`, never `go test`. Stdlib assertions only (`testing`/`strings`, no testify). The
suite includes the WASM boundary assertion and the drift guard: these are part of the
design, not incidental coverage. Weakening a guard requires the same justification as
changing the rule it protects — if a guard fails, the cause is almost always the code, not
the guard.

Defect reproducers stay in the suite as regression tests once the defect is closed.

## Publishing

`gopush 'message'`, never a raw `git commit`/`push`. Documentation is updated **before**
publishing, in the same commit as the code it describes.

## Documentation

`docs/ARCHITECTURE.md` (structure and invariants), `docs/SPECS.md` (exact API, values and
emitted output — the authority; if code and SPECS disagree, one of them is wrong and both
are fixed in the same commit), `docs/DESIGN.md` (why, and what was already rejected —
consult it before reopening a settled decision), `docs/TRADEOFFS.md` (accepted costs and
limitations: check here before reporting one as a bug), `docs/MIGRATION.md` (consumer
upgrade path), `GUIDE.md` (the decision table authors read instead of exercising
judgement). `README.md` indexes every file in `docs/`.

`docs/PLAN.md` and any `docs/PLAN-*.md` are ephemeral: they are deleted in the commit that
publishes the work they describe, so no permanent document may link to them. Everything
written in English. Diagrams: `flowchart TD`, no `subgraph`, `<br/>` for line breaks.
