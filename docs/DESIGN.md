# Design decisions — `webtyp/widget`

Justifies the technical decisions behind [ARCHITECTURE.md](ARCHITECTURE.md) and
records the alternatives that were rejected. Does not restate the architecture,
and does not specify behaviour — that is [SPECS.md](SPECS.md).

Read this when a decision looks arbitrary and you need to know whether it can be
changed.

---

## 1. Why `webtyp/css` stays

**Decision.** Keep the dependency, and narrow this module's use of it to one rule:
*reference tokens, never invent values.*

The question is fair, because the dependency looks thinner than it is. Only three
things are used from `css`, and one of them is nearly free:

1. The **token catalog** — `ColorPrimary`, `Space4`, `RadiusMd`, `DurationFast` —
   resolved through `.Var()`.
2. **`css.Token`**, the type.
3. **`css.Stylesheet`** as an opaque container. The typed DSL inside `css`
   (`rule`, `decl`, `media`, `keyframes`) is unexported, so the engine builds its
   whole sheet as a string and wraps it in `css.Raw`.

Point 3 contributes a return type and nothing else. It would be tempting to
conclude the dependency is ceremonial. It is not, for two reasons:

**Someone must declare the variables.** This module emits *references*:
`var(--color-primary, #1E6B9E)`. A reference with no declaration falls back to a
literal, which means no theming and no dark mode. `css.RootCSS()` and
`css.Theme(css.Set(…))` are what make the fallback a fallback rather than the
value.

**The contrast guarantee cannot live here.** `css` carries a contrast test across
the palette. That test is precisely what makes `As(Primary)` safe to hand to
someone who cannot evaluate contrast themselves — the entire premise of this
library. A guarantee about values belongs with the values; duplicating the palette
here to avoid a dependency would move the guarantee somewhere it cannot be
maintained.

**Rejected: vendoring the tokens.** Removes the dependency, duplicates the
palette, and splits the contrast guarantee across two repositories that will
drift. The drift is not hypothetical — see §2.

**Rejected: exporting the `css` DSL and building sheets with it.** Larger API in
`css`, no benefit here: the engine's output is fully determined by the sheet, and
a string builder emits it deterministically in one pass. Revisit only if `css`
needs the DSL for its own reasons.

**Consequence.** `css.Stylesheet` stays as the return type even though the DSL is
unused, because it is the integration contract `ssr` matches on.

---

## 2. Why the boundary needs enforcing, not just stating

The rule "never invent a value" was already implied by the architecture and is
already violated in four places. That is the evidence that a stated rule is not
enough:

- Eighteen tokens with hardcoded hex live in the style package instead of `css`,
  outside the reach of the contrast test.
- Two surfaces return literal `rgba(0,0,0,…)` washes, which are invisible on a
  dark surface — broken in dark mode by construction, and invisible to a guard
  that only inspects `var()` calls.
- One overlay declaration emits a `var()` with no fallback, drifting from the
  catalog.
- Stacking is hardcoded to integers while `css` publishes a `--z-*` scale, so a
  sticky element can render above a modal.

**Decision.** The drift guard becomes the enforcement mechanism: every `var()` is
compared against the catalog *including its fallback*, the sheet is asserted to
contain no colour literals, and the fixture must exercise **every** option — the
current fixture does not, which is how the overlay drift got in.

**Rejected: a lint rule.** Slower to write, easier to disable, and it cannot see
the emitted output, which is where the invariant actually matters.

---

## 3. Why visibility is one declaration, not a pair

**Decision.** A single option binds hiding to a state; the engine restores the
display mode the element's flow requires.

