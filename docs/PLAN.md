---
PLAN: "feat(kind): a Form may hold Open — a form with a conditional section"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **This is a GATE.** `webtyp/components` cannot land its `scheduleeditor` fix
> until this ships. Orchestrator:
> [webtyp/docs/AGENDA_VIEW_FIXES_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_VIEW_FIXES_MASTER_PLAN.md).

# Plan — `Kind.Allows(Form, Open)`

## 1. Context you need (this repo, zero assumptions)

`webtyp.com/widget` owns the widget vocabulary: a `Name`, a `Part`, a `Kind`,
and a `State`. `Kind.Allows(State) bool` answers "is this state meaningful for a
widget of this kind?" — and `widget/style`'s `Sheet.validateStates` plus the
`components` repo's `TestKindAllowsEveryState` both enforce the answer.

The file is [`kind.go`](../kind.go). The relevant function today:

```go
func (k Kind) Allows(s State) bool {
	// Universal set, allowed for every Kind
	if s == Disabled || s == Locked || s == Busy {
		return true
	}

	switch k {
	case Listbox, Tabs, Grid:
		return s == Selected || s == Current
	case Menu:
		return s == Open || s == Current
	case Dialog, Disclosure, Alert:
		return s == Open
	case Combobox:
		return s == Open || s == Selected || s == Invalid
	case Form:
		return s == Invalid // Busy is already covered by universal set
	default:
		return false
	}
}
```

## 2. Why the change

A `Form` routinely contains a section that appears only under a condition: an
"other, please specify" text box, a shipping address revealed by a checkbox, an
hours window that applies to two of three radio choices but not the third.

`webtyp.com/components/scheduleeditor` is exactly this. It is a `Form` whose
per-date exception panel reveals an add-form once a calendar day is picked, and
reveals an hour window for two of its three exception types. It already writes
`data-open` from Go — and its `css.go` cannot declare the matching
`RevealedBy(widget.Open)` rule, because `Allows` says no. The result shipped in
`components v0.6.20`: **the attribute is written and never styled, so the form
is permanently visible** — the silent, unverifiable failure
`docs/DESIGN.md` §17 exists to warn about.

The two escapes are both worse:

- Re-declaring the component `Disclosure` buys `Open` and **loses `Invalid`**,
  which the same stylesheet already uses for its invalid weekly row.
- Hand-rolling visibility with a `SignalNodes` subtree rebuild abandons the
  reveal mechanism this package designed, in the one component that needs it.

`Open` on a `Form` is not a new concept — it is the concept `Disclosure` already
carries, applied to a container that is also a form. Nothing new is exported.

## 3. Design gate

Not required: this plan **exports no new symbol and changes no signature**. It
widens the return of an existing predicate. Recorded for the reviewer:

| Ledger row | Δ |
|---|---|
| Concepts the developer must learn | **0** — `Open` and `Form` both already exist |
| Files they must touch to do X | **0** |
| Lines at the call site | **0** |
| Ways to do the same thing | **−1** — the `Disclosure`-instead-of-`Form` workaround and the hand-rolled-rebuild workaround both stop being reachable answers |
| Exported surface | **0** |

## 4. Stage 1 — the predicate

**File: [`kind.go`](../kind.go)** — replace the `Form` case:

```go
	case Form:
		// A form with a conditional section: an "other, please specify" box, a
		// panel revealed by a choice. Open is the same reveal a Disclosure
		// carries; a Form that also holds Invalid must not have to give it up
		// to get it. Busy is already covered by the universal set.
		return s == Invalid || s == Open
```

Change nothing else. Do **not** touch `Listbox`, `Tabs`, `Grid`, `Menu`,
`Dialog`, `Disclosure`, `Alert`, `Combobox` or `default`.

## 5. Stage 2 — the tests

**File: [`kind_test.go`](../kind_test.go)** (create it if absent; if a
`TestKindAllows`-shaped test already exists, extend it rather than adding a
second one — two tests over one predicate is the duplication this repo forbids).

