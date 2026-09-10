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

---

# Registro — lo que siguió a `Button`, en el mismo hilo

`Button` (§1-§10, shipped `v0.6.26`) arregló los botones, pero la vista que
motivó el plan seguía sin ser usable: sus selectores de día eran checkboxes
nativos. Dos ampliaciones más de este paquete lo cerraron. Ambas se ejecutaron
en local y se publicaron; se registran aquí porque el repo no debe tener un
símbolo público sin la razón que lo justifica.

## A. `style.VisuallyHidden()` — `v0.6.27`

**El defecto.** Un componente que quiere pintar un checkbox nativo no podía: el
input trae su propia piel y no se estiliza, y la única forma de sacarlo de la
vista era `Hide()`, que lo saca **también** del orden de tabulación y del árbol
de accesibilidad. La elección era entre feo y accesible, y
`grep -rn 'clip\|sr-only\|visually' style/` no devolvía nada.

**Design gate.** *Prior art:* `.visually-hidden` de Bootstrap 5, `sr-only` de
Tailwind, `.cdk-visually-hidden` de Angular Material — el patrón tiene nombre
estable en todos, y `VisuallyHidden` es literalmente el de Bootstrap; inventar
otro costaría una búsqueda por cero ganancia. *Ledger:* +1 concepto, +1 símbolo;
**−1 forma de hacerlo mal** (el emparejamiento input-oculto/label-pastilla deja
de exigir elegir entre accesibilidad y estética). *Dónde:* `widget/style`, junto
a `Hide()`/`Show()`, que es la familia de la que es el complemento. *Qué borra:*
nada — capacidad nueva, y el plan lo dice en vez de fingir lo contrario.

**Emisión:** la receta estándar — `position:absolute`, caja de 1px,
`clip-path: inset(50%)`, `overflow:hidden` — y **nunca** `display:none`, que es
justamente lo que la haría inútil. `Validate()` rechaza `Hide()` a su lado, y
`VisuallyHidden` entra al conjunto que ya se disputaba `position`.

## B. `Kind.Allows(Form, Selected)` — `v0.6.28`

Mismo movimiento que trajo `Open` a `Form`, un control más abajo: un formulario
suele contener un juego de chips donde uno queda elegido — un selector de días,
un juego de filtros. Ese es el estado que `Listbox` ya carga, y re-declarar todo
el formulario como `Listbox` para conseguirlo le costaría `Invalid` y `Open`,
que su misma hoja de estilos ya usa.

No exporta ningún símbolo: amplía el retorno de un predicado existente. `Current`
sigue fuera — significa "en la que estás" en una navegación, y un formulario no
tiene esa noción. El test positivo vive en un solo lugar
(`TestFormAllowsInvalidOpenAndSelected`); dos tests sobre un predicado es la
duplicación que este repo prohíbe.


---

# Fase B — la familia de interacción se separa de la superficie

