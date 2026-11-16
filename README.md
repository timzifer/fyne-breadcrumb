# fyne-breadcrumb

[![CI](https://github.com/timzifer/fyne-breadcrumb/actions/workflows/ci.yml/badge.svg)](https://github.com/timzifer/fyne-breadcrumb/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/timzifer/fyne-breadcrumb.svg)](https://pkg.go.dev/github.com/timzifer/fyne-breadcrumb)

Layered navigation for [Fyne](https://fyne.io) applications: views kept as a
stack, with a breadcrumb bar on top — and, like the address bar of Windows
Explorer, a chevron after a step that lists its siblings. With browser-like
back and forward, keyboard shortcuts, hooks for stale layers and a saveable
route.

```
Plant ›  Hall 2 ›  Press A
      ▾
      ┌──────────┐
      │   Hall 1 │
      │ ✓ Hall 2 │
      │   Hall 3 │
      └──────────┘
```

```
go get github.com/timzifer/fyne-breadcrumb
```

## Usage

```go
import breadcrumb "github.com/timzifer/fyne-breadcrumb"

nav := breadcrumb.New("Orders", orderList)
win.SetContent(nav)

nav.Push("Order 66340", orderView) // open a sub view as a new layer
nav.PopTo(0)                        // back to the list, dropping everything above
```

Opening a sub view pushes a layer; tapping an earlier step jumps back to it.
Every layer keeps its view object while it is on the stack, so a list you
return to still has its scroll position and selection. A layer is a view, not
a dialog: modal dialogs opened from a layer are not steps.

| Call | Effect |
|---|---|
| `New(title, root, opts...)` | starts with a root layer |
| `Push(title, view, opts...)` | puts a layer on top and returns its `*Layer` |
| `Pop()` / `PopTo(i)` | drops the top layer / everything above layer `i`; invalid indices are ignored |
| `Depth()`, `Titles()` | number of layers, steps from root to top |
| `Root()`, `Top()`, `Layer(i)` | layer handles |
| `Layer.SetTitle(t)` | renames a step, e.g. once a heading arrived from a server |
| `Back()` / `Forward()` | browser-like history, see below |
| `Route()` / `Restore(route, open)` | save and rebuild the stack |
| `OnChanged` | called after every navigation |

## Siblings

A layer that knows which layers may stand below it gets a chevron. Tapping it
opens a menu of those entries with the one currently open checked; picking an
entry jumps back to that step and opens the entry on top. The siblings of any
open layer are therefore one tap away.

There are two ways to supply the entries, and they combine:

**Just in time** — a list function asked each time the menu opens:

```go
nav := breadcrumb.New("Plant", plantView, breadcrumb.WithChildren(func() []breadcrumb.Entry {
	return []breadcrumb.Entry{
		{Title: "Hall 1", Open: func() { nav.Push("Hall 1", hallView(1)) }},
		{Title: "Hall 2", Open: func() { nav.Push("Hall 2", hallView(2)) }},
	}
}))
```

**Later** — `Layer.SetChildren` at any time, for example when a fetch
finishes. An open menu is redrawn in place with the new entries:

```go
hall := nav.Push("Hall 2", hallView(2))
hall.SetChildren([]breadcrumb.Entry{{Title: "Loading…", Disabled: true}})

go func() {
	machines := fetchMachines(2)
	fyne.Do(func() { hall.SetChildren(machineEntries(machines)) })
}()
```

A list function that takes a while returns what it knows (or a disabled
placeholder), starts the fetch and hands the result to `SetChildren`.
`Layer.RefreshChildren` asks the list function again when the data behind it
has changed.

## History and keyboard

Every navigation is recorded as in a browser. `Back` returns to the stack
before the last navigation — after a jump to the root it brings the dropped
layers back, with their views as they were — and `Forward` undoes that. A new
navigation after `Back` drops the forward states. Picking a sibling from a
menu counts as one navigation. The history keeps 50 states by default
(`SetHistoryLimit`).

```go
nav.SetHistoryButtons(true)    // ← → in front of the steps
nav.BindShortcuts(win.Canvas()) // Alt+Left, Alt+Right, Alt+Up (one layer up)
```

Steps and chevrons are ordinary buttons: Tab reaches them, Space activates
them, and an open menu takes the arrow keys, Enter and Escape. Fyne gives a
focused text entry the first say on shortcuts, so the Alt shortcuts work while
nothing or a button has the focus; bind `Back`, `Forward` and `Pop` to main
menu items to have them while typing.

## Stale layers

A covered layer may go stale, for example after the user switched areas
further up. `WithOnReturn` runs whenever its layer becomes the top layer again
— not on the first push, and not for a layer only passed through on the way to
a picked sibling:

```go
nav.Push("Orders", list, breadcrumb.WithOnReturn(list.Reload))
```

## Saving the stack

Views cannot be saved, but what they show can be named. Layers pushed with
`WithKey` make up the `Route`, a plain `[]string`; `Restore` hands each key
back to the application, which pushes the layer again. It stops at the first
key that no longer opens and starts a fresh history.

```go
nav.OnChanged = func() { prefs.SetStringList("route", nav.Route()) }

nav.Restore(prefs.StringList("route"), func(key string) bool {
	order, ok := lookup(key)
	if ok {
		nav.Push(order.Title, orderView(order), breadcrumb.WithKey(key))
	}
	return ok
})
```

## Long titles

A title wider than 240 units is shortened in the middle —
`Order 66340 – Cus…ss Line B` — so its start and end stay readable; the full
title stays in `Layer.Title` and in sibling menus. `SetMaxTitleWidth` changes
the limit, 0 turns it off. Steps that do not fit in one row wrap onto the next.

## Threading

`Nav` is a Fyne widget and, like every widget, must only be used from the UI
goroutine. Wrap calls from other goroutines in `fyne.Do`.

## Demo

```
go run ./cmd/demo
```

## Compatibility

Go 1.22 and Fyne v2.6 or newer. CI builds against both floors and against the
latest Fyne release.

## License

[MIT](LICENSE)
