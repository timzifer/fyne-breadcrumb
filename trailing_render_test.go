package breadcrumb

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// Trailing element sits once in the row; the host (content) keeps its size.
func TestTrailingRenderOnce(t *testing.T) {
	test.NewTempApp(t)
	body := widget.NewLabel("body")
	nav := New("Root", body)
	trailing := widget.NewLabel("ampel")
	nav.SetTrailing(trailing)
	win := test.NewWindow(nav)
	defer win.Close()
	win.Resize(fyne.NewSize(800, 400))
	nav.Refresh()
	r := test.WidgetRenderer(nav)
	r.Layout(fyne.NewSize(800, 400))
	count := 0
	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == trailing {
			count++
		}
		if c, ok := o.(*fyne.Container); ok {
			for _, x := range c.Objects {
				walk(x)
			}
		}
	}
	for _, o := range r.Objects() {
		walk(o)
	}
	if count != 1 {
		t.Fatalf("trailing element rendered %d times, want 1", count)
	}
	if body.Size().Height == 0 {
		t.Fatalf("content height is 0: content is not laid out")
	}
}

// After a push the new layer's view is laid out with the trailing element set.
func TestTrailingKeepsPushedLayerLaidOut(t *testing.T) {
	test.NewTempApp(t)
	nav := New("Root", widget.NewLabel("root"))
	nav.SetTrailing(widget.NewLabel("ampel"))
	win := test.NewWindow(nav)
	defer win.Close()
	win.Resize(fyne.NewSize(800, 400))
	view := widget.NewLabel("pushed view")
	nav.Push("Area", view)
	nav.Refresh()
	r := test.WidgetRenderer(nav)
	r.Layout(fyne.NewSize(800, 400))
	if view.Size().Height == 0 {
		t.Fatalf("pushed view height is 0: not laid out")
	}
}
