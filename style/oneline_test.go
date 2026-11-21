//go:build !wasm

package style

import (
	"strings"
	"testing"

	"webtyp.com/widget"
)

type oneLineWidget struct{}

func (oneLineWidget) WidgetName() widget.Name { return "oneline" }
func (oneLineWidget) WidgetKind() widget.Kind { return widget.Region }

// A header with an icon and a text column stays one line: the row does not wrap.
func TestOneLine_RowDoesNotWrap(t *testing.T) {
	css := For(oneLineWidget{}).Part("header", Row(SpaceNone), OneLine()).Stylesheet().String()
	if !strings.Contains(css, "flex-wrap: nowrap;") {
		t.Fatalf("OneLine must emit flex-wrap: nowrap, got:\n%s", css)
	}
}
