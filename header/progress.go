package header

import (
	"fmt"
	"math"

	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CircularProgressBar struct {
	*gtk.DrawingArea
	fraction float64
}

func NewCircularProgressBar() *CircularProgressBar {
	p := &CircularProgressBar{
		DrawingArea: gtk.NewDrawingArea(),
		fraction:    0.0,
	}

	p.SetDrawFunc(p.draw)
	return p
}

func (p *CircularProgressBar) SetFraction(fraction float64) {
	p.fraction = math.Max(0.0, math.Min(1.0, fraction))
	p.QueueDraw()
}

func (p *CircularProgressBar) draw(area *gtk.DrawingArea, cr *cairo.Context, width, height int) {
	cr.SetSourceRGBA(0, 0, 0, 0)
	cr.Paint()

	w := float64(width)
	h := float64(height)
	radius := math.Min(w, h) / 2.0 * 0.8
	cr.SetLineWidth(4.0)
	cr.SetSourceRGBA(0.5, 0.5, 0.5, 0.5)
	cr.Arc(w/2, h/2, radius, 0, 2*math.Pi)
	cr.Stroke()

	cr.SetSourceRGBA(0.1, 0.6, 0.9, 1.0)
	cr.Arc(w/2, h/2, radius, -math.Pi/2, 2*math.Pi*p.fraction-math.Pi/2)
	cr.Stroke()

	percentage := int(p.fraction * 100)
	text := fmt.Sprintf("%d%%", percentage)
	cr.SetSourceRGBA(0.1, 0.1, 0.1, 1.0)
	cr.SelectFontFace("sans-serif", cairo.FontSlantNormal, cairo.FontWeightBold)
	cr.SetFontSize(12.0)

	extents := cr.TextExtents(text)
	x := w/2 - (extents.Width/2 + extents.XBearing)
	y := h/2 - (extents.Height/2 + extents.YBearing)

	cr.MoveTo(x, y)
	cr.ShowText(text)
}
