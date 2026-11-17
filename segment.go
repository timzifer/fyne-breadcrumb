package breadcrumb

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// arrowRatio is the depth of a segment's point (and of the notch that takes
// the previous point) relative to its height.
const arrowRatio = 0.3

// segment is one step of the bar, drawn as an arrow: a notch on the left
// that takes the point of the step before, a point on the right. The first
// step has a straight left edge.
type segment struct {
	widget.BaseWidget

	title   string
	first   bool
	current bool

	// icon stands before (IconLeading) or after the title; nil shows none.
	icon   fyne.Resource
	iconAt IconPlacement
	// color fills the segment when set; the text takes black or white for contrast.
	color color.Color
	// neutral keeps the top step in the button colour; iconTint is the icon's theme colour.
	neutral  bool
	iconTint fyne.ThemeColorName

	// onTap jumps to this step; nil on the current one.
	onTap func()
	// onSiblings opens the sibling menu below the given position, relative
	// to the segment; nil when the step has no siblings.
	onSiblings func(at fyne.Position)
	// prev receives taps that land in the notch, which shows the point of
	// the step before even though it lies inside this segment's box.
	prev *segment

	hovered, focused bool
}

func newSegment(title string) *segment {
	s := &segment{title: title}
	s.ExtendBaseWidget(s)
	return s
}

// depth is how far the point reaches out, and the notch reaches in.
func (s *segment) depth() float32 { return s.Size().Height * arrowRatio }

var titleStyle = fyne.TextStyle{Bold: true}

// insets returns the space left of the title and right of the title or
// chevron at height h: the notch and the point take part of arrowRatio, the
// theme's inner padding comes on top.
func (s *segment) insets(h float32) (left, right float32) {
	d := h * arrowRatio
	left = theme.InnerPadding()
	if !s.first {
		left += d * 0.6
	}
	return left, d*0.6 + theme.InnerPadding()
}

func (s *segment) titleSize() fyne.Size {
	return fyne.MeasureText(s.title, theme.TextSize(), titleStyle)
}

// chevronX is where the sibling chevron starts, in segment coordinates.
func (s *segment) chevronX() float32 {
	left, _ := s.insets(s.Size().Height)
	return left + s.iconLead() + s.titleSize().Width + s.iconTrail() + theme.Padding()
}

// iconWidth is the room the icon and the gap to the title take, 0 without an icon.
func (s *segment) iconWidth() float32 {
	if s.icon == nil {
		return 0
	}
	return iconSize() + theme.Padding()
}

// iconLead is the room in front of the title taken by a leading icon.
func (s *segment) iconLead() float32 {
	if s.icon != nil && s.iconAt == IconLeading {
		return s.iconWidth()
	}
	return 0
}

// iconTrail is the room behind the title taken by a trailing icon.
func (s *segment) iconTrail() float32 {
	if s.icon != nil && s.iconAt == IconTrailing {
		return s.iconWidth()
	}
	return 0
}

func iconSize() float32 { return theme.IconInlineSize() }

// chevronZone is where the sibling zone starts; a tap at or right of it
// opens the menu.
func (s *segment) chevronZone() float32 { return s.chevronX() - theme.Padding()/2 }

// Tapped is part of [fyne.Tappable].
func (s *segment) Tapped(e *fyne.PointEvent) {
	if s.prev != nil && s.inNotch(e.Position) {
		s.prev.Tapped(&fyne.PointEvent{
			Position:         e.Position.AddXY(s.Position().X-s.prev.Position().X, s.Position().Y-s.prev.Position().Y),
			AbsolutePosition: e.AbsolutePosition,
		})
		return
	}
	if s.onSiblings != nil && e.Position.X >= s.chevronZone() {
		s.openSiblings()
		return
	}
	if s.onTap != nil {
		s.onTap()
	}
}

func (s *segment) openSiblings() {
	if s.onSiblings != nil {
		s.onSiblings(fyne.NewPos(s.chevronX(), s.Size().Height))
	}
}

// inNotch reports whether p lies left of the notch edge.
func (s *segment) inNotch(p fyne.Position) bool {
	if s.first {
		return false
	}
	h := s.Size().Height
	return p.X < s.depth()*(1-abs32(p.Y-h/2)/(h/2))
}

// MouseIn is part of [desktop.Hoverable].
func (s *segment) MouseIn(*desktop.MouseEvent) { s.setHovered(true) }

// MouseMoved is part of [desktop.Hoverable].
func (s *segment) MouseMoved(*desktop.MouseEvent) {}

// MouseOut is part of [desktop.Hoverable].
func (s *segment) MouseOut() { s.setHovered(false) }

func (s *segment) setHovered(h bool) {
	if s.hovered != h {
		s.hovered = h
		s.Refresh()
	}
}

// FocusGained is part of [fyne.Focusable].
func (s *segment) FocusGained() { s.focused = true; s.Refresh() }

// FocusLost is part of [fyne.Focusable].
func (s *segment) FocusLost() { s.focused = false; s.Refresh() }

// TypedRune is part of [fyne.Focusable].
func (s *segment) TypedRune(rune) {}

// TypedKey is part of [fyne.Focusable]: Space or Enter jumps to the step,
// Down opens its siblings.
func (s *segment) TypedKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeySpace, fyne.KeyReturn, fyne.KeyEnter:
		if s.onTap != nil {
			s.onTap()
		}
	case fyne.KeyDown:
		s.openSiblings()
	}
}

