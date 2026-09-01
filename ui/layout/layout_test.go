package layout

import (
	"testing"

	"github.com/joysriramsarkar/nilx-framework/ui/engine"
)

func TestFlexboxColumnLayout(t *testing.T) {
	root := engine.NewWidgetNode("col_1", "Column")
	root.Props["spacing"] = 10.0
	root.Props["padding"] = 16.0

	t1 := engine.NewWidgetNode("text_1", "Text")
	t1.Props["text"] = "Header"
	t1.Props["fontSize"] = 20.0
	root.AddChild(t1)

	spacer := engine.NewWidgetNode("sp_1", "Spacer")
	root.AddChild(spacer)

	btn := engine.NewWidgetNode("btn_1", "Button")
	btn.Props["height"] = 50.0
	root.AddChild(btn)

	ctx := LayoutContext{ViewportWidth: 400, ViewportHeight: 800}
	ComputeLayout(root, ctx)

	if root.Bounds.Width != 400 || root.Bounds.Height != 800 {
		t.Errorf("expected root bounds 400x800, got %fx%f", root.Bounds.Width, root.Bounds.Height)
	}

	// Button should be at bottom
	if btn.Bounds.Y < 700 {
		t.Errorf("expected button pushed to bottom by spacer, got Y=%f", btn.Bounds.Y)
	}
}

func TestFlexboxRowJustifyCenter(t *testing.T) {
	root := engine.NewWidgetNode("row_1", "Row")
	root.Props["justifyContent"] = "center"
	root.Props["width"] = 300.0
	root.Props["height"] = 100.0

	b1 := engine.NewWidgetNode("b1", "Button")
	b1.Props["width"] = 100.0
	b1.Props["height"] = 40.0
	root.AddChild(b1)

	ctx := LayoutContext{ViewportWidth: 300, ViewportHeight: 100}
	ComputeLayout(root, ctx)

	// Since row width=300 and child=100 with justifyCenter, child X should be (300-100)/2 = 100
	if b1.Bounds.X != 100 {
		t.Errorf("expected centered button X=100, got X=%f", b1.Bounds.X)
	}
}
