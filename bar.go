package breadcrumb

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// bar is the breadcrumb row: arrow segments that nest into each other. It
// runs in one line while it fits and wraps between steps when the window
// gets narrow.
type bar struct {
	widget.BaseWidget

	nav   *Nav
	steps []*segment

	// lead holds the back and forward buttons, or is nil.
	lead          *fyne.Container
	back, forward *widget.Button
}

func newBar(n *Nav) *bar {
	b := &bar{nav: n}
	b.ExtendBaseWidget(b)
	return b
}

// rebuild recreates the bar from the navigation state.
func (b *bar) rebuild() {
	n := b.nav
	b.lead, b.back, b.forward = nil, nil, nil
	if n.historyButtons {
		b.back = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { n.Back() })
		b.forward = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { n.Forward() })
		for _, btn := range []*widget.Button{b.back, b.forward} {
			btn.Importance = widget.LowImportance
		}
		if !n.CanBack() {
			b.back.Disable()
		}
		if !n.CanForward() {
			b.forward.Disable()
		}
		b.lead = container.NewHBox(b.back, b.forward)
	}

	last := len(n.stack) - 1
	b.steps = make([]*segment, len(n.stack))
	for i, l := range n.stack {
		s := newSegment(shorten(l.title, n.maxTitleWidth))
		s.icon, s.iconAt = l.icon, l.iconAt
		s.first = i == 0
		s.current = i == last
		if !s.current {
			s.onTap = func() { n.PopTo(i) }
		}
		if i > 0 {
			s.prev = b.steps[i-1]
			if n.stack[i-1].chevron {
				parent := i - 1
				s.onSiblings = func(at fyne.Position) { n.showChildren(parent, s, at) }
			}
		}
		b.steps[i] = s
	}
	b.Refresh()
}

// CreateRenderer is part of [fyne.Widget].
func (b *bar) CreateRenderer() fyne.WidgetRenderer {
	return &barRenderer{bar: b}
}

type barRenderer struct {
	bar     *bar
	objects []fyne.CanvasObject
}

func (r *barRenderer) Destroy() {}

func (r *barRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *barRenderer) Refresh() {
	r.objects = r.objects[:0]
	if r.bar.lead != nil {
		r.objects = append(r.objects, r.bar.lead)
	}
	for _, s := range r.bar.steps {
		r.objects = append(r.objects, s)
	}
	r.Layout(r.bar.Size())
	for _, o := range r.objects {
		o.Refresh()
	}
	// The children may all be new. The GL driver only knows the canvas of
	// objects it has drawn before, so their own refreshes are dropped, and
	// BaseWidget.Refresh does not tell the canvas either: without this, a
	// change that touches only the bar (an async SetTitle, say) stays
	// invisible until something else redraws the window.
	canvas.Refresh(r.bar)
}

func (r *barRenderer) Layout(size fyne.Size) {
	r.bar.arrange(size.Width, true)
}

// MinSize is as wide as the widest item and as high as the rows the items
// need at the bar's current width. Before the bar has a width it asks for a
// single row; once it is laid out narrower, Fyne notices the taller minimum
// and lays the parent out again.
func (r *barRenderer) MinSize() fyne.Size {
	b := r.bar
	width := b.Size().Width
	widest := b.widestItem()
	if width <= 0 {
		width = math.MaxFloat32
	} else if width < widest {
		width = widest
	}
	return fyne.NewSize(widest, b.arrange(width, false))
}

// items are what arrange flows: the history buttons, then the segments.
func (b *bar) items() []fyne.CanvasObject {
	items := make([]fyne.CanvasObject, 0, len(b.steps)+1)
	if b.lead != nil {
		items = append(items, b.lead)
	}
	for _, s := range b.steps {
		items = append(items, s)
	}
	return items
}

func (b *bar) widestItem() float32 {
	var w float32
	for _, o := range b.items() {
		w = max(w, o.MinSize().Width)
	}
	return w
}

// arrange flows the items into rows of the given width and returns the
// height they take. With place set it also moves and sizes them. A segment
// following another reaches back under its point, so the two nest with the
// navigation's segment gap between point and notch.
func (b *bar) arrange(width float32, place bool) float32 {
	pad := theme.Padding()
	gap := b.nav.SegmentGap()
	var x, y, rowH float32
	var prevSeg bool
	for _, o := range b.items() {
		s := o.MinSize()
		_, isSeg := o.(*segment)
		if x > 0 && isSeg && prevSeg {
			// x is one padding past the point before; move back so the notch
			// starts gap after it.
			x += gap - pad - s.Height*arrowRatio
		}
		if x > 0 && x+s.Width > width {
			x, y, rowH = 0, y+rowH+pad, 0
		}
		if place {
			o.Move(fyne.NewPos(x, y))
			o.Resize(s)
		}
		x += s.Width + pad
		rowH = max(rowH, s.Height)
		prevSeg = isSeg
	}
	return y + rowH
}
