---
PLAN: "feat(style): Rotate option for state-driven icon rotation"
TAG: v0.7.0
EXECUTOR: jules
REVIEWER: none
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.

# Plan — `style.Rotate`: rotación tipada para iconos que reaccionan a un estado

> **Idioma:** este documento está en español porque lo pidió el autor.
> **El código, los comentarios de código y los nombres de símbolos van SIEMPRE en
> inglés** — `webtyp/*` es librería pública. No traduzcas identificadores ni
> escribas comentarios en español dentro de los `.go`.

## Prerrequisito (ejecutar primero)

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

Se ejecuta `gotest` (nunca `go test`) desde la raíz del repo. No invoques
`gopush` ni `codejob`: son herramientas del desarrollador, fuera de este plan.

---

## 1. Contexto: por qué falta esta pieza

`webtyp/widget/style` es el DSL que emite el CSS de todos los componentes del
ecosistema. Hoy **no existe ninguna forma de rotar un elemento**. Verificable:

```bash
grep -rnE "^func (Rotate|Turn|Spin)" style/*.go   # → vacío hoy
```

El consumidor concreto que lo necesita es `webtyp/components/selectsearch`:
su chevron (`▼`) debe girar 180° cuando el desplegable se abre y volver a su
posición al cerrarse. Ese componente ya tiene el estado (`widget.Open`) y ya
tiene la parte (`PartIcon`), pero no tiene con qué expresar la rotación.

Según `CONSTRUCTION_HARNESS.md` (doctrina del ecosistema):

> *"A missing contract at a boundary is a defect in the library, not in the
> consumer."* — un consumidor **nunca** recrea localmente un símbolo que falta.

Por eso la pieza se añade aquí, en la librería dueña del concern, y no dentro
de `selectsearch`.

## 2. Qué construir

Un `Option` tipado por intención, con **enum cerrado** (principios 1, 3 y 4 del
harness: *typed over any*, *illegal states unrepresentable*, *one way to do
each thing*). Nada de aceptar grados libres como `string` o `int`: un chevron
sólo necesita pasos de cuarto de vuelta, y un valor arbitrario sería un hueco
genérico.

```go
// Turn is the closed set of rotations the DSL supports — quarter-turn steps,
// which is every rotation a chevron, caret or disclosure arrow needs. A free
// degree value is deliberately not accepted: it would be a generic hole with
// no intent, and no part in the ecosystem has ever needed one.
type Turn uint8

const (
	TurnNone Turn = iota // 0deg — the resting position
	TurnQuarter          // 90deg
	TurnHalf             // 180deg — a chevron flipped to point the other way
	TurnThreeQuarter     // 270deg
)

// Rotate turns the element by a fixed quarter-turn step. Pair it with a state
// rule — When()/WhenWithin() — so the rotation IS the state, and with
// Animate() on the base rule so the turn is a transition instead of a jump.
//
// It cannot be combined with OnEdge() or Drawer() on the same rule: both of
// those already own the element's `transform`, and a second one would silently
// replace the first. Validate() rejects the combination.
func Rotate(t Turn) Option
```

## 3. Archivos exactos

| Archivo | Acción |
|---|---|
| `style/scale.go` | Añadir el tipo `Turn` y sus 4 constantes, junto a los otros enums de escala (`Radius`, `Motion`, `Space`). |
| `style/except.go` | Añadir `func Rotate(t Turn) Option`. |
| `style/sheet.go` | Añadir a `type rule struct`: `hasRotate bool` y `rotate Turn`. |
| `style/emit_helpers.go` | Añadir `func turnValue(t Turn) string`. |
| `style/emit_decls.go` | Emitir la declaración; añadir `hasRotate` a `emitsNothing`. |
| `style/validate.go` | Rechazar `Rotate` + `OnEdge` / `Rotate` + `Drawer`. |
| `style/shell_test.go` | Tests nuevos (ver §5). |

### 3.1 `style/scale.go`

Añadir **al final del bloque de enums**, con este comentario exacto:

```go
// Turn is the closed set of rotations the DSL supports — quarter-turn steps,
// which is every rotation a chevron, caret or disclosure arrow needs. A free
// degree value is deliberately not accepted: it would be a generic hole with
// no intent, and no part in the ecosystem has ever needed one.
type Turn uint8

const (
	TurnNone Turn = iota // 0deg — the resting position
	TurnQuarter          // 90deg
	TurnHalf             // 180deg
	TurnThreeQuarter     // 270deg
)
```

### 3.2 `style/except.go`

Colocar `Rotate` **inmediatamente después** de `func Animate(m Motion) Option`
(son la pareja natural: uno declara el destino, el otro cómo se llega):

