package breadcrumb

import (
	"fmt"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
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
		ys[s.Position().Y] = true
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
		if s.Position().X <= x {
			t.Fatal("steps are not left to right")
		}
		x = s.Position().X
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
	widest := n.bar.widestItem()
	if got := n.bar.MinSize().Width; got != widest {
		t.Fatalf("min width = %v, want widest step %v", got, widest)
	}

	// Squeezed below the minimum the bar lays out as if it had the minimum.
	n.bar.Resize(fyne.NewSize(1, 1))
	if h := n.bar.MinSize().Height; h != n.bar.arrange(widest, false) {
		t.Fatalf("min height %v does not match the layout at the widest step", h)
	}
}

// assertGap checks that every notch starts want after the point before it,
// and that the segments nest: the next one starts before that point.
func assertGap(t *testing.T, n *Nav, want float32) {
	t.Helper()
	n.bar.Resize(fyne.NewSize(5000, n.bar.MinSize().Height))
	n.bar.Refresh()
	for i := 1; i < len(n.bar.steps); i++ {
		prev, s := n.bar.steps[i-1], n.bar.steps[i]
		point := prev.Position().X + prev.Size().Width
		notch := s.Position().X + s.depth()
		if s.Position().X >= point {
			t.Fatalf("step %d starts at %v, after the point before at %v: no nesting", i, s.Position().X, point)
		}
		if gap := notch - point; math.Abs(float64(gap-want)) > 0.01 {
			t.Fatalf("step %d: gap between point and notch = %v, want %v", i, gap, want)
		}
	}
}

func TestSegmentGap(t *testing.T) {
	test.NewTempApp(t)
	n := deepNav(3)
	def := theme.Padding() * 1.5
	if n.SegmentGap() != def {
		t.Fatalf("default gap = %v, want %v", n.SegmentGap(), def)
	}
	assertGap(t, n, def)

	n.SetSegmentGap(0) // touching
	if n.SegmentGap() != 0 {
		t.Fatal("gap 0 not kept")
	}
	assertGap(t, n, 0)

	n.SetSegmentGap(10)
	assertGap(t, n, 10)

	n.SetSegmentGap(-3) // back to the default
	if n.SegmentGap() != def {
		t.Fatalf("negative gap gave %v, want the default %v", n.SegmentGap(), def)
	}
	assertGap(t, n, def)
}

func TestBarRendererObjectsFollowSteps(t *testing.T) {
	test.NewTempApp(t)
	n := deepNav(3)
	r := test.TempWidgetRenderer(t, n.bar)
	if got := len(r.Objects()); got != 3 {
		t.Fatalf("objects = %d, want one per step", got)
	}
	n.PopTo(0)
	if got := len(r.Objects()); got != 1 {
		t.Fatalf("objects after PopTo(0) = %d, want 1", got)
	}
	r.Destroy()
}
