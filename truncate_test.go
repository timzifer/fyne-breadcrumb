package breadcrumb

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestShortenKeepsShortTitles(t *testing.T) {
	test.NewTempApp(t)
	for _, title := range []string{"", "Hall 1", "Order 66340"} {
		if got := shorten(title, 500); got != title {
			t.Fatalf("shorten(%q) = %q", title, got)
		}
	}
	long := strings.Repeat("x", 500)
	if got := shorten(long, 0); got != long {
		t.Fatal("limit 0 shortened a title")
	}
}

func TestShortenCutsInTheMiddle(t *testing.T) {
	test.NewTempApp(t)
	title := "Order 66340 – Customer Plant Hall 2 Press Line B"
	limit := textWidth("Order 66340 – Cust")
	got := shorten(title, limit)

	if textWidth(got) > limit {
		t.Fatalf("%q is %v wide, limit %v", got, textWidth(got), limit)
	}
	head, tail, ok := strings.Cut(got, ellipsis)
	if !ok || head == "" || tail == "" {
		t.Fatalf("%q is not cut in the middle", got)
	}
	if !strings.HasPrefix(title, head) || !strings.HasSuffix(title, tail) {
		t.Fatalf("%q does not keep start and end of %q", got, title)
	}
	if d := len([]rune(head)) - len([]rune(tail)); d < 0 || d > 1 {
		t.Fatalf("%q is not balanced", got)
	}
	// One more rune would not fit: the cut is as long as possible.
	runes := []rune(title)
	keep := len([]rune(head)) + len([]rune(tail)) + 1
	h := (keep + 1) / 2
	longer := string(runes[:h]) + ellipsis + string(runes[len(runes)-(keep-h):])
	if textWidth(longer) <= limit {
		t.Fatalf("%q would still fit; cut is too short", longer)
	}
}

func TestShortenTinyLimit(t *testing.T) {
	test.NewTempApp(t)
	if got := shorten("Something long", 1); got != ellipsis {
		t.Fatalf("shorten to 1 = %q, want only the ellipsis", got)
	}
}

func TestBarShowsShortenedTitle(t *testing.T) {
	test.NewTempApp(t)
	long := "Order 66340 – Customer Plant Hall 2 Press Line B with a very long name"
	n := New(long, view("a"))
	if title := n.bar.steps[0].title; !strings.Contains(title, ellipsis) {
		t.Fatalf("step title %q not shortened", title)
	}
	if n.Titles()[0] != long || n.Root().Title() != long {
		t.Fatal("full title lost")
	}

	n.SetMaxTitleWidth(0)
	if n.bar.steps[0].title != long {
		t.Fatal("shortening not turned off")
	}
	n.SetMaxTitleWidth(-5)
	if n.maxTitleWidth != 0 {
		t.Fatal("negative width not clamped")
	}
}
