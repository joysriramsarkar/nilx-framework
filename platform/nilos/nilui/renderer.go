// Package nilui implements the native NilOS Vulkan / Wayland UI rendering bridge for NilX.
package nilui

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/joysriramsarkar/nilx-framework/ui/engine"
)

// DrawCommandKind specifies the primitive operation for nilui-gpu.
type DrawCommandKind int

const (
	DrawRect DrawCommandKind = iota
	DrawRoundedRect
	DrawText
	DrawImage
	DrawClipBegin
	DrawClipEnd
)

// DrawCommand represents a GPU rasterization primitive sent to nilui-gpu.
type DrawCommand struct {
	Kind         DrawCommandKind `json:"kind"`
	X            float64         `json:"x"`
	Y            float64         `json:"y"`
	Width        float64         `json:"width"`
	Height       float64         `json:"height"`
	Color        string          `json:"color,omitempty"`
	BorderColor  string          `json:"borderColor,omitempty"`
	BorderWidth  float64         `json:"borderWidth,omitempty"`
	BorderRadius float64         `json:"borderRadius,omitempty"`
	Text         string          `json:"text,omitempty"`
	FontSize     float64         `json:"fontSize,omitempty"`
	FontWeight   string          `json:"fontWeight,omitempty"`
}

// FramePacket encapsulates a complete serialized render pass for Vulkan present.
type FramePacket struct {
	Width    int           `json:"width"`
	Height   int           `json:"height"`
	Commands []DrawCommand `json:"commands"`
}

// Renderer manages the NilOS Wayland surface and Vulkan command pipeline.
type Renderer struct {
	mu           sync.RWMutex
	SurfaceTitle string
	Width        int
	Height       int
	ScaleFactor  float64
	frameCount   uint64
}

// NewRenderer creates a new NilUI GPU renderer.
func NewRenderer(title string, width, height int) *Renderer {
	return &Renderer{
		SurfaceTitle: title,
		Width:        width,
		Height:       height,
		ScaleFactor:  1.0,
	}
}

// RenderTree converts a NilX UI node hierarchy into GPU draw commands.
func (r *Renderer) RenderTree(root *engine.WidgetNode) *FramePacket {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frameCount++

	packet := &FramePacket{
		Width:    r.Width,
		Height:   r.Height,
		Commands: make([]DrawCommand, 0, 64),
	}

	if root == nil {
		return packet
	}

	r.traverseNode(root, packet)
	return packet
}

func (r *Renderer) traverseNode(node *engine.WidgetNode, packet *FramePacket) {
	if node == nil {
		return
	}

	// 1. Generate background draw command
	bgColor := "#FFFFFF"
	if bg, ok := node.Props["backgroundColor"].(string); ok && bg != "" {
		bgColor = bg
	}

	borderRadius := 0.0
	if rad, ok := node.Props["borderRadius"].(float64); ok {
		borderRadius = rad
	}

	if node.Type == "Card" || node.Type == "Button" || node.Props["backgroundColor"] != nil {
		packet.Commands = append(packet.Commands, DrawCommand{
			Kind:         DrawRoundedRect,
			X:            node.Bounds.X,
			Y:            node.Bounds.Y,
			Width:        node.Bounds.Width,
			Height:       node.Bounds.Height,
			Color:        bgColor,
			BorderRadius: borderRadius,
		})
	}

	// 2. Generate text draw command
	if node.Type == "Text" {
		text := ""
		if t, ok := node.Props["text"].(string); ok {
			text = t
		}
		textColor := "#000000"
		if c, ok := node.Props["color"].(string); ok && c != "" {
			textColor = c
		}
		fontSize := 14.0
		if fs, ok := node.Props["fontSize"].(float64); ok {
			fontSize = fs
		}
		fontWeight := "normal"
		if fw, ok := node.Props["fontWeight"].(string); ok {
			fontWeight = fw
		}

		packet.Commands = append(packet.Commands, DrawCommand{
			Kind:       DrawText,
			X:          node.Bounds.X,
			Y:          node.Bounds.Y,
			Width:      node.Bounds.Width,
			Height:     node.Bounds.Height,
			Text:       text,
			Color:      textColor,
			FontSize:   fontSize,
			FontWeight: fontWeight,
		})
	}

	// 3. Traverse child hierarchy
	for _, child := range node.Children {
		r.traverseNode(child, packet)
	}
}

// ExportFrameJSON produces the GPU JSON payload consumed by nilui-gpu Vulkan presenter.
func (r *Renderer) ExportFrameJSON(packet *FramePacket) (string, error) {
	data, err := json.Marshal(packet)
	if err != nil {
		return "", fmt.Errorf("failed serializing frame packet: %w", err)
	}
	return string(data), nil
}
