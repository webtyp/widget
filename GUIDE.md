# User Guide — `webtyp/widget`

This guide helps you choose the correct layout, styling, and structural components using `webtyp/widget`.

## The Decision Table

The substitute for design judgement: you do not choose, you look up.

| I want… | Use |
|---|---|---|
| a column of things | `Stack(Space2)` |
| a row of buttons | `Row(Space1)` |
| a grid that adapts by itself | `Grid(ColumnNarrow, Space2)` |
| a grid with an exact column count that never reflows (a calendar week, a fixed strip) | `FixedGrid(7, Space2)` |
| list plus detail | `Split(SplitTwoThirds, Space3)` |
| a centred column of text | `Center(Readable)` |
| a horizontal scrolling strip | `ScrollRow(Space2)` |
| an image with a fixed proportion | `MediaBox(Aspect16x9)` |
| an svg icon that keeps its box | `IconBox(IconLg)` |
| the item in a row that pushes the rest aside | `Grow()` |
| an item pinned to the trailing edge of its line | `PushEnd()` |
| the frame of a whole application | `Cover()` |
| list and detail swiped on a phone | `On(css.Mobile, "", MasterDetail(Most))` |
| a fixed nav rail beside the content | `Sidebar(SideEnd, RailNarrow, SpaceNone)` |
| a button holding only an icon | `CenterContent()` |
| a chevron (or disclosure arrow) that flips when its host opens | `Rotate(TurnNone)` on the base rule + `When(widget.Open, "icon", Rotate(TurnHalf))`, with `Animate(MotionBase)` so the turn is a transition — the rotation IS the state, never a class toggled by hand |
| a fixed-size item (an `IconBox()`) centered inside a wider row/grid track | `CenterSelf()` |
| a runtime fill level — an occupancy bar, a progress meter | `Meter(Space1)`, host sets `--meter-fill:N%;` per instance |
| a floating action pinned to a corner | `Anchor()` on the panel + `Docked(Parent, EdgeBottom, SideEnd, Space4)` |
| an action that stays put while panels swipe | `Docked(Viewport, EdgeBottom, SideEnd, Space4)` |
| a full-height strip of chrome pinned along one whole edge, sized to its own content (not a corner, not a forced width) | `Anchor()` on the panel + `EdgeStrip(Parent, SideStart)` |
| a chip riding a box's border, legend-style | `Anchor()` on the box + `OnEdge(EdgeTop, SideStart, …)` — the straddle is half the shared `--chip-height` token |
| a scroll container that never hides what floats over its end | `FloatingChrome(EdgeBottom, IconLg, Space2)` on the floating element + `Scroll()` on the container |
| a dropdown that does not push the page | `Anchor()` on the trigger + `Flyout(SideEnd)` on the list |
| a dropdown whose trigger sits inside another positioned part | `Within("menu", "options", Flyout(SideStart))` — declare the nesting, or `Validate()` rejects the sheet: the positioned part between the Flyout and its Anchor steals the containing block |
| a dropdown inside a scrolling list | an accordion in flow, or move the panel out of the scroller — `Validate()` rejects a Flyout whose chain passes through a `Scroll()` region, because the scroller clips the panel |
| a panel that slides in from an edge | `Drawer(SideEnd, TwoThirds)` + `RevealedBy(widget.Open)` |
| a floating control pinned to the vertical middle of an edge | `FloatMiddle(SideEnd, Space4)` — corners stay `Docked`'s job |
| panels that slide instead of jumping | `SlideDeck(MotionBase)` — the panel on screen carries `widget.Current` |
| a container that reveals its children on hover | `CueWithin(Hover, "menu", "link-text", …)` |
| a part that exists everywhere but a phone | `On(css.Mobile, "header", Hide())` |
| a different arrangement on phones | `On(css.Mobile, "part", …)` |
| an element that exists only on phones | `OnlyOn(css.Mobile, "part", …)` |
| the page background | `As(Page)` |
| a card or panel | `As(Panel)` |
| something clickable | `Interactive(Primary)` |
| something clickable, secondary | `Interactive(Secondary)` |
| the selected item of a list | `When(widget.Selected, "item", As(Highlight))` |
| secondary text | `As(Subtle)` |
| an error | `As(Danger)` |
| to fill the remaining height | `Fill()` |
| to scroll internally | `Scroll()` |
| something that expands | `RevealedBy(widget.Open)` — plus `Animate(MotionBase)` on the same rule if the swap should fade in and out instead of cutting (see SPECS §5.2; instant without it) |
| a modal dialog | `Backdrop(Viewport)` + `Veil()` |

## Principles

1. **Deterministic CSS:** What you declare in Go is exactly what is generated in CSS.
2. **Zero-value compatibility:** Component stylesheet builders should not read fields; they should work on `&T{}`.
3. **No custom property leakage:** All styling relies on a closed catalog of CSS variables and design tokens.
