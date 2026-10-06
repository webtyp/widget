//go:build !wasm

package style

type flowType uint8

const (
	flowNone flowType = iota
	flowStack
	flowRow
	flowSplit
	flowGrid
	flowCenter
	flowFillCentered
	flowScrollRow
	flowMediaBox
	flowCover
	flowSidebar
	flowMasterDetail
	flowSlideDeck
	flowFixedGrid
	flowAutoRotate
)

// AutoRotateLayers is the fixed number of stacked children AutoRotate()
// choreographs. RenderCSS runs on a zero-value receiver (see the package's
// zero-value contract), so the rule can never see how many images a real
// instance holds — it has to commit to a layer count at compile time instead
// of computing one from instance data. A caller with fewer real images than
// AutoRotateLayers must cycle through them to fill every slot (Images[i %
// len(Images)] for i in [0, AutoRotateLayers)); leaving a slot's position
// empty in the DOM produces a silent gap in the rotation — no image visible
// — for that slot's turn, because the CSS has no way to know the slot is
// unused and shrink the cycle around it.
const AutoRotateLayers = 6

// Stack defines a vertical rhythm with children at full width.
func Stack(gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowStack
		r.flowGap = gap
	}
}

// Row defines a horizontal flow that wraps when it does not fit.
func Row(gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowRow
		r.flowGap = gap
	}
}

// Split defines two panels that stack below their own width.
func Split(ratio SplitRatio, gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowSplit
		r.flowRatio = ratio
		r.flowGap = gap
	}
}

// Grid defines auto-fit + minmax without a fixed number of columns.
func Grid(min ColumnWidth, gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowGrid
		r.flowWidth = min
		r.flowGap = gap
	}
}

// FixedGrid lays out children in exactly cols equal-width columns; unlike
// Grid()'s auto-fit/minmax, the column count never reflows on its own. Use
// Grid() when the item count varies and should reflow with the container's
// width; use FixedGrid() when the column count is a structural fact — a
// calendar's 7 weekdays, a fixed-size month strip — and every column must
// stay equal regardless of content.
//
// cols becomes the --cols custom property, not a literal repeat(N, 1fr): a
// stylesheet builder must work on a zero-value receiver (it cannot read
// instance fields), so a column count only known at runtime is set the same
// way any other per-instance value crosses into an otherwise-static
// stylesheet — the host overrides --cols inline on the element, never
// grid-template-columns itself.
func FixedGrid(cols int, gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowFixedGrid
		r.flowCols = cols
		r.flowGap = gap
	}
}

// Center defines a centered column with an optional maximum size (defaults to Readable).
func Center(max ...Size) Option {
	sz := Readable
	if len(max) > 0 {
		sz = max[0]
	}
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowCenter
		r.hasSize = true
		r.size = sz
	}
}

// FillCentered fills the container with a centered child.
func FillCentered() Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowFillCentered
	}
}

// ScrollRow defines a horizontal scrolling strip with scroll-snap and smooth
// scroll-behavior — so a same-page anchor link (<a href="#childID">) or a
// programmatic scroll into one of its children slides instead of jumping,
// with no JS and no per-consumer opt-in.
func ScrollRow(gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowScrollRow
		r.flowGap = gap
	}
}

// MediaBox defines a box of fixed aspect ratio.
func MediaBox(a Aspect) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowMediaBox
		r.flowAspect = a
	}
}

// Cover locks the frame to the viewport height and stacks its children
// vertically. It is the outermost frame of an application shell: use KeepSize()
// on the children that must not shrink (a header) and Fill() on the one that
// takes the remaining height. Do not nest one Cover inside another.
//
// The height is definite, not a floor, so a Fill() descendant resolves against
// it and a HideOverflow() or Scroll() descendant actually clips. A shell whose
// content can exceed the viewport must therefore give that descendant Scroll();
// otherwise the overflow is unreachable.
func Cover() Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowCover
	}
}

