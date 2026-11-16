package breadcrumb

import (
	"slices"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func view(name string) *widget.Label { return widget.NewLabel(name) }

func assertTitles(t *testing.T, n *Nav, want ...string) {
	t.Helper()
	if got := n.Titles(); !slices.Equal(got, want) {
		t.Fatalf("titles = %q, want %q", got, want)
	}
	if n.Depth() != len(want) {
		t.Fatalf("depth = %d, want %d", n.Depth(), len(want))
	}
}

func crumbTexts(n *Nav) []string {
	var texts []string
	for _, s := range n.bar.steps {
		texts = append(texts, s.crumb.Text)
	}
	return texts
}

func TestNewShowsRoot(t *testing.T) {
	test.NewTempApp(t)
	root := view("root")
	n := New("List", root)

	assertTitles(t, n, "List")
	if n.Top().View() != root {
		t.Fatal("root is not on top")
	}
	if len(n.host.Objects) != 1 || n.host.Objects[0] != root {
		t.Fatal("host does not show the root")
	}
	if n.bar.steps[0].sep != nil {
		t.Fatal("a lone root without children must not show a separator")
	}
	if n.bar.steps[0].crumb.Importance != widget.HighImportance {
		t.Fatal("the current step must be highlighted")
	}
}

func TestPushShowsNewLayerAndKeepsOldViews(t *testing.T) {
	test.NewTempApp(t)
	root, order := view("root"), view("order")
	n := New("List", root)
	n.Push("Order 66340", order)

	assertTitles(t, n, "List", "Order 66340")
	if n.Top().View() != order || n.host.Objects[0] != order {
		t.Fatal("pushed view is not shown")
	}
	if n.stack[0].view != root {
		t.Fatal("root view was not retained")
	}
	if !slices.Equal(crumbTexts(n), []string{"List", "Order 66340"}) {
		t.Fatalf("crumbs = %q", crumbTexts(n))
	}
	first, last := n.bar.steps[0], n.bar.steps[1]
	if first.crumb.Importance != widget.LowImportance || last.crumb.Importance != widget.HighImportance {
		t.Fatal("only the last step may be highlighted")
	}
	if first.sep == nil || first.sep.Text != separator {
		t.Fatal("steps must be separated by a plain ›")
	}
	if last.sep != nil {
		t.Fatal("the last step must not end in a separator")
	}
	if first.siblings != nil || last.siblings != nil {
		t.Fatal("sibling chevron without siblings")
	}
	if last.crumb.OnTapped != nil {
		t.Fatal("the current step must not be tappable")
	}
}

func TestPopToDropsLayersAbove(t *testing.T) {
	test.NewTempApp(t)
	root := view("root")
	n := New("A", root)
	n.Push("B", view("b"))
	n.Push("C", view("c"))
	n.Push("D", view("d"))

	n.PopTo(1)
	assertTitles(t, n, "A", "B")

	n.PopTo(0)
	assertTitles(t, n, "A")
	if n.Top().View() != root {
		t.Fatal("root is not shown after PopTo(0)")
	}
}

func TestPopToIgnoresInvalidIndex(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Push("B", view("b"))

	for _, i := range []int{-1, 2, 99, 1} {
		n.PopTo(i)
		assertTitles(t, n, "A", "B")
	}
}

func TestPopStopsAtRoot(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Push("B", view("b"))

	if !n.Pop() {
		t.Fatal("Pop with two layers reported nothing removed")
	}
	assertTitles(t, n, "A")
	if n.Pop() {
		t.Fatal("Pop removed the root")
	}
	assertTitles(t, n, "A")
}

func TestTappingCrumbJumpsBack(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Push("B", view("b"))
	n.Push("C", view("c"))
	w := test.NewWindow(n)
	defer w.Close()

	test.Tap(n.bar.steps[1].crumb)
	assertTitles(t, n, "A", "B")

	test.Tap(n.bar.steps[1].crumb) // current step: no-op
	assertTitles(t, n, "A", "B")

	test.Tap(n.bar.steps[0].crumb)
	assertTitles(t, n, "A")
}

func TestReturnedViewKeepsItsState(t *testing.T) {
	test.NewTempApp(t)
	list := widget.NewList(
		func() int { return 100 },
		func() fyne.CanvasObject { return widget.NewLabel("row") },
		func(widget.ListItemID, fyne.CanvasObject) {},
	)
	list.Select(42)
	n := New("List", list)
	n.Push("Detail", view("detail"))
	n.PopTo(0)

	if n.Top().View() != list {
		t.Fatal("returned to a different view object")
	}
	selected := -1
	list.OnSelected = func(id widget.ListItemID) { selected = id }
	list.Select(42) // already selected: no callback
	if selected != -1 {
		t.Fatal("selection was lost while the list was covered")
	}
}

func TestNilViewPanics(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	defer func() {
		if recover() == nil {
			t.Fatal("Push with a nil view did not panic")
		}
	}()
	n.Push("B", nil)
}

func TestNavRendersBarAboveView(t *testing.T) {
	test.NewTempApp(t)
	root := view("root")
	n := New("A", root)
	w := test.NewWindow(n)
	defer w.Close()
	w.Resize(fyne.NewSize(400, 300))

	if y := n.host.Position().Y; y < n.bar.Size().Height {
		t.Fatalf("view at y=%v overlaps bar of height %v", y, n.bar.Size().Height)
	}
	if root.Size().Height <= 0 {
		t.Fatal("view got no height")
	}
	if n.bar.Size().Height <= 0 {
		t.Fatal("bar has no height")
	}
}

func TestNavInsideContainer(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	c := container.NewStack(n)
	w := test.NewWindow(c)
	defer w.Close()
	n.Push("B", view("b"))
	if w.Canvas().Content() != c {
		t.Fatal("unexpected content")
	}
}