Add, using only `testing` and this package:

```go
// A Form holds Invalid AND Open: a form with a conditional section is
// ordinary, and it must not surrender its validation state to get the reveal.
func TestFormAllowsInvalidAndOpen(t *testing.T) {
	for _, s := range []State{Invalid, Open} {
		if !Form.Allows(s) {
			t.Errorf("Form.Allows(%s) = false, want true", s.String())
		}
	}
}

// Widening Form must not widen anything else: Selected and Current stay out.
func TestFormStillRejectsSelectionStates(t *testing.T) {
	for _, s := range []State{Selected, Current} {
		if Form.Allows(s) {
			t.Errorf("Form.Allows(%s) = true, want false", s.String())
		}
	}
}

// The kinds this change does not touch keep their exact answer for Open.
func TestOpenAllowanceUnchangedForOtherKinds(t *testing.T) {
	allowed := []Kind{Menu, Dialog, Disclosure, Alert, Combobox}
	for _, k := range allowed {
		if !k.Allows(Open) {
			t.Errorf("%s.Allows(Open) = false, want true", k.String())
		}
	}
	for _, k := range []Kind{Listbox, Tabs, Grid} {
		if k.Allows(Open) {
			t.Errorf("%s.Allows(Open) = true, want false", k.String())
		}
	}
}
```

Verify `State.String()` and `Kind.String()` exist with those names before
relying on them in the error messages — both are in this package
([`state.go`](../state.go), [`kind.go`](../kind.go)). If a name differs, use the
one that is actually there; do not add a stringer.

## 6. Stage 3 — the document that states the rule

**File: [`docs/DESIGN.md`](DESIGN.md)** — this repo's own rule is that a design
decision is written down, and that a rule lives in exactly one place. Find the
section that explains `Kind.Allows` (search for `Allows`). Add one sentence to
the existing prose stating that `Form` holds `Open` for a conditional section,
alongside `Invalid`. **Do not create a new section**, and do not restate it
anywhere else in the repo — two copies drift.

If no section discusses `Allows`, add the sentence to §17 ("Why `StateAttrs()`
exists"), which already discusses `RevealedBy` and `data-open`.

## 7. Constraints — read before writing code

- **No standard library in this package.** Use `webtyp.com/fmt`, never
  `strconv`, `strings` or `errors`. `_test.go` files may import `testing`; that
  is the only exception and it is already the norm here.
- This package is compiled into WASM. Nothing added here may allocate a `map`
  or pull a new import — the change is one boolean expression.
- **No `TODO`, no commented-out code, no deprecated path.** Before closing run
  `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .` and confirm every hit
  predates this change.
- Run `gotest`, never `go test`.

## 8. Acceptance criteria

| # | Check | Expected |
|---|-------|----------|
| 1 | `gotest ./...` | all green |
| 2 | `grep -n "case Form:" -A 4 kind.go` | shows `return s == Invalid \|\| s == Open` |
| 3 | `grep -rn "case Menu:\|case Dialog, Disclosure, Alert:\|case Combobox:" kind.go` | byte-identical to before this change |
| 4 | `grep -rn "Open" docs/DESIGN.md` | one new sentence tying `Form` to `Open`, in an existing section |
| 5 | `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .` | no hit introduced by this change |

## 9. Stages

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | Widen the predicate | `kind.go` | `Form.Allows(Open)` is true; every other kind unchanged |
| 2 | Tests | `kind_test.go` | the three tests above pass |
| 3 | Document the rule | `docs/DESIGN.md` | one sentence, one place |

## 10. What this deletes

Two workarounds, neither of which will be written now: declaring a form-shaped
widget as `Disclosure` to buy `Open`, and hand-rolling reveal with a subtree
rebuild. No code is removed, because the defect this fixes is code that was
never allowed to be written.
