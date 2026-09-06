# Module boundaries

How a value reaches the browser, and which module is allowed to decide what.
Context in [ARCHITECTURE.md §2](../ARCHITECTURE.md#2-position-in-the-suite).

```mermaid
flowchart TD
    A[webtyp/css<br/>owns VALUES<br/>token catalog, light/dark, contrast test]
    B[webtyp/widget<br/>owns DECISIONS<br/>which token, which part, which state]
    C[widget/style<br/>emits scoped CSS<br/>var references only, never literals]
    D[webtyp/ssr<br/>owns DELIVERY<br/>calls RenderCSS on a zero value]
    E[merged stylesheet<br/>layer statement hoisted, duplicates merged]
    F[browser<br/>:root declarations resolve every var]

    A -->|token references| C
    B -->|Name, Part, State, Kind| C
    C -->|css.Stylesheet per widget| D
    D --> E
    A -->|RootCSS declares the variables| E
    E --> F
```

The one rule that keeps the split honest: **`widget` never invents a value.**
Every emitted value is a reference to a token `css` declares. The drift guard in
the test suite enforces it — see
[SPECS.md §7.1](../SPECS.md#71-global-invariants).

## WASM boundary

```mermaid
flowchart TD
    A[package widget<br/>identity only<br/>Name, Part, Class, Kind, State, Cue]
    B[package widget/style<br/>go:build !wasm<br/>scales, surfaces, emission]
    C[WASM binary<br/>ships to the client]
    D[build-time SSR<br/>server side only]

    A --> C
    A --> B
    B --> D
    B -.->|must never reach| C
```

`widget` imports only `webtyp/fmt`. It must not import `webtyp/css`, which is
why `Kind.Layer()` returns an enum and `widget/style` maps it to the `--z-*`
catalog. A test asserts `widget/style` is absent from a consumer's WASM
dependency graph.
