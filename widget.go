package widget

import "webtyp.com/fmt"

// Name identifies a widget. It is the prefix of EVERY class it emits,
// so two widgets cannot collide even if they choose the same part name.
type Name string

// Part is a named slot of a widget's anatomy (Open UI vocabulary).
// It is local to the widget: e.g. "row", "menu", "header". It never carries a prefix.
type Part string

// Class is a CSS class name. It has NO public constructor: the only way to obtain
// one is to derive it from a Name. This ensures markup and stylesheet agree by construction.
type Class string

func (c Class) String() string {
	return string(c)
}

func (c Class) AsAttr() fmt.KeyValue {
	return fmt.KeyValue{Key: "class", Value: string(c)}
}

// Root is the outer class of the widget.
func (n Name) Root() Class {
	return Class(n)
}

// Class derives the class of an anatomical part: e.g. "targetlist__row".
func (n Name) Class(p Part) Class {
	return Class(string(n) + "__" + string(p))
}
