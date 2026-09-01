package render

import (
	"image/color"
	"testing"

	"github.com/joysriramsarkar/nilx-framework/ui/engine"
)

func TestCanvasRasterizer(t *testing.T) {
	canvas := NewCanvas(100, 100)
	canvas.Clear(color.RGBA{255, 255, 255, 255})

	col := ParseHexColor("#176BFF")
	if col.R != 0x17 || col.G != 0x6B || col.B != 0xFF {
		t.Errorf("expected #176BFF parsed as (23, 107, 255), got (%d, %d, %d)", col.R, col.G, col.B)
	}

	canvas.DrawRoundedRect(engine.Rect{X: 10, Y: 10, Width: 80, Height: 80}, 8.0, col)

	// Center pixel (50, 50) should be colored #176BFF
	p := canvas.Img.RGBAAt(50, 50)
	if p.R != col.R || p.G != col.G || p.B != col.B {
		t.Errorf("expected pixel (50, 50) to be %v, got %v", col, p)
	}
}