// CreateRenderer is part of [fyne.Widget].
func (s *segment) CreateRenderer() fyne.WidgetRenderer {
	r := &segmentRenderer{seg: s, text: canvas.NewText(s.title, color.Black)}
	r.text.TextStyle = titleStyle
	r.shape = canvas.NewRasterWithPixels(r.pixel)
	r.chevron = canvas.NewImageFromResource(siblingsIcon)
	r.chevron.FillMode = canvas.ImageFillContain
	r.Refresh()
	return r
}

type segmentRenderer struct {
	seg     *segment
	shape   *canvas.Raster
	text    *canvas.Text
	chevron *canvas.Image
	icon    *canvas.Image // nil until the segment has an icon once
	fill    color.NRGBA
}

func (r *segmentRenderer) Destroy() {}

func (r *segmentRenderer) Objects() []fyne.CanvasObject {
	out := []fyne.CanvasObject{r.shape, r.text}
	if r.seg.icon != nil && r.icon != nil {
		out = append(out, r.icon)
	}
	if r.seg.onSiblings != nil {
		out = append(out, r.chevron)
	}
	return out
}

func chevronSize() float32 { return theme.IconInlineSize() * 0.7 }

func (r *segmentRenderer) MinSize() fyne.Size {
	ts := r.seg.titleSize()
	h := ts.Height + 2*theme.InnerPadding()
	left, right := r.seg.insets(h)
	w := left + r.seg.iconWidth() + ts.Width + right
	if r.seg.onSiblings != nil {
		w += theme.Padding() + chevronSize()
	}
	return fyne.NewSize(w, h)
}

func (r *segmentRenderer) Layout(size fyne.Size) {
	r.shape.Resize(size)
	r.shape.Move(fyne.NewPos(0, 0))
	left, _ := r.seg.insets(size.Height)
	ts := r.text.MinSize()
	r.text.Move(fyne.NewPos(left+r.seg.iconLead(), (size.Height-ts.Height)/2))
	r.text.Resize(ts)
	if r.icon != nil {
		is := iconSize()
		x := left
		if r.seg.iconAt == IconTrailing {
			x = left + r.seg.iconLead() + ts.Width + theme.Padding()
		}
		r.icon.Resize(fyne.NewSquareSize(is))
		r.icon.Move(fyne.NewPos(x, (size.Height-is)/2))
	}
	cs := chevronSize()
	r.chevron.Resize(fyne.NewSquareSize(cs))
	r.chevron.Move(fyne.NewPos(r.seg.chevronX(), (size.Height-cs)/2))
}

func (r *segmentRenderer) Refresh() {
	s := r.seg
	th := s.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	bg, fg := theme.ColorNameButton, theme.ColorNameForeground
	if s.current && !s.neutral {
		bg, fg = theme.ColorNamePrimary, theme.ColorNameForegroundOnPrimary
	}
	fill := th.Color(bg, v)
	switch {
	case s.focused:
		fill = over(fill, th.Color(theme.ColorNameFocus, v))
	case s.hovered && !s.current:
		fill = over(fill, th.Color(theme.ColorNameHover, v))
	}
	r.fill = color.NRGBAModel.Convert(fill).(color.NRGBA)
	textColor := th.Color(fg, v)
	if s.color != nil {
		r.fill = color.NRGBAModel.Convert(s.color).(color.NRGBA)
		textColor = contrastOn(s.color)
	}

	r.text.Text = s.title
	r.text.Color = textColor
	r.text.TextSize = th.Size(theme.SizeNameText)
	r.chevron.Resource = theme.NewColoredResource(siblingsIcon, fg)
	if s.icon != nil {
		if r.icon == nil {
			r.icon = canvas.NewImageFromResource(s.icon)
			r.icon.FillMode = canvas.ImageFillContain
		}
		tint := fg
		if s.iconTint != "" {
			tint = s.iconTint
		}
		r.icon.Resource = theme.NewColoredResource(s.icon, tint)
		r.icon.Refresh()
	}

	r.Layout(s.Size())
	r.shape.Refresh()
	r.text.Refresh()
	r.chevron.Refresh()
}

// pixel draws the arrow with anti-aliased slanted edges. w and h are in
// device pixels, so the shape stays sharp at any scale.
func (r *segmentRenderer) pixel(x, y, w, h int) color.Color {
	fx, fy := float64(x)+0.5, float64(y)+0.5
	H, W := float64(h), float64(w)
	d := H * arrowRatio
	t := 1 - math.Abs(fy-H/2)/(H/2) // 0 at the edges, 1 in the middle
	// Distances along x scaled to distances across the slanted edges.
	k := (H / 2) / math.Hypot(H/2, d)
	cover := clamp((W-d+d*t-fx)*k + 0.5)
	if !r.seg.first {
		cover *= clamp((fx-d*t)*k + 0.5)
	}
	c := r.fill
	c.A = uint8(float64(c.A) * cover)
	return c
}

func clamp(v float64) float64 { return math.Max(0, math.Min(1, v)) }

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// over paints top over bottom, as Fyne paints its hover and focus colours
// over a button.
func over(bottom, top color.Color) color.Color {
	b := color.NRGBAModel.Convert(bottom).(color.NRGBA)
	t := color.NRGBAModel.Convert(top).(color.NRGBA)
	a := float64(t.A) / 255
	mix := func(bc, tc uint8) uint8 { return uint8(float64(bc)*(1-a) + float64(tc)*a + 0.5) }
	return color.NRGBA{R: mix(b.R, t.R), G: mix(b.G, t.G), B: mix(b.B, t.B), A: b.A}
}

// contrastOn is black or white, whichever reads better on c.
func contrastOn(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()
	lum := 0.299*float64(r)/65535 + 0.587*float64(g)/65535 + 0.114*float64(b)/65535
	if lum > 0.6 {
		return color.Black
	}
	return color.White
}
