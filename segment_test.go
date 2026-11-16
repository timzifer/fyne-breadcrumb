package breadcrumb

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

// laidOut returns a nav A > B > C (B offering siblings for C) in a window,
// so the segments have their real sizes and positions.
func laidOut(t *testing.T) *Nav {
	t.Helper()
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Push("B", view("b"), WithChildren(func() []Entry { return []Entry{{Title: "C"}, {Title: "D"}} }))
	n.Push("C", view("c"))
	w := test.NewWindow(n)
	t.Cleanup(w.Close)
	w.Resize(fyne.NewSize(600, 200))
	return n
}

func alphaAt(r *segmentRenderer, x, y int, s fyne.Size) uint8 {
	_, _, _, a := r.pixel(x, y, int(s.Width), int(s.Height)).RGBA()
	return uint8(a >> 8)
}

func TestSegmentShape(t *testing.T) {
	n := laidOut(t)
	for i, s := range n.bar.steps {
		r := test.TempWidgetRenderer(t, s).(*segmentRenderer)
		size := s.Size()
		w, h := int(size.Width), int(size.Height)
		mid := h / 2

		if alphaAt(r, w/2, mid, size) != r.fill.A {
			t.Fatalf("step %d: middle not filled", i)
		}
		// The point: filled at mid height up to the right edge, empty in
		// the top right corner.
		if alphaAt(r, w-2, mid, size) == 0 {
			t.Fatalf("step %d: no point at the right", i)
		}
		if alphaAt(r, w-1, 0, size) != 0 {
			t.Fatalf("step %d: top right corner filled", i)
		}
		// The left: straight for the first step, a notch for the others.
		left := alphaAt(r, 1, mid, size)
		if s.first && left != r.fill.A {
			t.Fatal("first step has no straight left edge")
		}
		if !s.first && left != 0 {
			t.Fatalf("step %d: no notch at the left", i)
		}
		if alphaAt(r, 0, 0, size) == 0 {
			t.Fatalf("step %d: top left corner empty", i)
		}
	}
}

func TestSegmentEdgesAreAntiAliased(t *testing.T) {
	n := laidOut(t)
	s := n.bar.steps[1]
	r := test.TempWidgetRenderer(t, s).(*segmentRenderer)
	size := s.Size()
	y := int(size.Height) / 4
	partial := false
	for x := int(size.Width) - int(s.depth()) - 2; x < int(size.Width); x++ {
		if a := alphaAt(r, x, y, size); a > 0 && a < r.fill.A {
			partial = true
		}
	}
	if !partial {
		t.Fatal("slanted edge has no partly covered pixels")
	}
}

func TestTapInNotchGoesToStepBefore(t *testing.T) {
	n := laidOut(t)
	n.Push("D", view("d"))
	assertTitles(t, n, "A", "B", "C", "D")

	// Left of C's notch edge, at mid height, lies the point of B.
	c := n.bar.steps[2]
	test.TapAt(c, fyne.NewPos(1, c.Size().Height/2))
	assertTitles(t, n, "A", "B")
}

func TestTapOnTitleAndChevronZone(t *testing.T) {
	n := laidOut(t)
	c := n.bar.steps[2] // current, with siblings

	test.TapAt(c, fyne.NewPos(c.chevronZone()-2, c.Size().Height/2))
	if n.menu != nil {
		t.Fatal("a tap on the title opened the menu")
	}
	assertTitles(t, n, "A", "B", "C") // current step: title tap does nothing

	tapChevron(c)
	if n.menu == nil || n.menu.layer != n.Layer(1) {
		t.Fatal("a tap on the chevron did not open the siblings of C")
	}
	labels, checked := menuLabels(n.childMenu(1))
	if len(labels) != 2 || !checked[0] {
		t.Fatalf("menu %q %v, want C checked", labels, checked)
	}
}

func TestSegmentKeys(t *testing.T) {
	n := laidOut(t)
	b, c := n.bar.steps[1], n.bar.steps[2]

	c.TypedRune('x') // ignored
	c.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	if n.menu == nil {
		t.Fatal("Down did not open the siblings")
	}
	n.closeMenu()
	b.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown}) // B has no siblings
	if n.menu != nil {
		t.Fatal("Down opened a menu on a step without siblings")
	}
	c.TypedKey(&fyne.KeyEvent{Name: fyne.KeySpace}) // current: nothing
	assertTitles(t, n, "A", "B", "C")
	b.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	assertTitles(t, n, "A", "B")
}

func TestSegmentColoursFollowTheTheme(t *testing.T) {
	n := laidOut(t)
	th := n.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	nrgba := func(c color.Color) color.NRGBA { return color.NRGBAModel.Convert(c).(color.NRGBA) }

	a, c := n.bar.steps[0], n.bar.steps[2]
	ra := test.TempWidgetRenderer(t, a).(*segmentRenderer)
	rc := test.TempWidgetRenderer(t, c).(*segmentRenderer)

	if ra.fill != nrgba(th.Color(theme.ColorNameButton, v)) || ra.text.Color != th.Color(theme.ColorNameForeground, v) {
		t.Fatal("other steps are not in the button colours")
	}
	if rc.fill != nrgba(th.Color(theme.ColorNamePrimary, v)) || rc.text.Color != th.Color(theme.ColorNameForegroundOnPrimary, v) {
		t.Fatal("the current step is not in the primary colours")
	}

	plain := ra.fill
	a.MouseIn(&desktop.MouseEvent{})
	a.MouseMoved(&desktop.MouseEvent{})
	if ra.fill == plain {
		t.Fatal("hover does not show")
	}
	a.MouseOut()
	if ra.fill != plain {
		t.Fatal("hover stays after the mouse left")
	}
	a.FocusGained()
	if ra.fill == plain {
		t.Fatal("focus does not show")
	}
	a.FocusLost()
	if ra.fill != plain {
		t.Fatal("focus stays after it was lost")
	}

	primary := rc.fill
	c.MouseIn(&desktop.MouseEvent{})
	if rc.fill != primary {
		t.Fatal("the current step reacts to hover")
	}
}

func TestSegmentObjects(t *testing.T) {
	n := laidOut(t)
	if got := len(test.TempWidgetRenderer(t, n.bar.steps[1]).Objects()); got != 2 {
		t.Fatalf("step without siblings has %d objects, want shape and title", got)
	}
	if got := len(test.TempWidgetRenderer(t, n.bar.steps[2]).Objects()); got != 3 {
		t.Fatalf("step with siblings has %d objects, want shape, title and chevron", got)
	}
}

func TestSegmentWidthMakesRoomForChevron(t *testing.T) {
	test.NewTempApp(t)
	plain, withChevron := newSegment("Hall"), newSegment("Hall")
	withChevron.onSiblings = func(fyne.Position) {}
	if withChevron.MinSize().Width <= plain.MinSize().Width {
		t.Fatal("no room for the chevron")
	}
	first := newSegment("Hall")
	first.first = true
	if first.MinSize().Width >= plain.MinSize().Width {
		t.Fatal("the first step does not save the notch")
	}
}

func TestOver(t *testing.T) {
	got := over(color.NRGBA{R: 0, G: 0, B: 0, A: 255}, color.NRGBA{R: 255, G: 255, B: 255, A: 128})
	c := color.NRGBAModel.Convert(got).(color.NRGBA)
	if c.A != 255 || c.R < 126 || c.R > 130 {
		t.Fatalf("over = %v", c)
	}
}