// SlideDeck apila a sus hijos en capas que ocupan el contenedor entero y muestra
// solo aquel que lleva el estado widget.Current; los demás quedan aparcados en el
// borde inline-start y entran deslizándose de izquierda a derecha cuando les toca.
//
// Es la forma de cambiar de panel en un shell SIN crear un scroller: un contenedor
// de scroll-snap horizontal aquí encadena con el scroll-snap horizontal que un
// módulo pueda tener adentro, y el gesto de deslizar dentro del contenido termina
// cambiando de sección sola.
//
// Todos los hijos siguen montados en el DOM. Ese es el trato: el estado decide
// cuál está en pantalla, nadie desmonta nada. El contenedor es el bloque contenedor
// de sus hijos, así que un Docked(Parent) dentro de un panel se resuelve contra SU
// panel — no hace falta Anchor() en el hijo, y ponerlo lo ROMPE: el position:
// relative de @layer widgets gana sobre el position: absolute que este flujo emite
// en @layer primitives.
//
// m gobierna la duración del deslizamiento. MotionNone conmuta sin animación.
func SlideDeck(m Motion) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowSlideDeck
		r.flowMotion = m
	}
}

// AutoRotate stacks up to AutoRotateLayers children as full-bleed layers and
// cross-fades between them forever, driven purely by a shared @keyframes
// rule and a per-child animation-delay staggered by DOM position — no state
// to manage, no JS, no scroller. It is the unattended counterpart of
// SlideDeck: SlideDeck changes panel because something set widget.Current,
// AutoRotate changes layer because time passed.
//
// The stagger is expressed as :nth-child selectors, not as a parameter: like
// every other rule in this package, AutoRotate runs on a zero-value receiver
// when the stylesheet is built (see the package doc), so it cannot read how
// many children a real instance renders. It always choreographs exactly
// AutoRotateLayers slots; a caller with fewer real images must tile them
// across all slots (see AutoRotateLayers) so no slot's turn goes dark.
//
// prefers-reduced-motion turns the animation off. Every layer then rests at
// its own plain (non-animated) opacity, which this rule sets to visible only
// for :first-child — the reduced-motion fallback is "first image, frozen",
// automatically, with no extra markup from the caller.
func AutoRotate() Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowAutoRotate
	}
}

// MasterDetail turns a two-panel container into a horizontal scroll-snap strip
// for a narrow screen: the master list rests where the browser's default scroll
// position already is, and the detail sits beside it at `detail` of the strip's
// width, so a sliver of the list stays visible and the panel it came from is
// obvious. Swiping is a native scroll; snapping to the detail is a plain
// ScrollIntoView from the row handler.
//
// The FIRST TWO element children are the panels, in the same DOM order a desktop
// Split uses: detail first, master second. Anything after them — a modal mount
// point, a portal anchor — is left alone, which is why this addresses them by
// position and not with :first-child/:last-child. The strip is laid out RTL so
// the master — the second child, given order 1 — lands at the start edge, which
// RTL puts on the right, exactly where scroll position 0 already rests. That is
// what removes the need for a scroll nudge at mount time, which this framework's
// component contract has no hook for. Each panel resets to LTR so only the
// outer strip's flow is mirrored, never the content.
func MasterDetail(detail Size) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowMasterDetail
		r.flowDetail = detail
	}
}

// Sidebar places a fixed-width rail beside a fluid content area. The rail keeps
// its width; the content takes everything else. Below the point where the content
// can no longer hold its minimum width the two reflow into a single column, with
// no media query involved.
//
// The container MUST have exactly two element children. Which one is the rail is
// decided by side, not by DOM order: SideEnd makes the LAST child the rail.
func Sidebar(side Side, width RailWidth, gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowSidebar
		r.flowSide = side
		r.flowRail = width
		r.flowGap = gap
	}
}

// SpanFull makes an element span across all columns of a grid container.
func SpanFull() Option {
	return func(r *rule) {
		r.spanFull = true
	}
}

