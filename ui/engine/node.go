// Package engine provides the core UI tree model, reactive state binding, and rendering abstractions for NilX.
package engine

import (
	"fmt"
	"strings"
)

// Rect represents element bounding box and dimensions.
type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Spacing represents padding or margin inside/outside elements.
type Spacing struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// EventHandler is a callback invoked when a UI event fires.
type EventHandler func(args ...interface{})

// WidgetNode represents a single declarative UI element in the NilX UI hierarchy.
type WidgetNode struct {
	ID       string                  `json:"id"`
	Type     string                  `json:"type"` // e.g. "Column", "Row", "Text", "Button"
	Props    map[string]interface{}  `json:"props"`
	Children []*WidgetNode           `json:"children"`
	Bounds   Rect                    `json:"bounds"`
	Events   map[string]EventHandler `json:"-"`
	Parent   *WidgetNode             `json:"-"`
}

// NewWidgetNode creates a new unattached UI node.
func NewWidgetNode(id, widgetType string) *WidgetNode {
	return &WidgetNode{
		ID:       id,
		Type:     widgetType,
		Props:    make(map[string]interface{}),
		Children: make([]*WidgetNode, 0),
		Events:   make(map[string]EventHandler),
	}
}

// SetProp assigns a property (like color, fontSize, margin, padding, text).
func (w *WidgetNode) SetProp(key string, value interface{}) {
	w.Props[key] = value
}

// GetProp retrieves a property value or returns default.
func (w *WidgetNode) GetProp(key string, fallback interface{}) interface{} {
	if val, ok := w.Props[key]; ok {
		return val
	}
	return fallback
}

// GetStringProp retrieves a string property.
func (w *WidgetNode) GetStringProp(key, fallback string) string {
	if val, ok := w.Props[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", val)
	}
	return fallback
}

// GetFloatProp retrieves a numeric property as float64.
func (w *WidgetNode) GetFloatProp(key string, fallback float64) float64 {
	if val, ok := w.Props[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		}
	}
	return fallback
}

// AddChild appends a child node.
func (w *WidgetNode) AddChild(child *WidgetNode) {
	if child == nil {
		return
	}
	child.Parent = w
	w.Children = append(w.Children, child)
}

// SetEvent registers an event handler (e.g. "onClick", "onChange").
func (w *WidgetNode) SetEvent(name string, handler EventHandler) {
	w.Events[name] = handler
}

// TriggerEvent invokes a registered event handler.
func (w *WidgetNode) TriggerEvent(name string, args ...interface{}) bool {
	if handler, ok := w.Events[name]; ok && handler != nil {
		handler(args...)
		return true
	}
	return false
}

// DumpString generates a human-readable indented view of the UI node.
func (w *WidgetNode) DumpString(indent int) string {
	var sb strings.Builder
	pad := strings.Repeat("  ", indent)
	sb.WriteString(fmt.Sprintf("%s<%s", pad, w.Type))

	for k, v := range w.Props {
		sb.WriteString(fmt.Sprintf(" %s=%q", k, fmt.Sprintf("%v", v)))
	}

	if len(w.Children) == 0 {
		sb.WriteString(" />\n")
		return sb.String()
	}

	sb.WriteString(">\n")
	for _, child := range w.Children {
		sb.WriteString(child.DumpString(indent + 1))
	}
	sb.WriteString(fmt.Sprintf("%s</%s>\n", pad, w.Type))
	return sb.String()
}
