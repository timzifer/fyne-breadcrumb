package breadcrumb

import (
	"slices"

	"fyne.io/fyne/v2"
)

// Layer is one view on the navigation stack. The handle stays valid after the
// layer was dropped: changing it then has no visible effect until Back or
// Forward brings the layer back.
type Layer struct {
	nav   *Nav
	title string
	view  fyne.CanvasObject
	key   string

	onReturn func()
	seen     bool // has been the top layer before

	// list, when set, is asked for entries each time the menu opens.
	list func() []Entry
	// entries is what the menu shows: the last answer of list, or what
	// SetChildren gave.
	entries []Entry
	// chevron is set once the layer offers children at all; the step of
	// the child open above it then carries the sibling chevron.
	chevron bool
}

// Entry is one item a layer offers for the layer above it, listed in the
// sibling menu of that layer's step; see [WithChildren].
type Entry struct {
	// Title is the menu text. The entry is checked when it equals the
	// title of the layer currently open above — the one whose step the menu
	// was opened from.
	Title string

	// Open is called after the navigation has jumped back to the layer that
	// offered the entry, dropping the sibling it replaces. It is expected to
	// push the entry's view, usually with [Nav.Push]; if it pushes nothing,
	// the offering layer stays on top.
	Open func()

	// Disabled shows the entry greyed out, for example as a "Loading…"
	// placeholder while the real entries are fetched.
	Disabled bool
}

// LayerOption configures a layer passed to [New] or [Nav.Push].
type LayerOption func(*Layer)

// WithChildren makes a layer offer the layers that may stand below it. The
// layer itself shows no menu; the step of whichever of them is open above it
// gets a down chevron that lists them as its siblings:
//
//	Plant › Hall 2 ⌄ › Press A
//
// Plant offers the halls, so Hall 2 carries the chevron. The offering stays
// with the parent because the parent is what knows — and fetches — the
// list, and a child pushed straight from the parent's view gets its chevron
// without further ado.
//
// list is called just in time, each time the menu opens, and again on
// [Layer.RefreshChildren]. It runs on the UI goroutine and should return at
// once. Entries that take a while to fetch go through [Layer.SetChildren]:
// list returns what is known (or a disabled placeholder) and starts the
// fetch, which hands its result to SetChildren; an open menu follows.
func WithChildren(list func() []Entry) LayerOption {
	return func(l *Layer) {
		l.list = list
		l.chevron = true
	}
}

// WithOnReturn calls fn whenever the layer becomes the top layer again after
// it was covered — by a breadcrumb or menu tap, Pop, Back or Forward — but
// not when it is first pushed. It is the place to reload a view that may
// have gone stale in the background, for example after the user switched
// areas further down:
//
//	nav.Push("Orders", list, breadcrumb.WithOnReturn(list.Reload))
func WithOnReturn(fn func()) LayerOption {
	return func(l *Layer) { l.onReturn = fn }
}

func (n *Nav) newLayer(title string, view fyne.CanvasObject, opts []LayerOption) *Layer {
	if view == nil {
		panic("breadcrumb: layer " + title + " has a nil view")
	}
	l := &Layer{nav: n, title: title, view: view}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Title is the layer's step in the breadcrumb bar.
func (l *Layer) Title() string { return l.title }

// Key is the layer's key from [WithKey], or "".
func (l *Layer) Key() string { return l.key }

// View is the object the layer shows.
func (l *Layer) View() fyne.CanvasObject { return l.view }

// SetTitle renames the layer's step, for example once a heading has arrived
// from a server after the layer was pushed.
func (l *Layer) SetTitle(title string) {
	if title == l.title {
		return
	}
	l.title = title
	if !l.onStack() {
		return
	}
	l.nav.bar.rebuild()
	if m := l.nav.menu; m != nil {
		l.nav.redrawMenu(m.layer) // the checked entry may have changed
	}
}

// SetChildren replaces the entries of the layer's menu. It may be called at
// any time — before the menu was ever opened, while it is open, or after a
// fetch started by the WithChildren list has finished. An open menu is
// redrawn with the new entries, or closed when there are none.
//
// Without a WithChildren list, SetChildren alone decides whether the step
// above carries a chevron: it appears with the first entries and goes away
// with nil. With a list, the next opening asks the list again.
func (l *Layer) SetChildren(entries []Entry) {
	l.entries = slices.Clone(entries)
	chevron := l.list != nil || entries != nil
	changed := chevron != l.chevron
	l.chevron = chevron
	if !l.onStack() {
		return
	}
	if changed {
		l.nav.bar.rebuild()
	}
	l.nav.redrawMenu(l)
}

// RefreshChildren asks the WithChildren list again and redraws the menu if it
// is open. Use it when the data behind the list has changed.
func (l *Layer) RefreshChildren() {
	l.ask()
	if l.onStack() {
		l.nav.redrawMenu(l)
	}
}

// Children returns the entries the menu currently shows.
func (l *Layer) Children() []Entry { return slices.Clone(l.entries) }

func (l *Layer) ask() {
	if l.list != nil {
		l.entries = slices.Clone(l.list())
	}
}

func (l *Layer) onStack() bool { return l.nav.indexOf(l) >= 0 }
