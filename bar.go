package breadcrumb

import (
	"math"

	"fyne.io/fyne/v2"
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

// rebuild recreates the steps from the navigation stack.
func (b *bar) rebuild() {
	stack := b.nav.stack
	last := len(stack) - 1
	b.steps = make([]step, len(stack))
	for i, l := range stack {
		s := step{crumb: widget.NewButton(l.title, nil)}
		if i == last {
			s.crumb.Importance = widget.HighImportance
		} else {
			idx := i
			s.crumb.Importance = widget.LowImportance
			s.crumb.OnTapped = func() { b.nav.PopTo(idx) }
		}
		switch {
		case l.chevron:
			idx := i
			chevron := widget.NewButton(separator, nil)
			chevron.Importance = widget.LowImportance
			chevron.OnTapped = func() { b.nav.showChildren(idx, chevron) }
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
	widest := b.widestStep()
	if width <= 0 {
		width = math.MaxFloat32
	} else if width < widest {
		width = widest
	}
	return fyne.NewSize(widest, b.arrange(width, false))
}

func (b *bar) widestStep() float32 {
	pad := theme.Padding()
	var w float32
	for _, s := range b.steps {
		w = max(w, s.width(pad))
	}
	return w
}

func (s step) width(pad float32) float32 {
	w := s.crumb.MinSize().Width
	if s.sep != nil {
		w += pad + s.sep.MinSize().Width
	}
	return w
}

func (s step) height() float32 {
	h := s.crumb.MinSize().Height
	if s.sep != nil {
		h = max(h, s.sep.MinSize().Height)
	}
	return h
}

// arrange flows the steps into rows of the given width and returns the
// height they take. With place set it also moves and sizes the objects.
func (b *bar) arrange(width float32, place bool) float32 {
	pad := theme.Padding()
	var x, y, rowH float32
	for _, s := range b.steps {
		w, h := s.width(pad), s.height()
		if x > 0 && x+w > width {
			x, y, rowH = 0, y+rowH+pad, 0
		}
		if place {
			cs := s.crumb.MinSize()
			s.crumb.Move(fyne.NewPos(x, y+(h-cs.Height)/2))
			s.crumb.Resize(cs)
			if s.sep != nil {
				ss := s.sep.MinSize()
				s.sep.Move(fyne.NewPos(x+cs.Width+pad, y+(h-ss.Height)/2))
				s.sep.Resize(ss)
			}
		}
		x += w + pad
		rowH = max(rowH, h)
	}
	return y + rowH
}
