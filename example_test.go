package breadcrumb_test

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	breadcrumb "github.com/timzifer/fyne-breadcrumb"
)

func Example() {
	a := test.NewApp()
	defer a.Quit()

	nav := breadcrumb.New("Orders", widget.NewLabel("order list"))
	w := a.NewWindow("Demo")
	w.SetContent(nav)

	nav.Push("Order 66340", widget.NewLabel("order details"))
	fmt.Println(nav.Titles())

	nav.PopTo(0)
	fmt.Println(nav.Titles())
	// Output:
	// [Orders Order 66340]
	// [Orders]
}

func ExampleWithChildren() {
	a := test.NewApp()
	defer a.Quit()

	var nav *breadcrumb.Nav
	hall := func(name string) breadcrumb.Entry {
		return breadcrumb.Entry{Title: name, Open: func() {
			nav.Push(name, widget.NewLabel(name))
		}}
	}
	// The root offers the halls; once one is open, its siblings are one tap
	// away on the chevron after "Plant".
	nav = breadcrumb.New("Plant", widget.NewLabel("plant overview"),
		breadcrumb.WithChildren(func() []breadcrumb.Entry {
			return []breadcrumb.Entry{hall("Hall 1"), hall("Hall 2")}
		}))

	var _ fyne.Widget = nav
	nav.Push("Hall 1", widget.NewLabel("Hall 1"))
	fmt.Println(nav.Titles())
	// Output: [Plant Hall 1]
}
