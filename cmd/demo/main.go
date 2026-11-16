// Command demo shows fyne-breadcrumb on a small plant: halls offered just in
// time from the root's chevron, machines fetched in the background, and a
// title that only arrives after its layer was pushed.
package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	breadcrumb "github.com/timzifer/fyne-breadcrumb"
)

var halls = []string{"Hall 1", "Hall 2", "Hall 3"}

func main() {
	a := app.New()
	w := a.NewWindow("fyne-breadcrumb demo")

	var nav *breadcrumb.Nav
	openHall := func(name string) { pushHall(nav, name) }

	// The halls are known locally: the list is simply asked each time the
	// root's chevron opens.
	nav = breadcrumb.New("Plant", picker("Pick a hall", halls, openHall),
		breadcrumb.WithChildren(func() []breadcrumb.Entry {
			entries := make([]breadcrumb.Entry, len(halls))
			for i, h := range halls {
				entries[i] = breadcrumb.Entry{Title: h, Open: func() { openHall(h) }}
			}
			return entries
		}))

	w.SetContent(nav)
	w.Resize(fyne.NewSize(640, 420))
	w.ShowAndRun()
}

// pushHall opens a hall whose machines come from a slow "server": the menu
// shows a placeholder, and the real entries replace it when they arrive —
// also while the menu is open.
func pushHall(nav *breadcrumb.Nav, hall string) {
	status := widget.NewLabel("Loading machines…")
	box := container.NewVBox(status)
	layer := nav.Push(hall, container.NewPadded(box))
	layer.SetChildren([]breadcrumb.Entry{{Title: "Loading…", Disabled: true}})

	go func() {
		machines := fetchMachines(hall)
		fyne.Do(func() {
			entries := make([]breadcrumb.Entry, len(machines))
			for i, m := range machines {
				entries[i] = breadcrumb.Entry{Title: m, Open: func() { pushMachine(nav, m) }}
			}
			layer.SetChildren(entries)
			status.SetText(fmt.Sprintf("%d machines", len(machines)))
			box.Add(picker("", machines, func(m string) { pushMachine(nav, m) }))
		})
	}()
}

// pushMachine pushes a layer whose title is only known once loaded.
func pushMachine(nav *breadcrumb.Nav, machine string) {
	layer := nav.Push("…", container.NewPadded(widget.NewLabel("Details of "+machine)))
	go func() {
		time.Sleep(400 * time.Millisecond)
		fyne.Do(func() { layer.SetTitle(machine + " (running)") })
	}()
}

func fetchMachines(hall string) []string {
	time.Sleep(1500 * time.Millisecond)
	return []string{hall + " / Press A", hall + " / Press B", hall + " / Lathe"}
}

func picker(heading string, items []string, open func(string)) fyne.CanvasObject {
	box := container.NewVBox()
	if heading != "" {
		box.Add(widget.NewLabelWithStyle(heading, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	}
	for _, it := range items {
		box.Add(widget.NewButton(it, func() { open(it) }))
	}
	return container.NewPadded(box)
}
