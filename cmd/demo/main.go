// Command demo shows fyne-breadcrumb on a small plant: halls offered just in
// time by the root, machines fetched in the background, a title that only
// arrives after its layer was pushed, back and forward with buttons and
// Alt+Left/Right/Up, and the stack restored after a restart.
package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	breadcrumb "github.com/timzifer/fyne-breadcrumb"
)

const routePref = "route"

var halls = []string{"Hall 1", "Hall 2", "Hall 3"}

type demo struct {
	nav     *breadcrumb.Nav
	visits  *widget.Label
	returns int
}

func main() {
	a := app.NewWithID("com.github.timzifer.fyne-breadcrumb.demo")
	w := a.NewWindow("fyne-breadcrumb demo")
	d := &demo{visits: widget.NewLabel("")}

	// The halls are known locally: the list is simply asked each time the
	// chevron on the open hall's step opens.
	root := container.NewVBox(picker("Pick a hall", halls, d.pushHall), d.visits)
	d.nav = breadcrumb.New("Plant", container.NewPadded(root),
		breadcrumb.WithChildren(func() []breadcrumb.Entry {
			entries := make([]breadcrumb.Entry, len(halls))
			for i, h := range halls {
				entries[i] = breadcrumb.Entry{Title: h, Open: func() { d.pushHall(h) }}
			}
			return entries
		}),
		// The overview "reloads" whenever the user comes back to it.
		breadcrumb.WithOnReturn(func() {
			d.returns++
			d.visits.SetText(fmt.Sprintf("Overview reloaded %d times", d.returns))
		}))
	d.nav.SetHistoryButtons(true)
	d.nav.BindShortcuts(w.Canvas())

	// Pick up where the last run ended, and remember every navigation.
	d.nav.Restore(a.Preferences().StringList(routePref), d.open)
	d.nav.OnChanged = func() {
		a.Preferences().SetStringList(routePref, d.nav.Route())
	}

	w.SetContent(d.nav)
	w.Resize(fyne.NewSize(720, 420))
	w.ShowAndRun()
}

// open restores one layer from its saved key.
func (d *demo) open(key string) bool {
	kind, name, _ := strings.Cut(key, ":")
	switch kind {
	case "hall":
		d.pushHall(name)
	case "machine":
		d.pushMachine(name)
	default:
		return false
	}
	return true
}

// pushHall opens a hall whose machines come from a slow "server": the menu
// shows a placeholder, and the real entries replace it when they arrive —
// also while the menu is open.
func (d *demo) pushHall(hall string) {
	status := widget.NewLabel("Loading machines…")
	box := container.NewVBox(status)
	layer := d.nav.Push(hall, container.NewPadded(box), breadcrumb.WithKey("hall:"+hall))
	layer.SetChildren([]breadcrumb.Entry{{Title: "Loading…", Disabled: true}})

	go func() {
		machines := fetchMachines(hall)
		fyne.Do(func() {
			entries := make([]breadcrumb.Entry, len(machines))
			for i, m := range machines {
				entries[i] = breadcrumb.Entry{Title: m, Open: func() { d.pushMachine(m) }}
			}
			layer.SetChildren(entries)
			status.SetText(fmt.Sprintf("%d machines", len(machines)))
			box.Add(picker("", machines, d.pushMachine))
		})
	}()
}

// pushMachine pushes a layer whose title is only known once loaded.
func (d *demo) pushMachine(machine string) {
	layer := d.nav.Push("…", container.NewPadded(widget.NewLabel("Details of "+machine)),
		breadcrumb.WithKey("machine:"+machine))
	go func() {
		time.Sleep(400 * time.Millisecond)
		fyne.Do(func() { layer.SetTitle(machine) })
	}()
}

func fetchMachines(hall string) []string {
	time.Sleep(1500 * time.Millisecond)
	return []string{
		hall + " / Press A",
		hall + " / Press B",
		hall + " / Lathe with an unusually long designation for the cut",
	}
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
