package breadcrumb

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestOnReturnFiresOnlyWhenRevealed(t *testing.T) {
	test.NewTempApp(t)
	returned := map[string]int{}
	hook := func(name string) LayerOption {
		return WithOnReturn(func() { returned[name]++ })
	}
	n := New("A", view("a"), hook("A"))
	n.Push("B", view("b"), hook("B"))
	if returned["A"] != 0 || returned["B"] != 0 {
		t.Fatalf("hook fired on push: %v", returned)
	}

	n.Push("C", view("c"))
	n.PopTo(0)
	if returned["A"] != 1 || returned["B"] != 0 {
		t.Fatalf("after PopTo(0): %v, want only A once", returned)
	}

	n.Back() // C on top again: no hook on C
	n.Pop()  // B revealed
	if returned["B"] != 1 {
		t.Fatalf("B returned %d times, want 1", returned["B"])
	}
}

func TestOnReturnSkipsLayersPassedInABatch(t *testing.T) {
	test.NewTempApp(t)
	returned := 0
	n, _ := plant(t, "Hall 1", "Hall 2")
	hall := n.Push("Hall 1", view("hall 1"), WithOnReturn(func() { returned++ }))
	_ = hall
	n.Push("Machine", view("m"))
	n.Root().RefreshChildren()

	// Jumps back through Hall 1's parent and opens Hall 2: Plant is never
	// shown on its own, so nobody returned.
	plantReturned := 0
	n.Root().onReturn = func() { plantReturned++ }
	n.childMenu(0).Items[1].Action()
	if returned != 0 || plantReturned != 0 {
		t.Fatalf("hooks fired during a sibling pick: hall=%d plant=%d", returned, plantReturned)
	}
}

func TestOnChangedFollowsEveryNavigation(t *testing.T) {
	test.NewTempApp(t)
	n := New("A", view("a"))
	var depths []int
	n.OnChanged = func() { depths = append(depths, n.Depth()) }

	n.Push("B", view("b"))
	n.Push("C", view("c"))
	n.PopTo(0)
	n.Back()
	n.Forward()
	n.PopTo(0) // already there: no navigation
	n.Pop()    // root: no navigation

	want := []int{2, 3, 1, 3, 1}
	if len(depths) != len(want) {
		t.Fatalf("OnChanged depths = %v, want %v", depths, want)
	}
	for i := range want {
		if depths[i] != want[i] {
			t.Fatalf("OnChanged depths = %v, want %v", depths, want)
		}
	}
}
