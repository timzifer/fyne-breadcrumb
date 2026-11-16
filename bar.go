package breadcrumb

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// separator is the text between two steps and on a chevron.
const separator = "›"

// bar is the breadcrumb row. It runs in one line while it fits and wraps
// between steps when the window gets narrow.
type bar struct {
	widget.BaseWidget

	nav   *Nav
	steps []step

	// lead holds the back and forward buttons, or is nil.
	lead          *fyne.Container
	back, forward *widget.Button
}

// step is one layer in the bar: its crumb and the separator after it. A step
// never wraps between crumb and separator.
type step struct {
	crumb *widget.Button
	// sep is a *widget.Button (a chevron with a menu), a *widget.Label
	// (a plain separator) or nil (the last step without a menu).
	sep fyne.CanvasObject
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
	b.steps = make([]step, len(n.stack))
	for i, l := range n.stack {
		s := step{crumb: widget.NewButton(shorten(l.title, n.maxTitleWidth), nil)}
		if i == last {
			s.crumb.Importance = widget.HighImportance
		} else {
			idx := i
			s.crumb.Importance = widget.LowImportance
			s.crumb.OnTapped = func() { n.PopTo(idx) }
		}
		switch {
		case l.chevron:
			idx := i
			chevron := widget.NewButton(separator, nil)
			chevron.Importance = widget.LowImportance
			chevron.OnTapped = func() { n.showChildren(idx, chevron) }
			s.sep = chevron
		case i < last:
			s.sep = widget.NewLabel(separator)
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
		r.objects = append(r.objects, s.crumb)
		if s.sep != nil {
			r.objects = append(r.objects, s.sep)
		}
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

// MinSize is as wide as the widest step and as high as the rows the steps
// need at the bar's current width. Before the bar has a width it asks for a
// single row; once it is laid out narrower, Fyne notices the taller minimum
// and lays the parent out again.
func (r *barRenderer) MinSize() fyne.Size {
	b := r.bar
	width := b.Size().Width
	widest := b.widestUnit()
	if width <= 0 {
		width = math.MaxFloat32
	} else if width < widest {
		width = widest
	}
	return fyne.NewSize(widest, b.arrange(width, false))
}

// unit is what arrange flows: the history buttons or one step.
type unit struct {
	a, b fyne.CanvasObject // b may be nil
}

func (b *bar) units() []unit {
	units := make([]unit, 0, len(b.steps)+1)
	if b.lead != nil {
		units = append(units, unit{a: b.lead})
	}
	for _, s := range b.steps {
		units = append(units, unit{a: s.crumb, b: s.sep})
	}
	return units
}

func (b *bar) widestUnit() float32 {
	pad := theme.Padding()
	var w float32
	for _, u := range b.units() {
		w = max(w, u.width(pad))
	}
	return w
}

func (u unit) width(pad float32) float32 {
	w := u.a.MinSize().Width
	if u.b != nil {
		w += pad + u.b.MinSize().Width
	}
	return w
}

func (u unit) height() float32 {
	h := u.a.MinSize().Height
	if u.b != nil {
		h = max(h, u.b.MinSize().Height)
	}
	return h
}

// arrange flows the units into rows of the given width and returns the
// height they take. With place set it also moves and sizes the objects.
func (b *bar) arrange(width float32, place bool) float32 {
	pad := theme.Padding()
	var x, y, rowH float32
	for _, u := range b.units() {
		w, h := u.width(pad), u.height()
		if x > 0 && x+w > width {
			x, y, rowH = 0, y+rowH+pad, 0
		}
		if place {
			as := u.a.MinSize()
			u.a.Move(fyne.NewPos(x, y+(h-as.Height)/2))
			u.a.Resize(as)
			if u.b != nil {
				bs := u.b.MinSize()
				u.b.Move(fyne.NewPos(x+as.Width+pad, y+(h-bs.Height)/2))
				u.b.Resize(bs)
			}
		}
		x += w + pad
		rowH = max(rowH, h)
	}
	return y + rowH
}
