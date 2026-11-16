// Package breadcrumb provides a layered navigation controller for [Fyne]
// applications: views are kept as a stack, and a breadcrumb bar on top shows
// one step per layer.
//
// Opening a sub view pushes a layer onto the stack; tapping an earlier step
// in the bar jumps back to it and drops everything above. Every layer keeps
// its view object while it is on the stack, so a list the user returns to
// still has its scroll position and selection.
//
//	nav := breadcrumb.New("Orders", orderList)
//	win.SetContent(nav)
//
//	// from the list: open an order as a new layer
//	nav.Push("Order 66340", orderView)
//
//	// back to the list
//	nav.PopTo(0)
//
// A layer is a view, not a dialog. Modal dialogs opened from a layer are not
// steps in the breadcrumb.
//
// # Siblings
//
// Like the address bar of Windows Explorer, a layer can list the layers that
// may stand below it. The bar then shows a chevron after that step; tapping
// it opens a menu of those entries, with the one currently open checked.
// Picking an entry jumps back to the step and opens the entry on top of it,
// which makes the siblings of any open layer one tap away:
//
//	nav := breadcrumb.New("Plant", plantView, breadcrumb.WithChildren(func() []breadcrumb.Entry {
//		return []breadcrumb.Entry{
//			{Title: "Hall 1", Open: func() { nav.Push("Hall 1", hallView(1)) }},
//			{Title: "Hall 2", Open: func() { nav.Push("Hall 2", hallView(2)) }},
//		}
//	}))
//
// The list is requested each time the menu opens, so it is never stale.
//
// # Threading
//
// Nav is a Fyne widget. Like every widget it must only be used from the Fyne
// UI goroutine; wrap calls from other goroutines in [fyne.Do].
//
// [Fyne]: https://fyne.io
package breadcrumb
