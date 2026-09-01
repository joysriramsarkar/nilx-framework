// Package layout implements full Flexbox layout calculation for NilX components.
package layout

import (
	"strings"

	"github.com/joysriramsarkar/nilx-framework/ui/engine"
)

// LayoutContext holds screen size and rendering constraints.
type LayoutContext struct {
	ViewportWidth  float64
	ViewportHeight float64
	Scale          float64
}

// Insets represents 4-sided margins or padding.
type Insets struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// FlexDirection specifies main axis orientation.
type FlexDirection int

const (
	FlexDirectionColumn FlexDirection = iota
	FlexDirectionRow
	FlexDirectionColumnReverse
	FlexDirectionRowReverse
)

// JustifyContent specifies alignment along the main axis.
type JustifyContent int

const (
	JustifyFlexStart JustifyContent = iota
	JustifyCenter
	JustifyFlexEnd
	JustifySpaceBetween
	JustifySpaceAround
	JustifySpaceEvenly
)

// AlignItems specifies alignment along the cross axis.
type AlignItems int

const (
	AlignFlexStart AlignItems = iota
	AlignCenter
	AlignFlexEnd
	AlignStretch
)

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
	LayoutNode(root, root.Bounds)
}

// LayoutNode computes layout bounds recursively for any node.
func LayoutNode(node *engine.WidgetNode, available engine.Rect) {
	if node == nil {
		return
	}

	// Apply margin to available space
	margin := parseInsets(node.Props["margin"])
	paddedAvailable := engine.Rect{
		X:      available.X + margin.Left,
		Y:      available.Y + margin.Top,
		Width:  available.Width - (margin.Left + margin.Right),
		Height: available.Height - (margin.Top + margin.Bottom),
	}
	if paddedAvailable.Width < 0 {
		paddedAvailable.Width = 0
	}
	if paddedAvailable.Height < 0 {
		paddedAvailable.Height = 0
	}

	switch node.Type {
	case "Column":
		layoutFlex(node, paddedAvailable, FlexDirectionColumn)
	case "Row":
		layoutFlex(node, paddedAvailable, FlexDirectionRow)
	default:
		layoutLeafOrContainer(node, paddedAvailable)
	}
}

func layoutFlex(node *engine.WidgetNode, available engine.Rect, dir FlexDirection) {
	padding := parseInsets(node.Props["padding"])
	contentRect := engine.Rect{
		X:      available.X + padding.Left,
		Y:      available.Y + padding.Top,
		Width:  available.Width - (padding.Left + padding.Right),
		Height: available.Height - (padding.Top + padding.Bottom),
	}
	if contentRect.Width < 0 {
		contentRect.Width = 0
	}
	if contentRect.Height < 0 {
		contentRect.Height = 0
	}

	justify := parseJustifyContent(node.Props["justifyContent"])
	align := parseAlignItems(node.Props["alignItems"])
	spacing := node.GetFloatProp("spacing", 0)

	children := node.Children
	numChildren := len(children)
	if numChildren == 0 {
		return
	}

	// Step 1: Measure non-flex children and calculate total flex-grow
	var totalGrow float64
	var fixedMainAxis float64
	measuredSizes := make([]engine.Rect, numChildren)

	for i, child := range children {
		grow := child.GetFloatProp("flexGrow", 0)
		if child.Type == "Spacer" && grow == 0 {
			grow = 1.0 // Spacers default to flex-grow 1
		}

		if grow > 0 {
			totalGrow += grow
		} else {
			mWidth := measureChildWidth(child, contentRect.Width)
			mHeight := measureChildHeight(child, contentRect.Height)
			measuredSizes[i] = engine.Rect{Width: mWidth, Height: mHeight}
			if dir == FlexDirectionColumn {
				fixedMainAxis += mHeight
			} else {
				fixedMainAxis += mWidth
			}
		}
	}

	totalSpacing := spacing * float64(numChildren-1)
	if totalSpacing < 0 {
		totalSpacing = 0
	}

	// Step 2: Distribute remaining main axis space among flex-grow items
	var remainingMainAxis float64
	if dir == FlexDirectionColumn {
		remainingMainAxis = contentRect.Height - (fixedMainAxis + totalSpacing)
	} else {
		remainingMainAxis = contentRect.Width - (fixedMainAxis + totalSpacing)
	}
	if remainingMainAxis < 0 {
		remainingMainAxis = 0
	}

	for i, child := range children {
		grow := child.GetFloatProp("flexGrow", 0)
		if child.Type == "Spacer" && grow == 0 {
			grow = 1.0
		}
		if grow > 0 && totalGrow > 0 {
			allocated := (grow / totalGrow) * remainingMainAxis
			if dir == FlexDirectionColumn {
				measuredSizes[i] = engine.Rect{
					Width:  measureChildWidth(child, contentRect.Width),
					Height: allocated,
				}
			} else {
				measuredSizes[i] = engine.Rect{
					Width:  allocated,
					Height: measureChildHeight(child, contentRect.Height),
				}
			}
		}
	}

	// Step 3: Arrange children according to justifyContent & alignItems
	var curX = contentRect.X
	var curY = contentRect.Y

	// Adjust starting position or spacing for justifyContent
	var gap = spacing
	if totalGrow == 0 && remainingMainAxis > 0 {
		switch justify {
		case JustifyCenter:
			if dir == FlexDirectionColumn {
				curY += remainingMainAxis / 2
			} else {
				curX += remainingMainAxis / 2
			}
		case JustifyFlexEnd:
			if dir == FlexDirectionColumn {
				curY += remainingMainAxis
			} else {
				curX += remainingMainAxis
			}
		case JustifySpaceBetween:
			if numChildren > 1 {
				gap = spacing + (remainingMainAxis / float64(numChildren-1))
			}
		case JustifySpaceAround:
			if numChildren > 0 {
				aroundGap := remainingMainAxis / float64(numChildren)
				gap = spacing + aroundGap
				if dir == FlexDirectionColumn {
					curY += aroundGap / 2
				} else {
					curX += aroundGap / 2
				}
			}
		}
	}

	for i, child := range children {
		cRect := measuredSizes[i]
		var itemX = curX
		var itemY = curY

		// Cross axis alignment
		if dir == FlexDirectionColumn {
			switch align {
			case AlignCenter:
				itemX = contentRect.X + (contentRect.Width-cRect.Width)/2
			case AlignFlexEnd:
				itemX = contentRect.X + (contentRect.Width - cRect.Width)
			case AlignStretch:
				cRect.Width = contentRect.Width
			}
		} else {
			switch align {
			case AlignCenter:
				itemY = contentRect.Y + (contentRect.Height-cRect.Height)/2
			case AlignFlexEnd:
				itemY = contentRect.Y + (contentRect.Height - cRect.Height)
			case AlignStretch:
				cRect.Height = contentRect.Height
			}
		}

		child.Bounds = engine.Rect{
			X:      itemX,
			Y:      itemY,
			Width:  cRect.Width,
			Height: cRect.Height,
		}

		LayoutNode(child, child.Bounds)

		if dir == FlexDirectionColumn {
			curY += cRect.Height + gap
		} else {
			curX += cRect.Width + gap
		}
	}
}