> **Fase B (GATE)** de
> [`webtyp/docs/TYPED_EVENTS_AND_SURFACES_MASTER_PLAN.md`](https://github.com/webtyp/webtyp/blob/main/docs/TYPED_EVENTS_AND_SURFACES_MASTER_PLAN.md).
> Independiente de la Fase A; `components` consume las dos.
>
> Escrito 2026-09-10; **corregido** tras auditar contra
> `webtyp/app-releases/docs/CONSTRUCTION_HARNESS.md`. Nada está implementado.
> Todo dato lleva su `archivo:línea` para comprobarlo a mano.

## 0. El gate de api-design — las cinco respuestas

El único cambio de API es el **contrato** de `Interactive(s Surface)`: deja de
pintar la superficie de reposo y pasa a declarar solo la familia de la que se
derivan los estados. La firma no cambia; el significado sí, así que el gate
aplica.

### 1. Prior art

| Sistema | Reposo e interacción | ¿Deriva uno del otro? |
|---|---|---|
| **Material Design 3** | *container color* y *state layer* son tokens distintos | no |
| **Tailwind** | `bg-transparent hover:bg-slate-200` — utilidades independientes | no |
| **Radix Themes** | escala indexada: `--accent-3` reposo, `--accent-4/5` hover/active | no, se indexa |
| **Bootstrap** | `.btn-outline-primary` — el par reposo/hover se nombra como variante | no |

El invariante: **ninguno deriva el color de interacción oscureciendo el de
reposo.** O nombran los dos, o indexan una escala.

**Por qué este ecosistema difiere, y por qué está bien.** Derivar es una
simplificación real: un token y los estados salen solos, sin que cada componente
elija tres colores. No se abandona. Lo que se corrige es que hoy la **fuente**
de esa derivación se decide por accidente — el último de `As()`/`Interactive()`
que se haya escrito — en vez de declararse.

### 2. El test del nombre novato

```go
Part(PartRow,
    style.As(style.Subtle),          // píntalo apagado y transparente
    style.Interactive(style.Page),   // sus estados salen de la familia Page
)
```

Leído por alguien sin contexto: *"píntalo así, y sus estados salen de allá"*.
Dos frases, dos decisiones. Hoy esas mismas dos líneas colapsan en una y la
segunda gana en silencio. No se inventa vocabulario: `As` e `Interactive` ya son
las palabras de este paquete.

### 3. El ledger de complejidad

```
Conceptos a aprender            +1 (superficie y familia son cosas distintas) / 0
Líneas en el call site          +6                                           / 0   ← PEOR
Archivos a tocar para hacer X   +0                                           / 0
Formas de hacer lo mismo         0                                           / −1  (se borra el "Interactive también pinta")
Modos de fallo silencioso        0                                           / −2  (D1 y D2 pasan a diagnóstico)
```

**La fila que empeora son 6 líneas** en todo el ecosistema: de 10 bloques con
`Interactive()`, 4 ya declaran `As()` y 6 dependen hoy del pintado implícito.
Esos 6 ganan una línea explícita. Es el precio de que las otras dos filas bajen.

### 4. Dónde va

`widget/style`. Es el dueño de las recetas — la tabla de capas de
`BUTTON_SYSTEM_MASTER_PLAN.md` §2 ya lo fija: `css` los valores, `widget/style`
las recetas, `components` solo ensambla. D3 existe justamente porque una regla de
receta terminó viviendo en el consumidor.

### 5. Qué borra este cambio

- La regla implícita "`Interactive(s)` además pinta `s`". Se borra, no se
  deprecia.
- `TestNoHandRolledIconCaps` en `components/conformance_test.go`: la regla se
  muda a `Validate()`, no se duplica.

>
> **Qué cambió respecto de la primera versión.** Ordenaba los defectos por
> ratio de contraste — o sea, por lo que se ve. El harness ordena por **modo de
> fallo**: *"compile error → loud development diagnostic → (never) silent
> failure"*. Reordenado con ese criterio, el defecto que motivó todo esto (un
> hover ilegible) resulta ser el **menos** grave, y el que nadie vio nunca es el
> más grave. También agrego dos violaciones que introduje yo en `IconCap()` y
> que la primera versión no auditaba.

## 1. Los defectos, ordenados por el criterio del harness

| # | Defecto | Modo de fallo | ¿Se ve? |
|---|---|---|---|
| **D1** | `Interactive(X)` y `As(Y)` escriben el mismo `r.surface`; **gana el último, sin aviso** | **silencioso** | no |
| **D2** | `IconCap()` en el cap + `IconBox()` en el glifo: la regla del cap gana por especificidad, la del glifo se ignora | **silencioso** | no |
| **D3** | La receta `IconCap()` se hace obligatoria con un test en `components` — el **consumidor**, no la librería dueña | sin cobertura fuera de un repo | no |
| **D4** | `familyBase(Subtle)` devuelve `css.ColorMuted`, un token de **texto**, y de ahí salen los fondos de hover/focus/press | valor incorrecto, **visible** | sí |

D4 es el que reportó el usuario. Es el último de la lista porque produce un
resultado *visible y equivocado*, que es el modo de fallo menos malo de los
cuatro. D1 y D2 no producen nada: ni error, ni diagnóstico, ni efecto.

## 2. D1 — la colisión silenciosa (el defecto principal)

**Los hechos**

| # | Hecho | Dónde |
|---|---|---|
| 1 | `Interactive(s)` escribe `r.hasSurface, r.surface, r.interactive`. **No guarda la familia en un campo propio.** | `style/surface.go:218-223` |
| 2 | `As(s)` escribe el mismo `r.surface`. | `style/surface.go` |
| 3 | El emisor de estados hace `base := familyBase(r.surface)`. | `style/emit_states.go:142` |

**La prueba de que engaña.** `components/targethour/css.go:43-45` escribe
`Interactive(style.Page)` y dos líneas después `As(style.Subtle)`. El comentario
del propio autor, ahí mismo, dice *"Kept as a side inset so the row keeps its
Interactive(Page) surface over the whole box"*. Creía que la familia `Page`
sobrevivía. No sobrevive. El DSL lo contradijo y no dijo nada.

**Por qué es el defecto principal, según el harness**

