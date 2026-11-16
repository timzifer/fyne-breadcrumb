package breadcrumb

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// separator is the text between two steps.
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

// step is one layer in the bar. It never wraps inside.
type step struct {
	crumb *widget.Button
	// siblings opens the menu of the layer's siblings — the entries its
	// parent offers. nil when the parent offers none, and always for the
	// root, which has no parent.
	siblings *widget.Button
	// sep is the separator to the next step, nil on the last one.
	sep *widget.Label
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
			s.crumb.Importance = widget.LowImportance
			s.crumb.OnTapped = func() { n.PopTo(i) }
		}
		if i > 0 && n.stack[i-1].chevron {
			parent := i - 1
			s.siblings = widget.NewButtonWithIcon("", siblingsIcon, nil)
			s.siblings.Importance = widget.LowImportance
			s.siblings.OnTapped = func() { n.showChildren(parent, s.siblings) }
		}
		if i < last {
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
	for _, u := range r.bar.units() {
		for _, p := range u {
			r.objects = append(r.objects, p.obj)
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

// unit is what arrange flows and never breaks: the history buttons, or one
// step's crumb, sibling chevron and separator.
type unit []piece

// piece is one object of a unit. A glued piece sits right against the one
// before it: the sibling chevron belongs to its title, not between steps.
type piece struct {
	obj  fyne.CanvasObject
	glue bool
}

func (b *bar) units() []unit {
	units := make([]unit, 0, len(b.steps)+1)
	if b.lead != nil {
		units = append(units, unit{{obj: b.lead}})
	}
	for _, s := range b.steps {
		u := unit{{obj: s.crumb}}
		if s.siblings != nil {
			u = append(u, piece{obj: s.siblings, glue: true})
		}
		if s.sep != nil {
			u = append(u, piece{obj: s.sep})
		}
		units = append(units, u)
	}
	return units
}

// gap is the space before piece i of a unit.
func (u unit) gap(i int, pad float32) float32 {
	if i == 0 || u[i].glue {
		return 0
	}
	return pad
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
	var w float32
	for i, p := range u {
		w += u.gap(i, pad) + p.obj.MinSize().Width
	}
	return w
}

func (u unit) height() float32 {
	var h float32
	for _, p := range u {
		h = max(h, p.obj.MinSize().Height)
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
			ox := x
			for i, p := range u {
				ox += u.gap(i, pad)
				s := p.obj.MinSize()
				p.obj.Move(fyne.NewPos(ox, y+(h-s.Height)/2))
				p.obj.Resize(s)
				ox += s.Width
			}
		}
		x += w + pad
		rowH = max(rowH, h)
	}
	return y + rowH
}
