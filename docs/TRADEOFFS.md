# Trade-offs — `webtyp/widget`

What this architecture buys, what it costs, and what remains unsolved.

Distinct from [DESIGN.md](DESIGN.md): that document justifies decisions *taken*.
This one records the price of those decisions and the weaknesses that survive
them. Each weakness carries a proposed improvement with its justification, and an
honest note where the proposal is a judgement call rather than a fix.

Read this before extending the library, and before concluding that a limitation
you hit is a bug.

Only **open** costs are recorded here. A weakness that has been accepted and
scheduled stops being a trade-off and becomes work: its reasoning moves to
[DESIGN.md](DESIGN.md), its behaviour to [SPECS.md](SPECS.md), and its steps to
the execution plan. Keeping a copy here would mean two places to keep in sync,
and the copy would rot first.

---

## Part 1 — What the architecture buys

### P-1. The largest class of CSS bugs becomes untypeable

Closed scales, no free strings, no arbitrary values. A misspelt property, a unit
mistake, a colour that fails contrast, a magic number — none can be expressed.
This is the whole point: CSS is error-prone largely because its surface is
enormous and untyped, while what is actually needed to build is small.

### P-2. Markup and stylesheet agree by construction

`Class` has no public constructor. A class name can only be derived from a `Name`
and a `Part`, so the selector in the sheet and the attribute in the markup come
from the same expression. Divergence is not a discipline problem; it is
impossible.

This holds for **classes** by construction and for **state attributes** only when
the consumer runs the `StateAttrs()` check (`docs/DESIGN.md §17`). State attributes
are runtime choices; the selector and the markup live in different build tags and
cannot be compared statically.

### P-3. Design knowledge is encoded rather than required

A surface is a complete decision, contrast is guaranteed by the token catalog's
own test, interaction states are derived, and role and stacking follow from
`Kind`. The author supplies product judgement — *this is a dialog, this row is
selected* — and never visual judgement.

### P-4. The cascade is deterministic

Fixed layer order, flat specificity, no `!important`, no descendant selectors. A
rule's effect can be predicted from the rule alone, which is what makes the
output safe to concatenate across dozens of components.

### P-5. Zero client cost

The style engine is excluded from WebAssembly by build tag, enforced by test. The
binary shipped to the browser carries identity only.

### P-6. Output is diffable and cacheable

Byte-stable emission means a stylesheet diff shows intent, and `ssr`'s
content-hash cache is sound.

### P-7. The boundary is mechanically enforced

"Never invent a value" is checked by a drift guard over the emitted CSS, not
asserted in prose. That is the difference between an architecture and an
intention — and the four leaks it has already let through are the argument for
tightening it rather than trusting it.

---

## Part 2 — Open costs

Three remain unresolved. Each carries a proposed improvement and the reason it
has **not** been built yet — in every case, that building it now would cost more
than leaving it.

### C-1. Zero escape has no sanctioned exit

**The cost.** Every real project eventually needs a value the scale does not
have. Zero escape means the author's only recourse is hand-written CSS outside
the system — strictly worse than a controlled escape: invisible to the drift
guard, outside the layer model, and not themeable.

The constraint is stated as "no new values". That is not what the architecture
actually needs.

**Proposed improvement.** Restate the rule as **no *undeclared* values**, and add
one option that accepts a `css.Token` and nothing else:

```go
func Custom(prop string, t css.Token) Option
```

A one-off must then still be declared as a token in `webtyp/css`, where it is
themeable, dark-mode-aware and contrast-tested. The escape stays inside the model.

**Justification.** What matters is not the size of the scale — it is that every
value has a declaration someone can theme and test. A `css.Token` parameter
preserves that while removing the incentive to leave the system entirely. A
`string` parameter would not, which is why the signature takes a token.

**Why it is not built.** Adding a function is not a breaking change, so it can
land in any later minor release — the one-breaking-window argument does not apply.
And shipping an escape *before* knowing which values are genuinely missing is how
an escape becomes the default path. It is also the only proposal here that
weakens the core guarantee, which deserves evidence rather than speculation.

