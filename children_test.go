package breadcrumb

import (
	"slices"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// plant builds a root "Plant" that offers halls; once a hall is open, its
// step carries the chevron listing them.
func plant(t *testing.T, halls ...string) (*Nav, *int) {
	t.Helper()
	asked := 0
	var n *Nav
	n = New("Plant", view("plant"), WithChildren(func() []Entry {
		asked++
		return hallEntries(&n, halls...)
	}))
	return n, &asked
}

func hallEntries(n **Nav, halls ...string) []Entry {
	entries := make([]Entry, len(halls))
	for i, h := range halls {
		entries[i] = Entry{Title: h, Open: func() { (*n).Push(h, view(h)) }}
	}
	return entries
}

func menuLabels(m *fyne.Menu) (labels []string, checked []bool) {
	if m == nil {
		return nil, nil
	}
	for _, it := range m.Items {
		labels = append(labels, it.Label)
		checked = append(checked, it.Checked)
	}
	return labels, checked
}

// shown opens, in a window, the chevron listing the entries layer i offers —
// the one on the step of the layer open above i — and returns the window.
func shown(t *testing.T, n *Nav, i int) fyne.Window {
	t.Helper()
	w := test.NewWindow(n)
	t.Cleanup(w.Close)
	w.Resize(fyne.NewSize(500, 300))
	test.Tap(chevronOf(t, n, i))
	return w
}

// chevronOf returns the sibling chevron that lists what layer i offers. It
// sits on the step above i, next to the title of the sibling that is open.
func chevronOf(t *testing.T, n *Nav, i int) *widget.Button {
	t.Helper()
	if i+1 >= len(n.bar.steps) {
		t.Fatalf("nothing open above layer %d: no step to carry its chevron", i)
	}
	b := n.bar.steps[i+1].siblings
	if b == nil {
		t.Fatalf("step %d has no sibling chevron", i+1)
	}
	return b
}

// topPopUp returns the chevron menu on screen, or nil. Fyne wraps a pop-up
// menu in an internal overlay container, so the canvas only tells whether
// some overlay is up.
func topPopUp(n *Nav, w fyne.Window) *widget.PopUpMenu {
	if n.menu == nil || n.menu.pop == nil || !n.menu.pop.Visible() || w.Canvas().Overlays().Top() == nil {
		return nil
	}
	return n.menu.pop
}

func TestSiblingChevronSitsOnTheChild(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1")
	if n.bar.steps[0].siblings != nil {
		t.Fatal("the root has no siblings, so no chevron")
	}

	n.Push("Hall 1", view("hall 1"))
	c := chevronOf(t, n, 0)
	if c.Icon != siblingsIcon || c.Text != "" {
		t.Fatalf("chevron shows %q / %v, want the down chevron icon only", c.Text, c.Icon)
	}
	if n.bar.steps[0].sep == nil || n.bar.steps[0].sep.Text != separator {
		t.Fatal("the separator between steps must stay a plain ›")
	}
	if n.bar.steps[1].sep != nil {
		t.Fatal("the last step must not end in a separator")
	}

	n.Push("Machine", view("m")) // Hall 1 offers nothing
	if n.bar.steps[2].siblings != nil {
		t.Fatal("a layer whose parent offers nothing got a chevron")
	}
	if n.bar.steps[1].siblings == nil {
		t.Fatal("chevron lost after a push")
	}
}

func TestSiblingChevronIsPlacedBetweenTitleAndSeparator(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1")
	n.Push("Hall 1", view("hall 1"))
	n.Push("Machine", view("m"))
	n.bar.Resize(fyne.NewSize(2000, n.bar.MinSize().Height))

	s := n.bar.steps[1]
	if !(s.crumb.Position().X < s.siblings.Position().X && s.siblings.Position().X < s.sep.Position().X) {
		t.Fatalf("order on the row: crumb %v, chevron %v, separator %v",
			s.crumb.Position().X, s.siblings.Position().X, s.sep.Position().X)
	}
	// The chevron belongs to its title: no gap between them, a gap before
	// the separator.
	if end := s.crumb.Position().X + s.crumb.Size().Width; s.siblings.Position().X != end {
		t.Fatalf("chevron at x=%v, title ends at %v", s.siblings.Position().X, end)
	}
	if end := s.siblings.Position().X + s.siblings.Size().Width; s.sep.Position().X <= end {
		t.Fatal("separator glued to the chevron")
	}
}

func TestListIsAskedJustInTime(t *testing.T) {
	test.NewTempApp(t)
	n, asked := plant(t, "Hall 1", "Hall 2")
	n.Push("Hall 1", view("hall 1"))
	if *asked != 0 {
		t.Fatalf("list asked %d times before the menu opened", *asked)
	}

	w := shown(t, n, 0)
	if *asked != 1 {
		t.Fatalf("list asked %d times, want once per opening", *asked)
	}
	if topPopUp(n, w) == nil {
		t.Fatal("no menu on screen")
	}
	labels, _ := menuLabels(n.childMenu(0))
	if !slices.Equal(labels, []string{"Hall 1", "Hall 2"}) {
		t.Fatalf("labels = %q", labels)
	}
}

func TestMenuOpensBelowChevron(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1")
	n.Push("Hall 1", view("hall 1"))
	w := shown(t, n, 0)

	chevron := chevronOf(t, n, 0)
	if pop := topPopUp(n, w); pop.Position().Y < chevron.Position().Y+chevron.Size().Height {
		t.Fatalf("menu at y=%v does not open below the chevron", pop.Position().Y)
	}
}

func TestMenuChecksOpenSibling(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1", "Hall 2", "Hall 3")
	n.Root().RefreshChildren()

	_, checked := menuLabels(n.childMenu(0))
	if slices.Contains(checked, true) {
		t.Fatal("entry checked although nothing is open below")
	}

	n.Push("Hall 2", view("hall 2"))
	_, checked = menuLabels(n.childMenu(0))
	if !slices.Equal(checked, []bool{false, true, false}) {
		t.Fatalf("checked = %v, want only Hall 2", checked)
	}
}

func TestPickingSiblingReplacesLayersAbove(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1", "Hall 2", "Hall 3")
	n.Push("Hall 1", view("hall 1"))
	n.Push("Machine A", view("machine a"))
	shown(t, n, 0)

	n.childMenu(0).Items[2].Action()
	assertTitles(t, n, "Plant", "Hall 3")
}

func TestEntryWithoutOpenJustJumpsBack(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Root().SetChildren([]Entry{{Title: "x"}})
	n.Push("B", view("b"))

	n.childMenu(0).Items[0].Action()
	assertTitles(t, n, "A")
}

func TestDisabledEntry(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Root().SetChildren([]Entry{{Title: "Loading…", Disabled: true}})
	if !n.childMenu(0).Items[0].Disabled {
		t.Fatal("disabled entry is enabled in the menu")
	}
}

func TestSetChildrenDecidesChevron(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Push("B", view("b"))
	if n.bar.steps[1].siblings != nil {
		t.Fatal("chevron before any siblings")
	}

	n.Root().SetChildren([]Entry{{Title: "B"}, {Title: "C"}})
	chevronOf(t, n, 0)

	n.Root().SetChildren(nil)
	if n.bar.steps[1].siblings != nil {
		t.Fatal("chevron must go away with nil children")
	}
}

func TestSetChildrenUpdatesOpenMenu(t *testing.T) {
	test.NewTempApp(t)
	var n *Nav
	loading := []Entry{{Title: "Loading…", Disabled: true}}
	n = New("Plant", view("plant"), WithChildren(func() []Entry { return loading }))
	n.Push("Hall 1", view("hall 1"))
	w := shown(t, n, 0)
	first := topPopUp(n, w)

	// The fetch finishes while the menu is open.
	n.Root().SetChildren(hallEntries(&n, "Hall 1", "Hall 2"))

	second := topPopUp(n, w)
	if second == nil || second == first {
		t.Fatal("open menu was not redrawn")
	}
	if first.Visible() {
		t.Fatal("old menu still visible")
	}
	if second.Position() != first.Position() {
		t.Fatalf("redrawn menu moved from %v to %v", first.Position(), second.Position())
	}
	labels, _ := menuLabels(n.childMenu(0))
	if !slices.Equal(labels, []string{"Hall 1", "Hall 2"}) {
		t.Fatalf("labels = %q", labels)
	}

	// The next opening asks the list again.
	n.closeMenu()
	test.Tap(chevronOf(t, n, 0))
	labels, _ = menuLabels(n.childMenu(0))
	if !slices.Equal(labels, []string{"Loading…"}) {
		t.Fatalf("labels after reopening = %q", labels)
	}
}

func TestSetChildrenWhileClosedOpensNothing(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	w := test.NewWindow(n)
	defer w.Close()

	n.Root().SetChildren([]Entry{{Title: "x"}})
	if topPopUp(n, w) != nil {
		t.Fatal("SetChildren opened a menu")
	}
	if got := n.Root().Children(); len(got) != 1 || got[0].Title != "x" {
		t.Fatalf("children = %v", got)
	}
}

func TestSetChildrenEmptyClosesOpenMenu(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Root().SetChildren([]Entry{{Title: "x"}})
	n.Push("x", view("x"))
	w := shown(t, n, 0)

	n.Root().SetChildren([]Entry{})
	if p := topPopUp(n, w); p != nil && p.Visible() {
		t.Fatal("menu still open without entries")
	}
	if n.menu != nil {
		t.Fatal("menu state not cleared")
	}
	chevronOf(t, n, 0) // an empty, non-nil list keeps the chevron
}

func TestRefreshChildrenAsksListAgain(t *testing.T) {
	test.NewTempApp(t)
	var n *Nav
	halls := []string{"Hall 1"}
	n = New("Plant", view("plant"), WithChildren(func() []Entry { return hallEntries(&n, halls...) }))
	n.Push("Hall 1", view("hall 1"))
	w := shown(t, n, 0)
	first := topPopUp(n, w)

	halls = append(halls, "Hall 2")
	n.Root().RefreshChildren()

	if topPopUp(n, w) == first {
		t.Fatal("menu not redrawn")
	}
	labels, _ := menuLabels(n.childMenu(0))
	if !slices.Equal(labels, []string{"Hall 1", "Hall 2"}) {
		t.Fatalf("labels = %q", labels)
	}
}

func TestOtherLayersMenuIsNotTouched(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1")
	hall := n.Push("Hall 1", view("hall 1"))
	w := shown(t, n, 0)
	open := topPopUp(n, w)

	hall.SetChildren([]Entry{{Title: "Machine A"}})
	if topPopUp(n, w) != open || !open.Visible() {
		t.Fatal("updating another layer redrew the open menu")
	}
}

func TestDismissClearsMenuState(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1")
	n.Push("Hall 1", view("hall 1"))
	w := shown(t, n, 0)

	topPopUp(n, w).Dismiss()
	if n.menu != nil {
		t.Fatal("menu state survives a dismiss")
	}
}

func TestDroppingMenuLayerClosesMenu(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1")
	hall := n.Push("Hall 1", view("hall 1"))
	hall.SetChildren([]Entry{{Title: "Machine A"}})
	n.Push("Machine A", view("m"))
	w := shown(t, n, 1)
	open := topPopUp(n, w)

	n.PopTo(0)
	if open.Visible() || n.menu != nil {
		t.Fatal("menu of a dropped layer stays open")
	}
}

func TestStaleMenuEntryIsIgnored(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1")
	hall := n.Push("Hall 1", view("hall 1"))
	hall.SetChildren([]Entry{{Title: "Machine A", Open: func() { n.Push("Machine A", view("m")) }}})
	m := n.childMenu(1)

	n.PopTo(0)
	n.Push("Hall 2", view("hall 2"))
	m.Items[0].Action()
	assertTitles(t, n, "Plant", "Hall 2")
}

func TestSetTitleRenamesStepAndCheck(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1", "Hall 2")
	hall := n.Push("…", view("hall"))
	w := shown(t, n, 0)

	hall.SetTitle("Hall 2")
	assertTitles(t, n, "Plant", "Hall 2")
	if n.bar.steps[1].crumb.Text != "Hall 2" {
		t.Fatal("bar not renamed")
	}
	_, checked := menuLabels(n.childMenu(0))
	if !slices.Equal(checked, []bool{false, true}) {
		t.Fatalf("checked = %v after rename", checked)
	}
	if topPopUp(n, w) == nil {
		t.Fatal("menu closed by a rename")
	}
	hall.SetTitle("Hall 2") // unchanged: no-op
}

func TestDroppedLayerHandleIsInert(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	b := n.Push("B", view("b"))
	n.PopTo(0)

	b.SetTitle("renamed")
	b.SetChildren([]Entry{{Title: "x"}})
	b.RefreshChildren()
	assertTitles(t, n, "A")
	if b.Title() != "renamed" {
		t.Fatal("handle lost its own title")
	}
}

func TestLayerAccessors(t *testing.T) {
	test.NewTempApp(t)
	root := view("a")
	n := New("A", root)
	b := n.Push("B", view("b"))

	if n.Root().View() != root || n.Layer(0) != n.Root() || n.Layer(1) != b || n.Top() != b {
		t.Fatal("accessors disagree")
	}
	if n.Layer(-1) != nil || n.Layer(2) != nil {
		t.Fatal("out of range layer is not nil")
	}
}

func TestNoMenuWithoutEntries(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"), WithChildren(func() []Entry { return nil }))
	if n.childMenu(-1) != nil || n.childMenu(5) != nil {
		t.Fatal("out of range index produced a menu")
	}
	n.Push("B", view("b"))
	w := shown(t, n, 0)
	if topPopUp(n, w) != nil {
		t.Fatal("tap on an empty chevron opened a menu")
	}
	n.showChildren(7, w.Content()) // out of range: nothing happens
}
