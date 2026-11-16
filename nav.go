package breadcrumb

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Nav is the layer controller: a stack of views with a breadcrumb bar on top.
// Only the topmost layer is shown; the layers below keep their view objects
// until they are dropped from the stack.
type Nav struct {
	widget.BaseWidget

	bar     *bar
	host    *fyne.Container
	content fyne.CanvasObject
	stack   []*Layer
	menu    *openMenu
}

// openMenu is the chevron menu currently on screen.
type openMenu struct {
	layer  *Layer
	pop    *widget.PopUpMenu
	canvas fyne.Canvas
	pos    fyne.Position
}

// New starts a navigation with root as its only layer. rootTitle is the
// first step in the breadcrumb bar.
func New(rootTitle string, root fyne.CanvasObject, opts ...LayerOption) *Nav {
	n := &Nav{host: container.NewStack()}
	n.bar = newBar(n)
	n.content = container.NewBorder(n.bar, nil, nil, nil, n.host)
	n.ExtendBaseWidget(n)
	n.stack = []*Layer{n.newLayer(rootTitle, root, opts)}
	n.update()
	return n
}

// CreateRenderer is part of [fyne.Widget].
func (n *Nav) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(n.content)
}

// Push puts a new layer on top and shows it. The returned handle changes the
// layer's title or menu entries later.
func (n *Nav) Push(title string, view fyne.CanvasObject, opts ...LayerOption) *Layer {
	l := n.newLayer(title, view, opts)
	n.stack = append(n.stack, l)
	n.update()
	return l
}

// Pop drops the topmost layer. The root layer is never dropped; Pop reports
// whether a layer was removed.
func (n *Nav) Pop() bool {
	if len(n.stack) < 2 {
		return false
	}
	n.PopTo(len(n.stack) - 2)
	return true
}

// PopTo jumps to layer i, 0 being the root, and drops every layer above it.
// An index outside the stack is ignored.
func (n *Nav) PopTo(i int) {
	if i < 0 || i >= len(n.stack) || i == len(n.stack)-1 {
		return
	}
	clear(n.stack[i+1:]) // let the dropped views be collected
	n.stack = n.stack[:i+1]
	n.update()
}

// Depth is the number of layers; 1 means only the root.
func (n *Nav) Depth() int { return len(n.stack) }

// Layer returns layer i, 0 being the root, or nil outside the stack.
func (n *Nav) Layer(i int) *Layer {
	if i < 0 || i >= len(n.stack) {
		return nil
	}
	return n.stack[i]
}

// Root returns the root layer.
func (n *Nav) Root() *Layer { return n.stack[0] }

// Top returns the topmost layer.
func (n *Nav) Top() *Layer { return n.stack[len(n.stack)-1] }

// Titles returns the breadcrumb steps from the root to the top.
func (n *Nav) Titles() []string {
	titles := make([]string, len(n.stack))
	for i, l := range n.stack {
		titles[i] = l.title
	}
	return titles
}

func (n *Nav) update() {
	n.host.Objects = []fyne.CanvasObject{n.Top().view}
	n.host.Refresh()
	n.bar.rebuild()
	if n.menu != nil {
		// The checked entry follows the layer above; a dropped layer's
		// menu closes.
		n.redrawMenu(n.menu.layer)
	}
}

func (n *Nav) indexOf(l *Layer) int { return slices.Index(n.stack, l) }

// childMenu builds the menu for the chevron of layer i from its current
// entries, or nil when it has none.
func (n *Nav) childMenu(i int) *fyne.Menu {
	if i < 0 || i >= len(n.stack) || len(n.stack[i].entries) == 0 {
		return nil
	}
	l := n.stack[i]
	current := ""
	if i+1 < len(n.stack) {
		current = n.stack[i+1].title
	}
	items := make([]*fyne.MenuItem, len(l.entries))
	for k, e := range l.entries {
		item := fyne.NewMenuItem(e.Title, func() { n.openEntry(l, e) })
		item.Checked = i+1 < len(n.stack) && e.Title == current
		item.Disabled = e.Disabled
		items[k] = item
	}
	return fyne.NewMenu("", items...)
}

func (n *Nav) openEntry(l *Layer, e Entry) {
	// The layer may have been dropped while the menu was open.
	i := n.indexOf(l)
	if i < 0 {
		return
	}
	n.PopTo(i)
	if e.Open != nil {
		e.Open()
	}
}

// showChildren opens the chevron menu of layer i below anchor, asking the
// layer's list function for fresh entries first.
func (n *Nav) showChildren(i int, anchor fyne.CanvasObject) {
	l := n.Layer(i)
	if l == nil {
		return
	}
	l.ask()
	drv := fyne.CurrentApp().Driver()
	c := drv.CanvasForObject(anchor)
	if c == nil {
		return
	}
	n.closeMenu()
	pos := drv.AbsolutePositionForObject(anchor).AddXY(0, anchor.Size().Height)
	n.menu = &openMenu{layer: l, canvas: c, pos: pos}
	n.redrawMenu(l)
}

// redrawMenu replaces the open menu of l with one built from its current
// entries. Fyne cannot swap the items of a shown menu, so a new pop-up takes
// the place of the old one at the same position.
func (n *Nav) redrawMenu(l *Layer) {
	m := n.menu
	if m == nil || m.layer != l {
		return
	}
	menu := n.childMenu(n.indexOf(l))
	if menu == nil {
		n.closeMenu()
		return
	}
	old := m.pop
	m.pop = widget.NewPopUpMenu(menu, m.canvas)
	pop := m.pop
	pop.OnDismiss = func() {
		pop.Hide()
		if n.menu != nil && n.menu.pop == pop {
			n.menu = nil
		}
	}
	if old != nil {
		old.Hide()
	}
	pop.ShowAtPosition(m.pos)
}

func (n *Nav) closeMenu() {
	if n.menu == nil {
		return
	}
	if n.menu.pop != nil {
		n.menu.pop.Hide()
	}
	n.menu = nil
}