**Trigger.** Build it when the third genuine request arrives. Until then the
answer is to extend the token catalog, which is the sanctioned path already.

**Risk to weigh when building it.** `prop` is a free string — the one place zero
escape would be relaxed. Constrain it to an allow-list of properties the engine
does not otherwise emit, and have the drift guard assert `Custom` is the only
source of such declarations.

### C-2. Stacking derived from `Kind` cannot see composition

**The cost.** `Kind` yields one stacking level per pattern. Correct stacking is a
property of the *composition*, not the pattern: a dialog opened from within a
dialog needs to sit above it, and both resolve to `--z-modal`. Nesting the same
pattern is unrepresentable.

**Proposed improvement.** Let a widget that genuinely nests declare a relative
bump — `Backdrop(Viewport, Above(parentKind))`, resolving to the parent's level
plus one step. Bounded and typed, unlike a free integer.

**Justification.** Role and valid states genuinely follow from the pattern and
should stay derived; stacking does not, because it depends on what contains what,
which `Kind` cannot know by construction.

**Why it is not built.** Nested overlays have not been shown to occur in this
suite. Designing an escape for a case that may not exist is the same failure as
C-1, and the cost of being wrong is bounded: a component that needs it cannot be
built without leaving the library, which is a loud failure rather than a silent
one.

**Trigger.** The first component in the suite that nests two overlays of the same
pattern.

### C-3. One visual change is a three-repository change

**The cost.** Adding a surface family means a token plus a contrast test in
`css`, a constant plus a resolution entry in `widget`, and a release of each in
order. What a designer thinks of as "add a colour" is a coordinated multi-repo
change — the boundary that makes the architecture sound is what makes this slow.

**Proposed improvement.** Generate the surface resolution table in `widget` from
the `css` catalog with a `go:generate` step, plus a test asserting the generated
table is current. The `widget` side of a new family becomes mechanical and
verified rather than hand-written and forgettable.

**Justification.** The boundary is worth its cost — see
[DESIGN.md §1](DESIGN.md#1-why-webtypcss-stays) — so the answer is to reduce
the friction, not remove the boundary. Generation also closes a drift class
directly: a hand-written table can disagree with the catalog, and four such
disagreements exist in the published code.

**Why it is not built.** Pure internal tooling with no API surface, so it can
land at any time, and the drift it prevents is already caught by the emitted-CSS
drift guard. Adding a code generator to a twelve-step release buys nothing.

**Trigger.** When a second surface family is added.

**Not proposed: merging `css` into `widget`.** It would collapse the contrast
guarantee into the component library and make `css` unusable on its own — a much
larger loss than the coordination cost.

---

## Part 3 — Accepted limitations

Recorded so they are not rediscovered as bugs. Unlike Part 2 these have no
proposed improvement: they are consequences of decisions worth keeping.

| Limitation | Why it is accepted |
|---|---|
| One token catalog for all widgets | Per-widget theming would defeat the single visual system the library exists to enforce. Scope-level overrides of `--color-*` on a subtree remain available. |
| `Kind` is a closed enum | A component fitting no ARIA pattern is almost always two components. Extending the enum is a deliberate act, which is the intent. |
| No animation beyond a transition scale | Keyframes are an open-ended language; admitting them would reopen the whole surface the closed scales exist to shut. Components needing them fall back to application CSS. |
| Height cannot be declared | Content-driven height is what makes the primitives composable. `Fill()` and `Scroll()` cover the cases that need it. |
| Appearance cannot vary with component data | Sheets are static per component type, which is what makes extraction and caching possible. The custom-property route is specified in [ARCHITECTURE.md §8.1](ARCHITECTURE.md#81-appearance-that-depends-on-component-data). |
| No viewport-scoped mechanism at component level | A component that reacts to the viewport is unusable in a sidebar. Shell-level decisions are covered in [ARCHITECTURE.md §6.4](ARCHITECTURE.md#64-flow-primitives). |

---

## Related documents

- [ARCHITECTURE.md](ARCHITECTURE.md) — the structure being assessed.
- [DESIGN.md](DESIGN.md) — why each decision was taken.
- [SPECS.md](SPECS.md) — exact behaviour.
