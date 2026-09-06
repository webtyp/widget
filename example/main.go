//go:build !wasm

package main

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

type MyWidget struct{}

func (w *MyWidget) WidgetName() widget.Name { return widget.Name("my-button") }
func (w *MyWidget) WidgetKind() widget.Kind { return widget.Region }

func main() {
	wd := &MyWidget{}
	sheet := style.For(wd).
		Root(style.As(style.Primary), style.Pad(style.Space4)).
		Part("label", style.FontSize(style.TextSm), style.FontWeight(style.WeightBold))

	cssOut := sheet.Stylesheet().String()
	fmt.Println(cssOut)
}
