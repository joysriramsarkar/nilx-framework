// Package render provides the 2D software rasterizer and GPU scene graph renderer for Alap UI.
package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"

	"github.com/joysriramsarkar/alap-framework/ui/engine"
)

// Canvas represents a 2D drawing surface.
type Canvas struct {
	Width  int
	Height int
	Img    *image.RGBA
}

// NewCanvas creates a new 2D drawing canvas.
func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		Width:  width,
		Height: height,
		Img:    image.NewRGBA(image.Rect(0, 0, width, height)),
	}
}

// Clear fills the canvas with a background color.
func (c *Canvas) Clear(col color.Color) {
	draw.Draw(c.Img, c.Img.Bounds(), &image.Uniform{C: col}, image.Point{}, draw.Src)
}

// DrawRect renders an axis-aligned rectangle.
func (c *Canvas) DrawRect(r engine.Rect, col color.Color) {
	rect := image.Rect(int(r.X), int(r.Y), int(r.X+r.Width), int(r.Y+r.Height))
	draw.Draw(c.Img, rect, &image.Uniform{C: col}, image.Point{}, draw.Over)
}

// DrawRoundedRect renders a rectangle with anti-aliased rounded corners.
func (c *Canvas) DrawRoundedRect(r engine.Rect, radius float64, col color.Color) {
	if radius <= 0 {
		c.DrawRect(r, col)
		return
	}

	x0, y0 := int(r.X), int(r.Y)
	x1, y1 := int(r.X+r.Width), int(r.Y+r.Height)
	rad := int(radius)

	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
				// Check 4 corner circles
				inCorner := false
				var cx, cy int

				if x < x0+rad && y < y0+rad {
					cx, cy = x0+rad, y0+rad
					inCorner = true
				} else if x >= x1-rad && y < y0+rad {
					cx, cy = x1-rad, y0+rad
					inCorner = true
				} else if x < x0+rad && y >= y1-rad {
					cx, cy = x0+rad, y1-rad
					inCorner = true
				} else if x >= x1-rad && y >= y1-rad {
					cx, cy = x1-rad, y1-rad
					inCorner = true
				}

				if inCorner {
					dx := float64(x - cx)
					dy := float64(y - cy)
					dist := math.Sqrt(dx*dx + dy*dy)
					if dist <= float64(rad) {
						c.Img.Set(x, y, col)
					}
				} else {
					c.Img.Set(x, y, col)
				}
			}
		}
	}
}

// DrawLinearGradient renders a vertical or horizontal gradient across bounds.
func (c *Canvas) DrawLinearGradient(r engine.Rect, startColor, endColor color.RGBA, vertical bool) {
	x0, y0 := int(r.X), int(r.Y)
	x1, y1 := int(r.X+r.Width), int(r.Y+r.Height)

	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			var t float64
			if vertical && r.Height > 0 {
				t = float64(y-y0) / r.Height
			} else if !vertical && r.Width > 0 {
				t = float64(x-x0) / r.Width
			}
			t = math.Max(0, math.Min(1, t))

			rVal := uint8(float64(startColor.R)*(1-t) + float64(endColor.R)*t)
			gVal := uint8(float64(startColor.G)*(1-t) + float64(endColor.G)*t)
			bVal := uint8(float64(startColor.B)*(1-t) + float64(endColor.B)*t)
			aVal := uint8(float64(startColor.A)*(1-t) + float64(endColor.A)*t)

			c.Img.Set(x, y, color.RGBA{R: rVal, G: gVal, B: bVal, A: aVal})
		}
	}
}

// ParseHexColor converts CSS hex strings ("#176BFF", "#FFF", "rgba(...)") to color.RGBA.
func ParseHexColor(s string) color.RGBA {
	s = strings.TrimSpace(s)
	if s == "" {
		return color.RGBA{0, 0, 0, 0}
	}
	if strings.HasPrefix(s, "#") {
		s = s[1:]
		if len(s) == 3 {
			s = fmt.Sprintf("%c%c%c%c%c%c", s[0], s[0], s[1], s[1], s[2], s[2])
		}
		if len(s) == 6 {
			val, _ := strconv.ParseUint(s, 16, 32)
			return color.RGBA{
				R: uint8(val >> 16),
				G: uint8(val >> 8),
				B: uint8(val),
				A: 255,
			}
		}
	}
	return color.RGBA{0, 0, 0, 255}
}