The pair form has two independent faults. The ordering rule ("the reveal goes in a
state rule, never in the base") lives only in a comment and nothing checks it. And
the reveal emitted a fixed `display`, which overrides the flow primitive's own
`display` from an earlier cascade layer — so revealing a row turned it into a
block, stacked its children, and killed its gap. The Go read correctly and the
symptom existed only in the revealed state.

**Rejected: `display: revert-layer`.** This is the obvious fix and it does not
work here. The base `display: none` sits in the `widgets` layer, so reverting from
the `states` layer rolls back exactly to that `none` and the element stays hidden.
Making it work would mean moving the base declaration out of `widgets`, which
complicates emission for no gain over resolving the value directly.

**Rejected: `[hidden]` in markup.** Moves a styling concern into the DOM and gives
up the state-attribute model the rest of the library uses.

---

## 4. Why an invalid sheet panics

**Decision.** Validation returns all problems; emission panics if any exist.

The failure this prevents is a rule attached to a part that does not exist —
typically a typo. It emits CSS that matches nothing, so the page is simply
unstyled in one spot, with no error anywhere. An author who does not already know
the library has no way to find it.

A returned `error` does not fix this. The code path that builds a sheet is
initialisation-shaped, the error would be handled by ignoring it, and the failure
would stay silent. A panic naming the offending part is actionable in seconds.

This mirrors `regexp.MustCompile`: a malformed pattern known at authoring time is
a programming error, not a runtime condition.

**Rejected: returning `(*css.Stylesheet, error)`.** Changes every call site into
ceremony to guard against a condition that is never recoverable.

**Rejected: logging a warning.** Warnings in a build that stays green are how the
original problem shipped.

---

## 5. Why a surface carries shape

**Decision.** A surface resolves background, text, border and **radius**
together, not colour alone. Padding is **not** folded in.

Splitting radius out means an author who has already said "this is a Panel" is
then asked to choose a radius — a decision they have no basis for, on a thing
whose identity already implies it. In practice every panel in a codebase gets the
same radius, chosen by copy-paste, until one does not.

The test for folding a property in is whether its value follows from the
surface's *identity*. Radius passes: two panels always want the same one, and
that sameness is what makes them read as one system. Padding fails: it follows
from what the part contains, which the surface cannot see, so folding it in would
convert a saved call into a remembered exception.

**Rejected: presets like `Card()` or `Toolbar()`.** More names, less
composability, and the set is never complete — every project needs the ninth one.

**History.** An earlier revision folded padding in as well. That was an
over-reach: it converted a saved call into a remembered exception precisely in
the cases that differ from the default, which is a net loss.

---

## 6. Why interaction states are derived, not chosen

**Decision.** One option applies a surface and derives its hover, focus and press
treatments.

The alternative — a constant per state per family — produced a large flat
namespace where nothing prevented pairing one family's base with another family's
hover. That combination compiles, emits, and looks like a bug rather than a
mistake. It is also the exact decision the library claims to have removed from the
author.

**Consequence.** The per-state constants become private. They are still needed
internally, but they are no longer a menu the author can pick wrongly from.

**Consequence.** The cue mechanism survives as an escape hatch for the rare
browser-owned state that is not part of an interaction family, not as a normal
tool.

---

## 7. Why stacking derives from `Kind`

**Decision.** Overlay stacking is looked up from the widget's pattern, mapped onto
the token scale. The author never names a level.

Z-index is the property non-specialists get wrong most reliably, because
correctness is global: the right value depends on every other overlay in the
application. No local decision can be correct in isolation, so asking for one
guarantees eventual breakage — which is what the hardcoded integers produced.

A pattern, by contrast, determines a level unambiguously: a dialog is above a
dropdown, a toast is above both. The author already declares the pattern.

**Consequence.** This is also the answer to "`Kind` is required but nothing reads
it." Either it earns its place this way, or it and the capability interfaces
should be deleted. An API that demands data it ignores teaches people not to trust
its other demands.

**Rejected: an explicit level scale.** Smaller change, keeps the decision with the
author, and keeps the failure mode.

---

## 8. Why the scales were resized

**Decision.** Scales are sized to what the token catalog can actually distinguish,
and each scale means exactly one thing.

The spacing scale advertised thirteen steps and resolved to six distinct values;
adjacent steps produced identical output. An author who changes a step and sees no
change concludes the library is broken, and they are not wrong.

One scale served two different meanings depending on the primitive consuming it —
a column share in one, an aspect ratio in another — so the same constant produced
a square where its name promised a half. Two meanings need two types.

One option changed meaning depending on which other option accompanied it. That is
undiscoverable by construction: the signature cannot express it and the name does
not hint at it. The primitive now takes the value it needs directly.

**Rejected: keeping the numeric spacing scale and documenting the duplicates.**
Documentation cannot fix a scale whose steps are indistinguishable at the point of
use.

---

## 9. Why the utility selectors are dropped

**Decision.** Stop emitting `.fl-*` / `.exc-*` selectors alongside widget
selectors.

`Class` has no public constructor, so no consumer can put those class names in
markup. The selectors are unreachable. They were also emitted once per sheet, and
`ssr` concatenates sheets, so the dead weight multiplied by the number of
components.

---

## 10. Why deduplication belongs to `ssr`, not here

**Decision.** This module minimises what a single sheet repeats; cross-sheet
merging is `ssr`'s responsibility.

A sheet is built from one widget's declarations and cannot know how many other
widgets exist, so it cannot decide what is redundant. `ssr` merges all sheets and
is the only layer that can see the duplication. Attempting it here would require
global state across sheet construction, which would break both determinism and the
zero-value provider contract.

---

## 11. Why one breaking release

**Decision.** Ship every change together rather than phasing it.

The module has no published tag, so the cost of breaking is at its lifetime
minimum. Phasing would mean designing several signatures twice — once
compatibly, once correctly — and would leave consumers migrating in two hops for
no benefit. The known consumers are inside the same suite.

---

## 12. Naming

**Decision.** A name must be readable by someone who knows Go and HTML but not
design, and must not require translation to read the CSS it emits.

This library's premise is that the author does not have a designer's vocabulary.
An API that then names things in that vocabulary contradicts itself: the closed
scales prevent the author from choosing wrongly, but a name they cannot read
prevents them from choosing at all.

Four distinct problems hide under "abbreviations", and each needs a different
fix. Conflating them produces a rename pass that expands the harmless cases and
leaves the harmful ones intact.

### 12.1 Truncations — resolved by mirroring the token catalog

`Sm`, `Md`, `Lg`, `Xs`, `Xl`, `Opt`. Shortenings of ordinary words. Expanding
them costs verbosity and buys unambiguity, so the question is where the line
sits.

The rule that settles it without argument: **a scale step is named after the
token it resolves to.** A name that mirrors the emitted variable needs no
vocabulary at all, and — more importantly — needs no translation when the author
is looking at the rendered rule in devtools and asking which Go constant produced
it.

Applied to spacing, this makes the scale numeric (`Space1` → `--space-1`), which
also fixes the original defect honestly: the catalog has no `--space-5`, so there
is no `Space5`, and the gap in the sequence *is* the documentation. Applied to
radius and text size, it keeps `Sm`/`Md`/`Lg`, because that is what those tokens
are called.

**Consequence.** Where the truncation is genuinely unreadable, the question
belongs to `webtyp/css`, not here. Renaming only on the Go side would break the
mirror and reintroduce the translation step this rule exists to remove.

`Opt` is not a scale and gets expanded to `Option`: it is the most repeated type
name in the generated documentation, and the Go standard library does not
abbreviate.

### 12.2 Specialist vocabulary — replaced, not expanded

`Reel`, `Cover`, `Frame`, `Scrim`, `Prose`, `Track`, `Sunken`, `Flush`. These are
whole words, so expansion does nothing; they are opaque because they belong to
vocabularies the reader does not have — *Every Layout* for the first three, stage
lighting for `Scrim`, typography for `Prose`, CSS Grid for `Track`, and design
convention for the rest.

Each is replaced by what the thing does. `Reel` becomes `ScrollRow`, because it is
a row that scrolls. `Frame` becomes `MediaBox`, because its generated child rules
literally target `img` and `video`. `Track` becomes `ColumnWidth`, because the
argument is a minimum column width.

**Rejected: keeping the vocabulary and documenting it in a glossary.** A glossary
is a lookup the author must know to perform, which is the same failure as the
decision it replaced — and it is only ever read once.

**Rejected: keeping the vocabulary because it is standard.** It is standard among
people who have read the same sources. The premise of this library is that the
author has not.

### 12.3 Names that collide with a different, well-known meaning

`Fixed()` emits `flex-shrink: 0; flex-grow: 0` — "does not change size". In CSS,
`position: fixed` means "positioned relative to the viewport". A reader who knows
CSS reads this name **exactly wrong**, which is worse than not reading it at all:
opacity prompts a lookup, false familiarity does not.

This is the one category where a rename is not a judgement call. `Fixed` becomes
`KeepSize`.

### 12.4 Near-synonyms that do not distinguish

`Muted` and `Dimmed` were introduced as separate surfaces meaning, respectively,
secondary text and a disabled control. Both words mean "faded". Nothing in either
name says which is which, so the pair can only be used correctly by someone who
already knows.

They are renamed for their **role**, not their appearance: `Subtle` and
`Inactive`. `Inactive` now maps to `--color-surface` and `--color-muted`, avoiding the
need for a separate, private disabled color.

`Accent` has the same defect in a quieter form: it resolves to `--color-primary`,
so the name adds a translation step and buys nothing. It becomes `Primary`.

### 12.5 What is deliberately not renamed

- `Stack`, `Row`, `Split`, `Grid`, `Center` — plain English that matches what they
  produce.
- `Name`, `Part`, `Class`, `State`, `Kind` — Open UI, ARIA and DOM vocabulary the
  author already meets while writing markup. Renaming them would break the
  correspondence with the thing being described.
- `Backdrop` — CSS has a `::backdrop` pseudo-element; the term is already the
  reader's.
- `Elevation`'s steps `Flat`, `Raised`, `Floating`, `Popover` — semantic and
  whole-word.
- `Cue` — not an abbreviation, and the distinction it draws from `State`
  (browser-owned versus widget-owned) is load-bearing. With `Interactive()`
  covering hover, focus and press, it is now rarely reached; that is a reason to
  document it as an escape hatch, not to rename it.

### 12.6 Grammatical agreement

The option set mixed forms: `Fill()`, `Scrolls()`, `Fixed()`, `Flush()`,
`Clip()` — verb, third-person verb, adjective, adjective, verb.

The rule is **not** "every option is an imperative verb". `Interactive`,
`RevealedBy` and `As` are not verbs and should not be forced into one. It is
narrower, and achievable: **no option is a bare adjective** — an adjective
describes the element, while an option describes what is done to it, so an
adjective always leaves the reader to infer which property is being set.

So `Scrolls()` becomes `Scroll()`, `Fixed()` becomes `KeepSize()`, and `Flush()`
becomes `EdgeToEdge()` — a phrase rather than a verb, but one that names the
resulting shape instead of asking the reader to guess.

### 12.7 The two closest calls

The old `Of` → new `For`, and the old `On` → new `As`, are the weakest entries in
this pass, and are recorded as such.

Neither old name is an abbreviation or a collision; both are prepositions that
state no relation. `style.Of(name)` does not say what the sheet is *of*, and
`On(Panel)` does not say what is being put on what — the more so now that a
surface carries radius and padding as well as colour, so "on" describes less than
half of what it does.

The replacements state the relation: `style.For(w)` is a sheet **for** this
widget; `As(Panel)` styles this part **as** a panel. The gain is real but small,
and the old `On` is by some distance the most-typed function in the API, so this
is the largest churn in this document. Worth doing inside a release that is
already breaking; not worth a release of its own.

---

## 14. Why `100dvh` is allowed in `Cover` alone

**Decision.** The `Cover` primitive emits `height: 100dvh`, which is a viewport
unit and therefore the only literal the drift guard permits outside geometry
(`Size` percentages and `Aspect` fractions).

The height is **definite, not a floor**. `min-height` leaves the frame auto-sized,
so `Fill()` (`height: 100%`) resolves against nothing and `HideOverflow()` has no
box to clip — a tall child expands the shell and the whole application scrolls,
rail and header included. A shell whose content can exceed the viewport gives that
child `Scroll()`; the frame itself never moves.

An application shell is by definition sized against the viewport — there is no
container to be relative to, because it *is* the outermost container. `dvh` (not
`vh`) because `vh` is wrong on mobile browsers with retracting toolbars, which is a
bug the previous hand-written chassis carried.

This is a single value in a single primitive, not permission to widen the hole
to `vw`, `svh`, or any other viewport unit.

## 15. Why device scoping is a closed enum from `css`

**Decision.** `On()` takes a `css.Device`, never a free-form string or a
media-query expression.

A string parameter would reopen the entire surface that the closed scales exist to
shut. Every escaped media query would be a unique, unshared, untested expression
that the drift guard cannot validate. The `Device` enum in `css` owns the
thresholds, the overlap test, and the partition proof; here we only reference it.

## 16. Why `On()` is a last resort

**Decision.** The intrinsic primitives (`Split`, `Grid`, `Sidebar`) reflow on their
own without any query at all. Reach for one of those first. `On()` exists only when
the ARRANGEMENT itself differs between devices — e.g. a nav rail that becomes a
drawer on mobile.

## 17. Why `StateAttrs()` exists

**Decision.** `RevealedBy()` emits `display: none` and a state attribute selector
like `[data-open="true"]`. If the markup never writes `data-open`, the element is
invisible forever and nothing — not the compiler, not `Validate()`, not any test —
says a word. The two halves (Go writes attributes, Go+CSS writes selectors) live
in different build tags and cannot be checked statically. `StateAttrs()` gives the
consumer a list to assert against in a test that renders their markup, closing the
loop.

This does not weaken the architecture's claim of *class* agreement by construction
(§P-2 in TRADEOFFS.md). Classes are derived from `Name` and `Part`, which are
compile-time identifiers; state attributes are runtime choices the component makes,
and the only mechanical check is `Kind.Allows()`, which says a state *may* be used,
not that it *is*.

## 18. Why a state never changes the box size

**Decision.** A state rule (`When`, `Cue`, `CueWithin`) repaints a bordered
surface as a shadow ring — `box-shadow: 0 0 0 1px <border color>`, the border's
width a hair's breadth outside the box — instead of `border:`. The ring takes no
layout space.

This is a correctness rule, not a pixel-pushing detail. On the nav rail of
`platformd` the `+2px` border under `:hover` grew the item, pushed its siblings,
the pointer ended up outside the item, the `:hover` dropped, the item shrank, the
pointer entered again — a continuous flicker while the mouse sat still. The state
rule is what makes the element grow, and the growth is what destroys the state:
the two failure modes feed each other.

**Rejected: outline.** An `outline` also takes no layout space, and the first
implementation used it (`outline` + `outline-offset: -1px`). It was a two-fold
defect: outlines paint at the END of the stacking context, over the element's
positioned descendants — a Locked border crossed over the legend riding the same
line — and Safari < 16.4 ignores `border-radius` on outlines, so every state
border in the system rendered square.

**Rejected: box-shadow for the border alone.** It also takes no layout space, but
`Raise()` already owns that property, and two features colliding on one
declaration is exactly the kind of coupling that makes a sheet impossible to
reason about. The collision is unavoidable — the ring IS a box-shadow — so the
two compose in one place instead: `boxShadowDecls` is the package's only
box-shadow decision point, and when a state rule also raises, the ring and the
elevation merge into a single declaration, ring first. When a state rule only
raises (no border), it emits the bare elevation like any other raised rule.

**Consequence.** The base rule (`Root`, `Part`, `On`, `OnlyOn`) keeps `border:`
— the base box owns the border and pays its layout cost once, at rest.

## 19. Why `Deck` was replaced by `SlideDeck`

**Decision.** `Deck` laid its children out as a horizontal scroll-snap strip and
changed panel with `ScrollIntoView`. Its only consumer, `platformd`, nests it
with `MasterDetail` — another horizontal snap scroller — and two snap scrollers
on the same axis chain the scroll: when the inner strip reaches its end, the
browser continues on the outer one and the application changes module on its own.
`SlideDeck` keeps the slide but removes the scroller: the children are absolute
layers, the one carrying `widget.Current` sits in the box, and the rest wait
parked at the inline-start edge, entering left-to-right when activated.

**Rejected: keeping both.** A second way to achieve the same thing is exactly
what the closed API exists to remove. `Deck` had one consumer; the replacement
ships in the same release with the mapping in
[MIGRATION.md §9](MIGRATION.md#9-upgrading-to-v060---slidedeck-and-state-borders).

**Why the selector derives from `widget.Current`.** The state is not an option
because markup and CSS must agree by construction: the markup writes
`data-current="true"` through the state writer, and the sheet selects on
`Current.Attr()`. The same principle already sustains `RevealedBy`. A configurable
state would reintroduce the drift the attribute type exists to prevent.

**Why `visibility`, not `display`.** The parked pages are out of the tab order
and invisible, yet still transitionable. `display` is discrete — the page would
appear in its final position at the end of the transition, never arriving.
`visibility` transitions after the movement finishes, so the page is visible
while it slides in and removed from the tab order after it rests.

**Consequence.** The only horizontal snap scroller left in a shell is the one a
module owns itself, so the swipe gesture can never chain out of the content.

## 20. Why the animated reveal is `Animate` + `RevealedBy`, not a new option

**Decision.** No new public option: `Animate(m)` on the same rule as
`RevealedBy(st)` choreographs the swap as a fade (entry and exit, SPECS §5.2).
A new option would be a second way to express motion that `Animate` already
owns — the closed API removes those, it does not mint them. The precedent is
the chevron recipe: `Rotate()` on the base rule means nothing animated until
`Animate()` is beside it; the reveal works the same way.

**Why opacity only, no slide.** A slide needs a length — 4px, 8px, anything —
and lengths live in closed scales that have no step for "nudge an entering
panel". Inventing one literal would force every future motion to answer why it
cannot have its own. Opacity is unitless: the fade softens the swap with
nothing invented.

**Why `allow-discrete` + `@starting-style`, not the `Drawer` pattern.**
`Drawer` parks with transform because it is `position: fixed` — layout never
enters the question. An in-flow reveal must leave the layout when hidden, so
only `display` removes it, and only the discrete transition choreographs a
property that is discrete. Browsers without `@starting-style` ignore the block
and get the instant swap `RevealedBy` always emitted: the fallback is today's
behavior, not a breakage.

**Rejected: keyframes.** TRADEOFFS already settles this: keyframes are an
open-ended language, and components that need them fall back to application
CSS. A fade needs no keyframes, so the closed surface stays closed.

**Incidental fix in the same release: `OnlyOn` keeps its base hide.**
`OnlyOn` documents "display:none by default", but it applied the hide only
when the part rule did not exist yet — and base options read before device
ones, so any part with both a `Part()` block and `OnlyOn()` painted on
desktop (calendarslider's collapsed chip did exactly this). `OnlyOn` now
merges `hidden` unconditionally, the same way `Part()` merges. One sheet in
the wild combined the two, and its new output is the documented behavior, so
no migration entry is needed — and neither is one for the animated reveal
itself, which changes output only for sheets that pair the two options, a
combination no consumer used.

## 21. Why `FloatMiddle` is its own option, not a `Docked` extension

**Decision.** A floating control pinned to the vertical middle of a viewport
edge gets `FloatMiddle(side, gap)` — `position:fixed`, `top` at half the
viewport (`Size.Half`, never a literal), pulled back by half its own height,
off the edge by one `Space` step, at the widget kind's layer.

**Why not extend `Docked`.** A corner owns all four insets; a middle owns one
edge plus its own height. One option covering both would need a mode for
"which semantics", and modes are choices the author must get right — the
harness removes those, it does not parameterise them. `Docked` stays corners;
corners stay `Docked`'s only answer.

**Why not compose it downstream.** Centring needs `top:50%` plus a
`translateY(-50%)`, and `transform` already has three owners (`Rotate`,
`Drawer`, `OnEdge`'s straddle) with `Validate()` rejecting every second
claim. A hand-rolled middle would either fight one of them silently or
re-implement the guard. Viewport scope only, for the same reason
middle-centring inside a scrolling parent fights the scroll instead of
floating over it: the need is viewport chrome, and the option says so.

**Rejected: no option, document the trick.** Two declarations an author must
write together by hand is exactly the glue §"Lego pieces" moves into the
piece that owns it — every floating middle control would repeat it, and each
copy could drift.

## Related documents

- [ARCHITECTURE.md](ARCHITECTURE.md) — the structure these decisions produce.
- [SPECS.md](SPECS.md) — the exact behaviour they specify.
- [TRADEOFFS.md](TRADEOFFS.md) — what these decisions cost, and what they leave unsolved.
- [MIGRATION.md](MIGRATION.md) — what changes for a consumer.

---

## 13. Why `Split` abandons container queries

**Decision.** `Split` is responsive through intrinsic sizing — `flex-wrap` plus a
flex basis that is either very large or negative — and emits no `@container` rule
and no `container-type` declaration.

The published implementation set `container-type: inline-size` on the same
selector its `@container` rule targeted. An element is never its own query
container, so the rule never applied and `Split` had no responsive behaviour at
all. This was measured, not inferred; the numbers are in
[SPECS.md §4.1](SPECS.md#41-split-uses-no-container-query-deliberately).

**Rejected: the correct query form.** Making the query work requires a separate
ancestor element carrying `container-type`, with the split itself as its child.
That element does not exist in the anatomy, so the engine would have to require
markup it does not control — breaking the flat-specificity and no-DOM-coupling
guarantees in one move.

**Rejected: a media query.** It would make the component react to the viewport
rather than to its own width, so a split placed in a sidebar would lay out as if
it had the whole page. That is precisely the bug container queries exist to
prevent, and it is why the fix is intrinsic sizing rather than a fallback to
`@media`.

**Consequence.** After this change no primitive uses a query of any kind:
`Grid` was already intrinsic through `auto-fit`/`minmax`, and `Split` now is too.
The architecture's responsiveness claim is simpler and stronger for it — the
layout does not depend on any ancestor being a container, so a component cannot be
broken by where it is placed.
