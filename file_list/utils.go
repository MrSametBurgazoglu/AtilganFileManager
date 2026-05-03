package file_list

import (
	"math"

	"github.com/diamondburned/gotk4/pkg/cairo"
)

func roundedRectangle(cr *cairo.Context, x, y, w, h, r float64) {
	if w < 2*r {
		r = w / 2
	}
	if h < 2*r {
		r = h / 2
	}
	cr.MoveTo(x+r, y)
	cr.LineTo(x+w-r, y)
	cr.Arc(x+w-r, y+r, r, -math.Pi/2, 0)
	cr.LineTo(x+w, y+h-r)
	cr.Arc(x+w-r, y+h-r, r, 0, math.Pi/2)
	cr.LineTo(x+r, y+h)
	cr.Arc(x+r, y+h-r, r, math.Pi/2, math.Pi)
	cr.LineTo(x, y+r)
	cr.Arc(x+r, y+r, r, math.Pi, 3*math.Pi/2)
	cr.ClosePath()
}