- *"Silent failures. Cases where misuse produces neither an error nor a visible
  effect → turn them into compile errors; if that is impossible, into a loud
  development diagnostic."*
- Principio 3, *"Illegal states unrepresentable. One intent = exactly one path,
  typed to demand what it needs"*: "reposo Subtle, interacción derivada de Page"
  es una intención legítima — un botón fantasma — y **hoy no tiene ningún
  camino**. No es que el autor se haya equivocado: escribió lo único que se
  parecía a lo que quería.
- *"A missing contract at a boundary is a defect in the library, not in the
  consumer."* El contrato que falta es el de "familia de interacción", que hoy
  no existe como concepto separado de "superficie".

**El error de razonamiento de la primera versión de este plan.** Decía que D1
"no bloquea, porque con la salida A targethour queda bien igual". Ese es
exactamente el razonamiento que el harness prohíbe: argumenta desde *este
consumidor queda bien de casualidad* en vez de desde *el próximo consumidor no
recibe ninguna señal*.

## 3. D2 y D3 — lo que introduje yo en `IconCap()`

`IconCap()` (`style/except.go`) hace que el cap emita su propia regla
`> svg { width: 50%; height: 50% }`, para que el tamaño del glifo deje de ser
una decisión por componente. Cierra el agujero que produjo el bug original —
cuatro componentes, tres tamaños distintos para una caja idéntica de 50px — pero
abre dos.

**D2 — comprobado, no supuesto.** Con `Part("cap", IconCap())` y
`Part("glyph", IconBox(IconLg))` en la misma hoja:

```
VALIDATE ACEPTA — sin error ni diagnóstico
.probe__cap > svg { width: 50%; ... }     ← especificidad (0,1,1)
.probe__glyph    { width: 2.5em; ... }    ← especificidad (0,1,0), se ignora
```

Mismo `@layer primitives`, así que decide la especificidad y gana el cap. El
autor escribe `IconBox(IconLg)`, no ve error, no ve efecto, y no tiene forma de
saber por qué.

Y mi propio comentario de `IconCap()` dice: *"it just must not re-declare
IconBox"*. El harness nombra eso literalmente: *"'remember to call…' / 'don't
forget…' — if it must be remembered, it is a hole in the harness."*

**D3 — la guarda está en el repo equivocado.** `TestNoHandRolledIconCaps`
(`components/conformance_test.go`) recorre los `css.go` con una expresión
regular buscando `MediaBox(AspectSquare) + ControlBox()`. Copia el precedente de
`TestNoHandRolledButtons`, así que es consistente con la casa — pero por el
harness está mal ubicada:

- *"The glue is written once, in the library that owns it."*
- Solo protege a `webtyp/components`. Cualquier otra app que ensamble widgets no
  recibe nada.
- `Validate()` **ya es** el "loud development diagnostic" de esta librería
  (panica en `Stylesheet()`). Es ahí donde vive esta regla, y desde ahí cubre a
  todo consumidor, no a un repo.

## 4. D4 — el valor incorrecto, y lo que medí

`Subtle` como superficie es `{bg: "transparent", text: ColorMuted}`
(`style/surface.go:183-187`): es un tratamiento de **texto**, no tiene fondo
propio. Pero `familyBase(Subtle)` devuelve `css.ColorMuted`
(`style/emit_values.go:264-265`), así que los estados oscurecen un color de
texto y lo usan de fondo, debajo de ese mismo texto.

Ratios WCAG reales, tokens en claro
(`--color-muted #6E6E73` · `--color-surface #F2F2F7` · `--color-on-surface #1C1C1E`):

| Estado | Texto | Fondo | Ratio | |
|---|---|---|---|---|
| Reposo | `#6E6E73` | `#F2F2F7` | **4.54:1** | pasa AA |
| Hover (hoy) | `#6E6E73` | `#57575b` | **1.42:1** | falla |
| Focus (hoy) | `#6E6E73` | `#414145` | **2.00:1** | falla |
| Press (hoy) | `#6E6E73` | `#2d2d2f` | **2.71:1** | falla |

El dato que importa no es que falle: es que **el reposo pasa y la interacción
empeora**. Pasar el puntero por encima de un control lo vuelve ilegible.

Corrección mínima probada en local (`familyBase(Subtle) → css.ColorSurface`),
revertida antes de escribir esto:

| Estado | Texto | Fondo | Ratio | |
|---|---|---|---|---|
| Hover | `#6E6E73` | `#c3c3c7` | **2.89:1** | sigue fallando |
| Hover + texto `on-surface` | `#1C1C1E` | `#c3c3c7` | **9.68:1** | pasa AA |

