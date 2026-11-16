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
// A layer can offer the layers that may stand below it. The step of the one
// open above it then carries a down chevron, "Plant › Hall 2 ⌄ › Press A":
// tapping it lists the siblings of Hall 2, Hall 2 itself checked, and
// picking one replaces Hall 2 — and everything above it — with that sibling.
// The separator › stays a plain separator; the root has no siblings.
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
// # History
//
// Every navigation is recorded the way a browser does it. [Nav.Back] returns
// to the stack before the last navigation — after a jump to the root it
// brings the dropped layers back with their views as they were — and
// [Nav.Forward] undoes that. [Nav.SetHistoryButtons] puts back and forward
// buttons in front of the steps, [Nav.BindShortcuts] binds Alt+Left,
// Alt+Right and Alt+Up (one layer up).
//
// # Stale layers
//
// A covered layer may go stale, for example when the user switches areas
// further up. [WithOnReturn] calls a function whenever its layer becomes the
// top layer again, and [Nav.OnChanged] reports every navigation.
//
// # Saving the stack
//
// Views cannot be saved, but what they show can be named. Layers pushed with
// [WithKey] make up the [Nav.Route], a plain string slice for preferences or
// JSON; [Nav.Restore] hands each key back to the application to push the
// layer again after a restart:
//
//	nav.OnChanged = func() { prefs.SetStringList("route", nav.Route()) }
//	nav.Restore(prefs.StringList("route"), openByKey)
//
// # Long titles
//
// A title wider than [DefaultMaxTitleWidth] is shortened in the middle, so
// that start and end — usually a name and a number — stay readable; see
// [Nav.SetMaxTitleWidth].
//
// # Threading
//
// Nav is a Fyne widget. Like every widget it must only be used from the Fyne
// UI goroutine; wrap calls from other goroutines in [fyne.Do].
//
// [Fyne]: https://fyne.io
package breadcrumb
