# `webtyp/widget`
<img src="docs/img/badges.svg">

Visual component contracts, states, layout, and styling for the `webtyp` suite.

## Usage

```go
package main

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

type MasterDetail struct{}

func (m *MasterDetail) WidgetName() widget.Name { return widget.Name("masterdetail") }
func (m *MasterDetail) WidgetKind() widget.Kind { return widget.Grid }

func (m *MasterDetail) Style() *style.Sheet {
	return style.For(m).
		Root(style.Grid(style.ColumnNarrow, style.Space2), style.As(style.Page), style.Scroll()).
		Part("master", style.Stack(style.Space1), style.As(style.Panel), style.Pad(style.Space3)).
		Part("detail", style.Stack(style.Space2), style.As(style.Panel), style.Pad(style.Space3)).
		Part("item",   style.Row(style.Space1),   style.As(style.Subtle), style.Interactive(style.Subtle)).
		When(widget.Selected, "item", style.As(style.Highlight))
}

func main() {
	m := &MasterDetail{}
	fmt.Println(m.Style().Stylesheet().String())
}
```

## Documentation

- [AGENTS.md](AGENTS.md) — constraints for anyone (human or agent) changing this library: the WASM boundary, zero escape, never invent a value, and the SSR contract every style builder must satisfy.
- [GUIDE.md](GUIDE.md) — task-oriented decision guide and standard scales.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — package boundaries (WASM / non-WASM), anatomy and state contracts, and CSS emission guarantees.
- [docs/SPECS.md](docs/SPECS.md) — exact public surface, scale-to-token mappings, emission tables, and validation conditions.
- [docs/DESIGN.md](docs/DESIGN.md) — the reasoning behind each decision and the alternatives rejected.
- [docs/TRADEOFFS.md](docs/TRADEOFFS.md) — what the architecture buys, what it costs, and the limitations it accepts.
- [docs/MIGRATION.md](docs/MIGRATION.md) — moving a consumer to the closed API; includes the v0.5.0 mapping for the one typed state writer (`dom.BindState`).
- [docs/diagrams/BOUNDARIES.md](docs/diagrams/BOUNDARIES.md) — how `css`, `widget` and `ssr` divide the problem, and the WASM boundary.
