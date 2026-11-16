package breadcrumb

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func deepNav(layers int) *Nav {
	n := New("Root", view("root"))
	for i := 1; i < layers; i++ {
		n.Push(fmt.Sprintf("Layer number %d", i), view("v"))
	}
	return n
}

func rows(b *bar) int {
	ys := map[float32]bool{}
	for _, s := range b.steps {
		ys[s.crumb.Position().Y] = true
	}
	return len(ys)
}

func TestBarRunsInOneRowWhenItFits(t *testing.T) {
	test.NewTempApp(t)
	n := deepNav(3)
	n.bar.Resize(fyne.NewSize(5000, n.bar.MinSize().Height))

	if r := rows(n.bar); r != 1 {
		t.Fatalf("rows = %d, want 1", r)
	}
	var x float32 = -1
	for _, s := range n.bar.steps {
		if s.crumb.Position().X <= x {
			t.Fatal("steps are not left to right")
		}
		x = s.crumb.Position().X
	}
}

func TestBarWrapsWhenNarrow(t *testing.T) {
	test.NewTempApp(t)
	n := deepNav(6)
	oneRow := n.bar.MinSize().Height // no width yet: single row

	n.bar.Resize(fyne.NewSize(n.bar.MinSize().Width, oneRow))
	if r := rows(n.bar); r < 2 {
		t.Fatalf("rows at the narrowest width = %d, want several", r)
	}
	if h := n.bar.MinSize().Height; h <= oneRow {
		t.Fatalf("min height %v did not grow beyond one row (%v) after wrapping", h, oneRow)
	}
}

func TestBarNeverNarrowerThanWidestStep(t *testing.T) {
	test.NewTempApp(t)
	n := deepNav(4)
	widest := n.bar.widestUnit()
	if got := n.bar.MinSize().Width; got != widest {
		t.Fatalf("min width = %v, want widest step %v", got, widest)
	}

	// Squeezed below the minimum the bar lays out as if it had the minimum.
	n.bar.Resize(fyne.NewSize(1, 1))
	if h := n.bar.MinSize().Height; h != n.bar.arrange(widest, false) {
		t.Fatalf("min height %v does not match the layout at the widest step", h)
	}
}

func TestSeparatorStaysWithItsCrumb(t *testing.T) {
	test.NewTempApp(t)
	n := deepNav(6)
	n.bar.Resize(fyne.NewSize(n.bar.MinSize().Width, 100))

	for i, s := range n.bar.steps {
		if s.sep == nil {
			continue
		}
		if s.sep.Position().X <= s.crumb.Position().X {
			t.Fatalf("step %d: separator left of its crumb", i)
		}
		cy := s.crumb.Position().Y + s.crumb.Size().Height/2
		sy := s.sep.Position().Y + s.sep.Size().Height/2
		if cy != sy {
			t.Fatalf("step %d: separator on another row than its crumb", i)
		}
	}
}

func TestBarRendererObjectsFollowSteps(t *testing.T) {
	test.NewTempApp(t)
	n := deepNav(3)
	r := test.TempWidgetRenderer(t, n.bar)
	// 3 crumbs + 2 separators
	if got := len(r.Objects()); got != 5 {
		t.Fatalf("objects = %d, want 5", got)
	}
	n.PopTo(0)
	if got := len(r.Objects()); got != 1 {
		t.Fatalf("objects after PopTo(0) = %d, want 1", got)
	}
	r.Destroy()
}