```go
// Rotate turns the element by a fixed quarter-turn step. Pair it with a state
// rule — When()/WhenWithin() — so the rotation IS the state, and with
// Animate() on the base rule so the turn is a transition instead of a jump.
//
// Not combinable with OnEdge() or Drawer(): both already own the element's
// `transform` and a second declaration would silently replace the first.
// Validate() rejects the combination rather than letting it fail on screen.
func Rotate(t Turn) Option {
	return func(r *rule) {
		r.hasRotate = true
		r.rotate = t
	}
}
```

### 3.3 `style/sheet.go`

Dentro de `type rule struct`, junto a los otros pares `hasX`/`x`:

```go
	hasRotate bool
	rotate    Turn
```

### 3.4 `style/emit_helpers.go`

```go
// turnValue maps a Turn to its CSS degrees. Quarter steps only — see Turn.
func turnValue(t Turn) string {
	switch t {
	case TurnQuarter:
		return "90deg"
	case TurnHalf:
		return "180deg"
	case TurnThreeQuarter:
		return "270deg"
	default:
		return "0deg"
	}
}
```

### 3.5 `style/emit_decls.go`

**(a)** En `genDecls`, emitir la rotación. Colocarla **justo después** del
bloque `if r.hasMotion { ... }` (que emite `transition:`), porque leerlas juntas
explica el par:

```go
	if r.hasRotate {
		decls = append(decls, "transform: rotate("+turnValue(r.rotate)+");")
	}
```

**(b)** En `func (r rule) emitsNothing(layer widget.Layer) bool`, añadir
`&& !r.hasRotate` a la cadena de negaciones. La línea completa queda:

```go
	return !r.hasFlow && !r.fill && !r.grow && !r.pushEnd && !r.scroll && !r.keepSize && !r.edgeToEdge && !r.hideOverflow && !r.hasIcon && !r.controlBox && !r.logoBox && !r.chipBox && !r.hasGlyph && !r.hasPadEdge && !r.hasChipSeat && !r.hasPadInline && !r.startContent && !r.shown && !r.hasRotate
```

> Sin esto, una regla que **sólo** tiene `Rotate()` se considera vacía y no se
> emite nunca. Es exactamente el modo de fallo silencioso que el harness
> prohíbe (principio 6).

### 3.6 `style/validate.go`

`Rotate` y `OnEdge`/`Drawer` compiten por la misma propiedad `transform`. Hoy
`OnEdge(EdgeTop)` emite `transform: translateY(-50%)` y `Drawer` emite
`transform: translateX(±100%)`. Verificable:

```bash
grep -n "transform:" style/emit_decls.go
```

Añadir, dentro del bucle que ya recorre las reglas en `Validate()` (el mismo
que valida Anchor/Docked/OnEdge/Flyout/Backdrop/Drawer — busca el mensaje
`"all set position; use one"` para localizarlo), esta comprobación por regla:

```go
		if r.hasRotate && (r.hasOnEdge || r.hasDrawer) {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Rotate cannot combine with OnEdge/Drawer — both own transform", string(s.widget.WidgetName()), string(p)))
		}
```

El mensaje va **literal**, entre backticks en el código como constante de
error formateada con `fmt.Errf` (paquete `webtyp.com/fmt`, ya
importado en ese archivo — **no** el `fmt` estándar).

## 4. Reglas de calidad obligatorias

- **Sin strings sueltos en la lógica.** Los valores CSS ya están centralizados
  (`turnValue`); no repitas `"180deg"` en ningún otro sitio.
- **Sin librería estándar en código compartido con WASM.** Este paquete usa
  `webtyp.com/fmt`, nunca `errors`/`strconv`/`strings` del stdlib.
  Comprueba el bloque de imports del archivo antes de escribir.
  *Anti-footgun:* `style/emit_*.go` llevan `//go:build !wasm` y **sí** pueden
  usar `sort` del stdlib — ya lo hacen. No "arregles" esos imports.
- **Superficie mínima.** Exporta `Turn`, sus 4 constantes y `Rotate`. Nada más.
  `turnValue` queda sin exportar.
- **Determinismo.** La suite tiene tests que comparan la salida byte a byte.
  No reordenes declaraciones existentes; añade la tuya en la posición indicada.

## 5. Tests (en `style/shell_test.go`)

Los tests de este repo son *consumer-shaped*: construyen una `style.Sheet` real
como lo haría un `css.go`, renderizan y afirman sobre el CSS. Sigue ese molde
(mira `TestOnEdgeStraddlesTheLine` en el mismo archivo).

**Test 1 — la rotación se emite y es un estado:**

