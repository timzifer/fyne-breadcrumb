package breadcrumb

import (
	"encoding/json"
	"slices"
	"testing"

	"fyne.io/fyne/v2/test"
)

// orders opens "order:<n>" and "pos:<n>" keys, unless they are in missing.
func orders(n *Nav, missing ...string) func(string) bool {
	return func(key string) bool {
		if slices.Contains(missing, key) {
			return false
		}
		n.Push("Title of "+key, view(key), WithKey(key))
		return true
	}
}

func TestRouteListsKeysAboveRoot(t *testing.T) {
	test.NewTempApp(t)
	n := New("Orders", view("list"), WithKey("ignored"))
	if r := n.Route(); len(r) != 0 {
		t.Fatalf("route of the root alone = %q", r)
	}
	l := n.Push("Order 1", view("o"), WithKey("order:1"))
	n.Push("Position 3", view("p"), WithKey("pos:3"))
	n.Push("Note", view("n")) // no key: route ends here
	n.Push("Deeper", view("d"), WithKey("never"))

	if r := n.Route(); !slices.Equal(r, Route{"order:1", "pos:3"}) {
		t.Fatalf("route = %q", r)
	}
	if l.Key() != "order:1" {
		t.Fatalf("key = %q", l.Key())
	}

	data, err := json.Marshal(n.Route())
	if err != nil || string(data) != `["order:1","pos:3"]` {
		t.Fatalf("json = %s, %v", data, err)
	}
}

func TestRestoreRebuildsStack(t *testing.T) {
	test.NewTempApp(t)
	n := New("Orders", view("list"))
	n.Push("Something else", view("x"))
	changes := 0
	n.OnChanged = func() { changes++ }

	if !n.Restore(Route{"order:1", "pos:3"}, orders(n)) {
		t.Fatal("restore reported failure")
	}
	assertTitles(t, n, "Orders", "Title of order:1", "Title of pos:3")
	if changes != 1 {
		t.Fatalf("OnChanged called %d times, want once for the whole restore", changes)
	}
	if n.CanBack() {
		t.Fatal("restore left history behind")
	}
	if !slices.Equal(n.Route(), Route{"order:1", "pos:3"}) {
		t.Fatal("route does not round-trip")
	}
}

func TestRestoreStopsAtMissingLayer(t *testing.T) {
	test.NewTempApp(t)
	n := New("Orders", view("list"))
	if n.Restore(Route{"order:1", "order:2", "pos:9"}, orders(n, "order:2")) {
		t.Fatal("restore of a missing layer reported success")
	}
	assertTitles(t, n, "Orders", "Title of order:1")
}

func TestRestoreStopsWhenOpenPushesNothing(t *testing.T) {
	test.NewTempApp(t)
	n := New("Orders", view("list"))
	if n.Restore(Route{"a"}, func(string) bool { return true }) {
		t.Fatal("open that pushed nothing counted as restored")
	}
	assertTitles(t, n, "Orders")
}

func TestRestoreEmptyRouteGoesToRoot(t *testing.T) {
	test.NewTempApp(t)
	n := New("Orders", view("list"))
	n.Push("x", view("x"))
	if !n.Restore(nil, orders(n)) {
		t.Fatal("empty route failed")
	}
	assertTitles(t, n, "Orders")
}
