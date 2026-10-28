//go:build !wasm

package style

import (
	"strings"

	"webtyp.com/css"
)

// Paint colours an SVG shape the way As() colours a box: fill = the surface's
// background, stroke = its border (when the role has one). An SVG shape ignores
// background-color and border, so As() on a <rect> paints nothing.
//
//	Part("box", style.Paint(style.Panel))
func Paint(s Surface) Option {
	return func(r *rule) {
		r.hasPaint = true
		r.paint = s
	}
}

// Ink fills SVG text with the surface's text colour: SVG text is filled, not
// coloured, and defaults to black whatever the theme.
//
//	Part("name", style.Ink(style.Panel))
func Ink(s Surface) Option {
	return func(r *rule) {
		r.hasInk = true
		r.ink = s
	}
}

// Line draws an SVG line or path in the surface's colour with nothing filled
// (a filled curve paints the area under it). The colour is the role's border,
// or for a solid role with no border (Danger, Primary) the role's own colour.
//
//	Part("edge-error", style.Line(style.Danger))
func Line(s Surface) Option {
	return func(r *rule) {
		r.hasLine = true
		r.line = s
	}
}

// Dashed dashes the stroke that Paint or Line draws: a link that is not a
// constraint, a thing that is outside the drawing.
//
//	Part("edge-logical", style.Line(style.Panel), style.Dashed())
func Dashed() Option {
	return func(r *rule) {
		r.dashed = true
	}
}

// paintDecls emits the SVG paints, static light value first and enhanced var
// second, the same fallback surfaceDecls uses.
func (r rule) paintDecls() []string {
	var decls []string
	pair := func(prop, static, enhanced string) {
		if enhanced == "" {
			return
		}
		if static != "" {
			decls = append(decls, prop+": "+static+";")
		}
		decls = append(decls, prop+": "+enhanced+";")
	}
	if r.hasPaint {
		t := r.paint.resolve()
		pair("fill", t.bgStatic, t.bg)
		pair("stroke", borderColour(t.borderStatic), borderColour(t.border))
	}
	if r.hasInk {
		t := r.ink.resolve()
		pair("fill", t.textStatic, t.text)
		decls = append(decls, "stroke: none;")
	}
	if r.hasLine {
		t := r.line.resolve()
		decls = append(decls, "fill: none;")
		if t.border != "" {
			pair("stroke", borderColour(t.borderStatic), borderColour(t.border))
		} else {
			pair("stroke", t.bgStatic, t.bg)
		}
	}
	if r.dashed {
		decls = append(decls, "stroke-dasharray: "+css.Space1.Var()+" "+css.Space1.Var()+";")
	}
	return decls
}

// borderColour is the colour half of a surface border ("1px solid <colour>").
func borderColour(border string) string {
	return strings.TrimPrefix(border, borderStyle)
}
