//go:build !wasm

package style_test

import (
	"fmt"

	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

type myButton struct{}

func (b *myButton) WidgetName() widget.Name { return widget.Name("btn") }
func (b *myButton) WidgetKind() widget.Kind { return widget.Region }

func ExampleFor() {
	btn := &myButton{}
	sheet := style.For(btn).
		Root(style.As(style.Primary))

	fmt.Println(sheet.Parts())
	// Output: []
}

func ExampleStack() {
	btn := &myButton{}
	_ = style.For(btn).
		Root(style.Stack(style.Space4))
}

func ExampleSplit() {
	btn := &myButton{}
	_ = style.For(btn).
		Root(style.Split(style.SplitTwoThirds, style.Space3))
}

func ExampleInteractive() {
	btn := &myButton{}
	_ = style.For(btn).
		Root(style.Interactive(style.Primary))
}

func ExampleRevealedBy() {
	btn := &myButton{}
	_ = style.For(btn).
		Part("menu", style.RevealedBy(widget.Open))
}

func ExampleBackdrop() {
	btn := &myButton{}
	_ = style.For(btn).
		Part("overlay", style.Backdrop(style.Viewport), style.Veil())
}