```go
func TestRotateTurnsOnState(t *testing.T) {
	// A chevron points down at rest and flips when its host opens. The
	// rotation IS the state: a When() rule, not a class the component toggles
	// by hand, and Animate() on the base rule makes the flip a transition.
	w := testWidget{name: "ss", kind: widget.Combobox}
	s := style.For(w).
		Root(style.Anchor()).
		Part("icon", style.Rotate(style.TurnNone), style.Animate(style.MotionBase)).
		WhenWithin(widget.Open, "", "icon", style.Rotate(style.TurnHalf)).
		Stylesheet().String()

	if !strings.Contains(s, "transform: rotate(0deg);") {
		t.Errorf("expected the resting rotation, got:\n%s", s)
	}
	if !strings.Contains(s, "transform: rotate(180deg);") {
		t.Errorf("expected the open-state rotation, got:\n%s", s)
	}
	if !strings.Contains(s, `.ss[data-open="true"] .ss__icon`) {
		t.Errorf("expected the state to reach the icon from the root, got:\n%s", s)
	}
	if !strings.Contains(s, "transition:") {
		t.Errorf("expected Animate to make the turn a transition, got:\n%s", s)
	}
}
```

**Test 2 — una regla que sólo rota sí se emite:**

```go
func TestRotateAloneIsNotDroppedAsEmpty(t *testing.T) {
	// emitsNothing() decides whether a rule is worth a selector. A part whose
	// only declaration is a rotation is a real rule; forgetting it there is a
	// silent failure — nothing renders and nothing complains.
	w := testWidget{name: "ss", kind: widget.Combobox}
	s := style.For(w).Part("icon", style.Rotate(style.TurnQuarter)).Stylesheet().String()
	if !strings.Contains(s, ".ss__icon") || !strings.Contains(s, "transform: rotate(90deg);") {
		t.Errorf("a Rotate-only rule must still be emitted, got:\n%s", s)
	}
}
```

**Test 3 — la combinación prohibida se rechaza:**

```go
func TestRotateRejectsTransformConflicts(t *testing.T) {
	// OnEdge and Drawer already own `transform`. Two declarations on one rule
	// means the later silently wins — a defect that is invisible in the CSS
	// and only shows on screen. Validate() is where it must die.
	w := testWidget{name: "ss", kind: widget.Combobox}
	sheet := style.For(w).
		Root(style.Anchor()).
		Part("chip", style.OnEdge(style.EdgeTop, style.SideStart, style.SpaceNone, style.Space2), style.Rotate(style.TurnHalf))
	if errs := sheet.Validate(); len(errs) == 0 {
		t.Fatal("expected Rotate + OnEdge on one rule to be rejected")
	}
}
```

## 6. Criterios de aceptación (verificables)

```bash
gotest                                              # vet ✅ race ✅ tests ✅
grep -rn "func Rotate(t Turn) Option" style/        # 1 resultado
grep -rn "TurnNone\|TurnHalf" style/scale.go        # las 4 constantes
grep -rn "hasRotate" style/emit_decls.go            # emisión + emitsNothing
grep -c "!r.hasRotate" style/emit_decls.go          # 1 (en emitsNothing)
grep -rn "Rotate cannot combine" style/validate.go  # 1 resultado
```

Además, **ningún test existente puede cambiar de resultado**: la suite incluye
comprobaciones de determinismo byte a byte. Si alguno falla, la causa es haber
movido una declaración existente, no el añadido.

## 7. Fuera de alcance (NO hacer)

- No toques `OnEdge`, `ChipSeat`, `Drawer` ni ninguna emisión de `transform`
  existente. Sólo se añade `Rotate` y su validación.
- No añadas un `Rotate` con grados libres, ni un `RotateDeg(int)`, ni un
  segundo camino para lo mismo (principio 4: *one way to do each thing*).
- No modifiques `components/` ni `layout/`: van en sus propios planes.

## 8. Etapas

| # | Etapa | Archivos | Cierra cuando |
|---|---|---|---|
| 1 | Tipo y opción | `style/scale.go`, `style/except.go`, `style/sheet.go` | compila; `grep "func Rotate"` da 1 |
| 2 | Emisión | `style/emit_helpers.go`, `style/emit_decls.go` | Test 1 y Test 2 pasan |
| 3 | Validación | `style/validate.go` | Test 3 pasa |
| 4 | Suite completa | — | `gotest` verde, sin regresiones de determinismo |

Etapa 1 es **gate** de la 2; la 3 es independiente de la 2.

## 9. Este plan es gate de otros

`components/docs/PLAN.md` (selectsearch) **no puede empezar** hasta que este
esté publicado: consume `style.Rotate` y `style.TurnHalf`.

Referencias externas (sólo lectura opcional; lo crítico ya está inline):

- Doctrina: <https://github.com/webtyp/app-releases/blob/main/docs/CONSTRUCTION_HARNESS.md>
- Consumidor: <https://github.com/webtyp/components/blob/main/selectsearch/css.go>
