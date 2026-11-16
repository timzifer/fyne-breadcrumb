package breadcrumb

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// DefaultMaxTitleWidth is the width beyond which a step's title is shortened
// in the middle, see [Nav.SetMaxTitleWidth].
const DefaultMaxTitleWidth float32 = 240

// Nav is the layer controller: a stack of views with a breadcrumb bar on top.
// Only the topmost layer is shown; the layers below keep their view objects
// until they are dropped from the stack and from the history.
type Nav struct {
	widget.BaseWidget

	// OnChanged is called after every navigation — Push, Pop, PopTo, a
	// breadcrumb or menu tap, Back, Forward, Restore — once the new top
	// layer is shown. A pick from a sibling menu counts as one navigation.
	// It suits saving the [Nav.Route] or updating a window title.
	OnChanged func()

	bar     *bar
	host    *fyne.Container
	content fyne.CanvasObject
	stack   []*Layer
	menu    *openMenu

	history *history
	quiet   int  // > 0 while a batch runs
	dirty   bool // the stack changed during the batch
	shown   *Layer

	maxTitleWidth  float32
	historyButtons bool
}

// openMenu is the sibling menu currently on screen.
type openMenu struct {
	layer  *Layer
	pop    *widget.PopUpMenu
	canvas fyne.Canvas
	pos    fyne.Position
}

// New starts a navigation with root as its only layer. rootTitle is the
// first step in the breadcrumb bar.
func New(rootTitle string, root fyne.CanvasObject, opts ...LayerOption) *Nav {
	n := &Nav{host: container.NewStack(), maxTitleWidth: DefaultMaxTitleWidth}
	n.bar = newBar(n)
	n.content = container.NewBorder(n.bar, nil, nil, nil, n.host)
	n.ExtendBaseWidget(n)
	n.stack = []*Layer{n.newLayer(rootTitle, root, opts)}
	n.history = newHistory(n.stack)
	n.commit(false)
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
	n.changed()
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
	clear(n.stack[i+1:]) // the history keeps what it still needs
	n.stack = n.stack[:i+1]
	n.changed()
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

// SetMaxTitleWidth sets the width beyond which a step's title is shortened in
// the middle ("Order 6634…ant B"), keeping its start and end readable. The
// full title stays available through [Layer.Title] and in sibling menus. 0
// turns shortening off; the default is [DefaultMaxTitleWidth].
func (n *Nav) SetMaxTitleWidth(w float32) {
	n.maxTitleWidth = max(w, 0)
	n.bar.rebuild()
}

// SetHistoryButtons shows or hides back and forward buttons at the start of
// the breadcrumb bar. They are hidden by default.
func (n *Nav) SetHistoryButtons(show bool) {
	n.historyButtons = show
	n.bar.rebuild()
}

// changed is called by every mutation of the stack.
func (n *Nav) changed() {
	if n.quiet > 0 {
		n.dirty = true
		return
	}
	n.commit(true)
}

// batch runs fn as one navigation: whatever fn does to the stack is shown,
// recorded and reported once, at the end.
func (n *Nav) batch(fn func()) {
	n.quiet++
	defer func() {
		n.quiet--
		if n.quiet == 0 && n.dirty {
			n.dirty = false
			n.commit(true)
		}
	}()
	fn()
}

// commit shows the current stack and runs what follows a navigation.
func (n *Nav) commit(record bool) {
	if record {
		n.history.record(n.stack)
	}
	n.update()
	if top := n.Top(); top != n.shown {
		n.shown = top
		returned := top.seen
		top.seen = true
		if returned && top.onReturn != nil {
			top.onReturn()
		}
	}
	if n.OnChanged != nil {
		n.OnChanged()
	}
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

// childMenu builds the sibling menu from the entries layer i offers, or nil
// when it has none.
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
	n.batch(func() {
		n.PopTo(i)
		if e.Open != nil {
			e.Open()
		}
	})
}

// showChildren opens the entries of layer i at position at of anchor — below
// the sibling chevron of the step above — asking the layer's list function
// for fresh entries first.
func (n *Nav) showChildren(i int, anchor fyne.CanvasObject, at fyne.Position) {
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
	pos := drv.AbsolutePositionForObject(anchor).Add(at)
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
