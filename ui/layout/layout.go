// Package layout implements flexbox-style UI layout calculation for NilX components.
package layout

import (
	"github.com/joysriramsarkar/nilx-framework/ui/engine"
)

// LayoutContext holds screen size and rendering constraints.
type LayoutContext struct {
	ViewportWidth  float64
	ViewportHeight float64
	Scale          float64
}

// ComputeLayout calculates bounds (X, Y, Width, Height) for the widget tree.
func ComputeLayout(root *engine.WidgetNode, ctx LayoutContext) {
	if root == nil {
		return
	}
	root.Bounds = engine.Rect{
		X:      0,
		Y:      0,
		Width:  ctx.ViewportWidth,
		Height: ctx.ViewportHeight,
	}
	layoutNode(root, root.Bounds)
}

func layoutNode(node *engine.WidgetNode, available engine.Rect) {
	switch node.Type {
	case "Column":
		layoutColumn(node, available)
	case "Row":
		layoutRow(node, available)
	default:
		layoutLeafOrContainer(node, available)
	}
}

func layoutColumn(node *engine.WidgetNode, available engine.Rect) {
	curY := available.Y
	width := available.Width
	if wProp := node.GetFloatProp("width", 0); wProp > 0 {
		width = wProp
	}

	spacing := node.GetFloatProp("spacing", 0)

	for _, child := range node.Children {
		childH := child.GetFloatProp("height", 40) // default 40px
		if child.Type == "Text" {
			childH = child.GetFloatProp("fontSize", 16) * 1.5
		}

		childBounds := engine.Rect{
			X:      available.X,
			Y:      curY,
			Width:  width,
			Height: childH,
		}
		child.Bounds = childBounds
		layoutNode(child, childBounds)

		curY += childH + spacing
	}
}

func layoutRow(node *engine.WidgetNode, available engine.Rect) {
	curX := available.X
	height := available.Height
	if hProp := node.GetFloatProp("height", 0); hProp > 0 {
		height = hProp
	}

	spacing := node.GetFloatProp("spacing", 0)
	childCount := float64(len(node.Children))
	defaultW := 80.0
	if childCount > 0 {
		totalSpacing := spacing * (childCount - 1)
		defaultW = (available.Width - totalSpacing) / childCount
		if defaultW < 40 {
			defaultW = 40
		}
	}

	for _, child := range node.Children {
		childW := child.GetFloatProp("width", defaultW)
		childH := child.GetFloatProp("height", height)

		childBounds := engine.Rect{
			X:      curX,
			Y:      available.Y,
			Width:  childW,
			Height: childH,
		}
		child.Bounds = childBounds
		layoutNode(child, childBounds)

		curX += childW + spacing
	}
}

func layoutLeafOrContainer(node *engine.WidgetNode, available engine.Rect) {
	curY := available.Y
	for _, child := range node.Children {
		childH := child.GetFloatProp("height", 30)
		childBounds := engine.Rect{
			X:      available.X,
			Y:      curY,
			Width:  available.Width,
			Height: childH,
		}
		child.Bounds = childBounds
		layoutNode(child, childBounds)
		curY += childH
	}
}