Cambiar la familia **no alcanza**: mientras `Subtle` mantenga el texto apagado y
el fondo se oscurezca, el contraste sigue cayendo.

Tres sitios afectados, y dos son usos que el framework bendice:
`components/calendarslider/css.go` (el que se reportó, ya migrado a
`Interactive(Panel)`), `components/targethour/css.go:43-45`, y
`components/usermenu/css.go:23` con `style.Button(style.Subtle)` — mientras el
doc de `Button()` en `style/except.go:165-170` lista textualmente
*"s paints it — Primary, Secondary, Danger, Subtle"*.

## 5. Salidas, evaluadas contra los principios

| Salida | D1 | D2 | D3 | D4 | Veredicto |
|---|---|---|---|---|---|
| **A** — solo `familyBase(Subtle) → ColorSurface` | ✗ | ✗ | ✗ | parcial | insuficiente: deja los dos fallos silenciosos |
| **B** — A + texto `on-surface` en los estados | ✗ | ✗ | ✗ | ✓ | llega a AA, pero no toca ningún fallo silencioso |
| **C** — prohibir `Subtle` interactivo con una guarda | ✗ | ✗ | ✗ | ✓ | trata el síntoma: `familyBase` sigue devolviendo un token de texto, y hay que sacar `Subtle` del doc de `Button()` |
| **D** — campo propio para la familia de interacción, + B, + `Validate()` rechaza `IconBox` bajo un `IconCap`, + mover la guarda de caps a `Validate()` | ✓ | ✓ | ✓ | ✓ | la única alineada con el principio 6 |

La primera versión de este plan recomendaba **A** y anotaba D1 como "aparte".
Corregido: **D**. Las tres primeras arreglan lo que se ve y dejan intacto lo que
no se ve, que es justo el orden inverso al que pide el harness.

**Nota sobre el alcance.** D no es "más trabajo por prolijidad": D1 y D2 son
fallos silenciosos, y este documento existe porque el harness dice que ésos son
los que nunca hay que dejar pasar. Si aun así hay que recortar, el corte honesto
es **B + el diagnóstico de D2**, dejando D1 anotado — pero anotado como deuda
declarada, no como "no bloquea".

## 6. El test ácido, aplicado a esta sesión

El documento cierra con esto:

> *"And if that agent has to **ask a question** to proceed — 'is it acceptable to
> declare this interface locally?' — the harness has already failed by its own
> definition: the signature did not guide, and the compiler rejected the correct
> intent."*

En esta sesión hubo que preguntar **qué superficie usar** para arreglar el
hover. La firma no guió. Por la propia definición del repo, eso ya es la prueba
de que el harness está abierto acá — independientemente de qué salida se elija.


## 7. Etapas

| # | Etapa | Archivos | Listo cuando |
|---|---|---|---|
| 1 | Campo propio para la familia | `style/sheet.go`, `style/surface.go` | `Interactive(s)` escribe `interactiveFamily`, **no** `surface`; `As(s)` sigue escribiendo solo `surface` |
| 2 | El emisor lee el campo nuevo | `style/emit_states.go:142` | `familyBase(r.interactiveFamily)`; sin familia declarada no se emiten estados |
| 3 | D4 — el valor | `style/emit_values.go`, `style/surface.go` | `familyBase(Subtle)` → `ColorSurface`, y `Subtle` pasa su texto a `on-surface` en hover/focus/press; hover ≥ 4.5:1 verificado en test |
| 4 | D2 — diagnóstico | `style/validate*.go` | `IconBox` en una parte bajo un `IconCap` es rechazado con mensaje explícito |
| 5 | D3 — la guarda se muda | `style/validate*.go` | `MediaBox(AspectSquare)` + `ControlBox()` juntos son rechazados y apuntan a `IconCap()` |
| 6 | Test con forma de consumidor | `style/consumer_test.go` | una hoja real declara `As(Subtle) + Interactive(Page)` y se asserta que reposo y estados vienen de familias distintas |
| 7 | Docs | `docs/` | ninguna prosa que repita lo que la firma ya obliga |

`gotest` verde en cada etapa. El tag se publica recién al cerrar la 7.

## 8. Cero deuda técnica — chequeo de cierre

- `grep -rn "interactive.*surface" style/` no muestra ningún punto donde una
  familia se infiera de la superficie.
- Ningún test acepta `IconCap()` junto a `IconBox()`.
- `components/conformance_test.go` ya no contiene `TestNoHandRolledIconCaps`
  (se borra en la Fase C, no se deja duplicada).
- Sin `TODO`, sin shim de compatibilidad, sin una vía que restaure el
  comportamiento viejo.
