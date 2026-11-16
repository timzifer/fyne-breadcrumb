package breadcrumb

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// The GL driver only knows the canvas of objects it has already drawn, and
// canvas.Refresh on any other object is dropped. The test driver hands out a
// canvas for every object, so it would hide a widget that replaces its
// children without telling the canvas about itself. These wrappers restore
// the GL behaviour and record what reaches the canvas.

type glLikeApp struct {
	fyne.App
	drv *glLikeDriver
}

func (a *glLikeApp) Driver() fyne.Driver { return a.drv }

type glLikeDriver struct {
	fyne.Driver
	canvas *spyCanvas
	drawn  map[fyne.CanvasObject]bool
}

func (d *glLikeDriver) CanvasForObject(o fyne.CanvasObject) fyne.Canvas {
	if d.drawn[o] {
		return d.canvas
	}
	return nil
}

type spyCanvas struct {
	fyne.Canvas
	refreshed map[fyne.CanvasObject]int
}

func (c *spyCanvas) Refresh(o fyne.CanvasObject) {
	c.refreshed[o]++
	c.Canvas.Refresh(o)
}

// glLike shows n in a window, then switches to a driver that only knows the
// objects drawn so far: the nav and its bar. Every bar child is replaced on
// a rebuild, so none of them counts as drawn.
func glLike(t *testing.T, n *Nav) *spyCanvas {
	t.Helper()
	a := test.NewTempApp(t)
	w := test.NewWindow(n)
	t.Cleanup(w.Close)
	spy := &spyCanvas{Canvas: w.Canvas(), refreshed: map[fyne.CanvasObject]int{}}
	fyne.SetCurrentApp(&glLikeApp{App: a, drv: &glLikeDriver{
		Driver: a.Driver(),
		canvas: spy,
		drawn:  map[fyne.CanvasObject]bool{n: true, n.bar: true},
	}})
	t.Cleanup(func() { fyne.SetCurrentApp(a) })
	return spy
}

func TestBarOnlyChangesReachTheCanvas(t *testing.T) {
	for name, change := range map[string]func(n *Nav, top *Layer){
		"SetTitle":          func(_ *Nav, top *Layer) { top.SetTitle("Lathe") },
		"SetChildren":       func(_ *Nav, top *Layer) { top.SetChildren([]Entry{{Title: "x"}}) },
		"SetHistoryButtons": func(n *Nav, _ *Layer) { n.SetHistoryButtons(true) },
		"SetMaxTitleWidth":  func(n *Nav, _ *Layer) { n.SetMaxTitleWidth(10) },
		"ClearHistory":      func(n *Nav, _ *Layer) { n.ClearHistory() },
	} {
		t.Run(name, func(t *testing.T) {
			test.NewTempApp(t)
			n := New("Plant", view("plant"))
			top := n.Push("…", view("machine"))
			spy := glLike(t, n)

			change(n, top)
			if spy.refreshed[n.bar] == 0 {
				t.Fatal("the bar did not ask the canvas to repaint; the change stays invisible until something else redraws the window")
			}
		})
	}
}
