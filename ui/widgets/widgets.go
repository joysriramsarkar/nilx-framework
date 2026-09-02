// Package widgets provides built-in declarative UI component definitions and builder helpers for Alap.
package widgets

import (
	"github.com/joysriramsarkar/alap-framework/ui/engine"
)

// Column creates a vertical flex container.
func Column(children ...*engine.WidgetNode) *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Column")
	for _, child := range children {
		node.AddChild(child)
	}
	return node
}

// Row creates a horizontal flex container.
func Row(children ...*engine.WidgetNode) *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Row")
	for _, child := range children {
		node.AddChild(child)
	}
	return node
}

// Text creates a text label.
func Text(content string) *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Text")
	node.SetProp("text", content)
	node.SetProp("fontSize", 16.0)
	node.SetProp("color", "#000000")
	return node
}

// Button creates an interactive button widget.
func Button(label string, onClick engine.EventHandler) *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Button")
	node.SetProp("text", label)
	node.SetProp("backgroundColor", "#176BFF")
	node.SetProp("color", "#FFFFFF")
	node.SetProp("borderRadius", 8.0)
	if onClick != nil {
		node.SetEvent("onClick", onClick)
	}
	return node
}

// TextInput creates a text input field.
func TextInput(placeholder string, onChange engine.EventHandler) *engine.WidgetNode {
	node := engine.NewWidgetNode("", "TextInput")
	node.SetProp("placeholder", placeholder)
	node.SetProp("fontSize", 16.0)
	node.SetProp("border", "#D1D1D6")
	if onChange != nil {
		node.SetEvent("onChange", onChange)
	}
	return node
}

// Image creates an image view with a URL or local asset path.
func Image(source string, width, height float64) *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Image")
	node.SetProp("src", source)
	if width > 0 {
		node.SetProp("width", width)
	}
	if height > 0 {
		node.SetProp("height", height)
	}
	return node
}

// Card creates a container with elevation, padding, and rounded corners.
func Card(children ...*engine.WidgetNode) *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Card")
	node.SetProp("backgroundColor", "#FFFFFF")
	node.SetProp("borderRadius", 12.0)
	node.SetProp("padding", 16.0)
	for _, child := range children {
		node.AddChild(child)
	}
	return node
}

// Divider creates a horizontal separator line.
func Divider() *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Divider")
	node.SetProp("height", 1.0)
	node.SetProp("color", "#E5E5EA")
	return node
}

// Spacer creates flexible empty space inside Column or Row.
func Spacer() *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Spacer")
	node.SetProp("flex", 1.0)
	return node
}

// Switch creates a toggle switch.
func Switch(isOn bool, onToggle engine.EventHandler) *engine.WidgetNode {
	node := engine.NewWidgetNode("", "Switch")
	node.SetProp("value", isOn)
	if onToggle != nil {
		node.SetEvent("onChange", onToggle)
	}
	return node
}
