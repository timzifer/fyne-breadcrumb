package breadcrumb

import (
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const ellipsis = "…"

// shorten cuts title in the middle until its text is at most maxWidth wide.
// Start and end usually carry the distinguishing parts of a title (a name and
// a number), so both stay. maxWidth 0 means no limit.
func shorten(title string, maxWidth float32) string {
	if maxWidth <= 0 || textWidth(title) <= maxWidth {
		return title
	}
	runes := []rune(title)
	cut := func(keep int) string {
		head := (keep + 1) / 2
		return string(runes[:head]) + ellipsis + string(runes[len(runes)-(keep-head):])
	}
	// The largest number of kept runes that still fits; width grows with it.
	keep := sort.Search(len(runes), func(k int) bool {
		return textWidth(cut(k+1)) > maxWidth
	})
	return cut(keep)
}

// textWidth measures like a button label: bold, at the theme's text size,
// so a shortened title never ends up wider than the limit.
func textWidth(s string) float32 {
	return fyne.MeasureText(s, theme.TextSize(), fyne.TextStyle{Bold: true}).Width
}