func layoutLeafOrContainer(node *engine.WidgetNode, available engine.Rect) {
	node.Bounds = available
	for _, child := range node.Children {
		LayoutNode(child, available)
	}
}

func measureChildWidth(node *engine.WidgetNode, availableWidth float64) float64 {
	if w := node.GetFloatProp("width", 0); w > 0 {
		return w
	}
	if node.Type == "Text" {
		text := node.GetStringProp("text", "")
		fontSize := node.GetFloatProp("fontSize", 14)
		// Approximate character width (~0.6 * fontSize)
		w := float64(len(text)) * (fontSize * 0.6)
		if w > availableWidth && availableWidth > 0 {
			return availableWidth
		}
		if w < 20 {
			return 20
		}
		return w
	}
	return availableWidth
}

func measureChildHeight(node *engine.WidgetNode, availableHeight float64) float64 {
	if h := node.GetFloatProp("height", 0); h > 0 {
		return h
	}
	if node.Type == "Text" {
		fontSize := node.GetFloatProp("fontSize", 14)
		return fontSize * 1.4
	}
	if node.Type == "Button" {
		return 40
	}
	if node.Type == "TextInput" {
		return 44
	}
	if node.Type == "Divider" {
		return 1
	}
	if node.Type == "Spacer" {
		return 0
	}
	return availableHeight
}

func parseInsets(v interface{}) Insets {
	if v == nil {
		return Insets{}
	}
	switch val := v.(type) {
	case float64:
		return Insets{Top: val, Right: val, Bottom: val, Left: val}
	case int:
		f := float64(val)
		return Insets{Top: f, Right: f, Bottom: f, Left: f}
	case map[string]interface{}:
		var ins Insets
		if t, ok := val["top"].(float64); ok {
			ins.Top = t
		}
		if r, ok := val["right"].(float64); ok {
			ins.Right = r
		}
		if b, ok := val["bottom"].(float64); ok {
			ins.Bottom = b
		}
		if l, ok := val["left"].(float64); ok {
			ins.Left = l
		}
		if h, ok := val["horizontal"].(float64); ok {
			ins.Left = h
			ins.Right = h
		}
		if v, ok := val["vertical"].(float64); ok {
			ins.Top = v
			ins.Bottom = v
		}
		return ins
	}
	return Insets{}
}

func parseJustifyContent(v interface{}) JustifyContent {
	if s, ok := v.(string); ok {
		switch strings.ToLower(s) {
		case "center":
			return JustifyCenter
		case "flex-end":
			return JustifyFlexEnd
		case "space-between":
			return JustifySpaceBetween
		case "space-around":
			return JustifySpaceAround
		case "space-evenly":
			return JustifySpaceEvenly
		}
	}
	return JustifyFlexStart
}

func parseAlignItems(v interface{}) AlignItems {
	if s, ok := v.(string); ok {
		switch strings.ToLower(s) {
		case "center":
			return AlignCenter
		case "flex-end":
			return AlignFlexEnd
		case "stretch":
			return AlignStretch
		}
	}
	return AlignFlexStart
}
