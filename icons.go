package breadcrumb

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Fyne's theme has a drop-down triangle but no downward chevron, so the bar
// brings its own. Themed, it takes the foreground colour like any icon.
var siblingsIcon = theme.NewThemedResource(fyne.NewStaticResource("siblings.svg", []byte(
	`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">`+
		`<path d="M5.4 8.6 12 15.2l6.6-6.6L20 10l-8 8-8-8z"/></svg>`)))

// siblingsIconIn is the sibling chevron in the given colour. It follows the step's text colour,
// so the chevron of a step that is not the top one keeps the colour of its label.
func siblingsIconIn(c color.Color) fyne.Resource {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">`+
		`<path fill="#%02x%02x%02x" d="M5.4 8.6 12 15.2l6.6-6.6L20 10l-8 8-8-8z"/></svg>`, n.R, n.G, n.B)
	return fyne.NewStaticResource("siblings.svg", []byte(svg))
}
