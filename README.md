# fyne-breadcrumb

[![CI](https://github.com/timzifer/fyne-breadcrumb/actions/workflows/ci.yml/badge.svg)](https://github.com/timzifer/fyne-breadcrumb/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/timzifer/fyne-breadcrumb.svg)](https://pkg.go.dev/github.com/timzifer/fyne-breadcrumb)

Layered navigation for [Fyne](https://fyne.io) applications: views kept as a
stack, with a breadcrumb bar on top — and, like the address bar of Windows
Explorer, a chevron after a step that lists its siblings.

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
