package breadcrumb

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestBackAndForwardWalkTheHistory(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	if n.CanBack() || n.CanForward() || n.Back() || n.Forward() {
		t.Fatal("a fresh navigation has history")
	}

	b := view("b")
	n.Push("B", b)
	n.Push("C", view("c"))
	n.PopTo(0)
	assertTitles(t, n, "A")

	if !n.Back() {
		t.Fatal("Back after PopTo did nothing")
	}
	assertTitles(t, n, "A", "B", "C")
	n.Back()
	assertTitles(t, n, "A", "B")
	if n.Top().View() != b {
		t.Fatal("Back brought a different view object")
	}
	n.Back()
	assertTitles(t, n, "A")
	if n.CanBack() {
		t.Fatal("history goes past the start")
	}

	n.Forward()
	n.Forward()
	assertTitles(t, n, "A", "B", "C")
	n.Forward()
	assertTitles(t, n, "A")
	if n.CanForward() {
		t.Fatal("history goes past the end")
	}
}

func TestNewNavigationDropsForwardStates(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Push("B", view("b"))
	n.Back()
	n.Push("C", view("c"))

	if n.CanForward() {
		t.Fatal("forward states survive a new navigation")
	}
	n.Back()
	assertTitles(t, n, "A")
}

func TestSiblingPickIsOneHistoryEntry(t *testing.T) {
	test.NewTempApp(t)
	n, _ := plant(t, "Hall 1", "Hall 2")
	n.Push("Hall 1", view("hall 1"))
	n.Root().RefreshChildren()

	n.childMenu(0).Items[1].Action()
	assertTitles(t, n, "Plant", "Hall 2")
	n.Back()
	assertTitles(t, n, "Plant", "Hall 1")
}

func TestHistoryLimit(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.SetHistoryLimit(3)
	for _, s := range []string{"B", "C", "D", "E"} {
		n.Push(s, view(s))
	}
	steps := 0
	for n.Back() {
		steps++
	}
	if steps != 2 {
		t.Fatalf("went back %d times, want 2 with a limit of 3", steps)
	}
	assertTitles(t, n, "A", "B", "C")

	n.SetHistoryLimit(0)
	if n.CanBack() || n.CanForward() {
		t.Fatal("limit 0 still keeps history")
	}
}

func TestClearHistory(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Push("B", view("b"))
	n.ClearHistory()
	if n.CanBack() {
		t.Fatal("history not cleared")
	}
	assertTitles(t, n, "A", "B")
}

func TestUnchangedStackIsNotRecorded(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	n.Push("B", view("b"))
	changes := 0
	n.OnChanged = func() { changes++ }

	before := len(n.history.states)
	n.history.record(n.stack)
	if len(n.history.states) != before {
		t.Fatal("recording the same stack twice added a state")
	}
	n.batch(func() {}) // nothing happened: no navigation
	if changes != 0 {
		t.Fatal("an empty batch counted as a navigation")
	}
}

func TestHistoryButtons(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	if n.bar.lead != nil {
		t.Fatal("history buttons shown by default")
	}
	n.SetHistoryButtons(true)
	if n.bar.lead == nil || !n.bar.back.Disabled() || !n.bar.forward.Disabled() {
		t.Fatal("buttons missing or enabled without history")
	}

	n.Push("B", view("b"))
	if n.bar.back.Disabled() || !n.bar.forward.Disabled() {
		t.Fatal("back should be enabled after a push")
	}
	test.Tap(n.bar.back)
	assertTitles(t, n, "A")
	if !n.bar.back.Disabled() || n.bar.forward.Disabled() {
		t.Fatal("forward should be enabled after back")
	}
	test.Tap(n.bar.forward)
	assertTitles(t, n, "A", "B")

	r := test.TempWidgetRenderer(t, n.bar)
	if r.Objects()[0] != n.bar.lead {
		t.Fatal("history buttons are not the first object")
	}

	n.SetHistoryButtons(false)
	if n.bar.lead != nil {
		t.Fatal("buttons not hidden")
	}
}

func TestShortcuts(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	w := test.NewWindow(n)
	defer w.Close()
	n.BindShortcuts(w.Canvas())
	press := func(s fyne.Shortcut) { w.Canvas().(fyne.Shortcutable).TypedShortcut(s) }

	n.Push("B", view("b"))
	n.Push("C", view("c"))
	press(ShortcutUp)
	assertTitles(t, n, "A", "B")
	press(ShortcutBack)
	assertTitles(t, n, "A", "B", "C")
	press(ShortcutForward)
	assertTitles(t, n, "A", "B")
}

func TestStepsAreFocusable(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"), WithChildren(func() []Entry { return nil }))
	n.Push("B", view("b"))
	// Tab reaches every step and chevron, Space activates it.
	var _ fyne.Focusable = n.bar.steps[0].crumb
	var _ fyne.Focusable = n.bar.steps[0].sep.(*widget.Button)
}
