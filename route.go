package breadcrumb

// Route is the saveable form of a navigation stack: the keys of the layers
// above the root, bottom first. It is a plain string slice, so it goes into
// JSON, fyne.Preferences.SetStringList or a URL path as it is.
type Route []string

// WithKey names a layer for [Nav.Route]. A key identifies what the layer
// shows — an order number, a hall — well enough for the application to open
// it again after a restart.
func WithKey(key string) LayerOption {
	return func(l *Layer) { l.key = key }
}

// Route returns the keys of the layers above the root. It ends before the
// first layer without a key, since nothing above that could be restored.
func (n *Nav) Route() Route {
	var r Route
	for _, l := range n.stack[1:] {
		if l.key == "" {
			break
		}
		r = append(r, l.key)
	}
	return r
}

// Restore rebuilds a stack saved with [Nav.Route]. It jumps to the root and
// calls open for each key in turn; open is expected to push exactly that one
// layer, for example by looking the key up and calling [Nav.Push] with
// [WithKey]. When open returns false or does not push one layer — the order
// no longer exists — Restore stops there and reports false, keeping the
// layers restored so far.
//
// The whole restore is one navigation, and it starts a fresh history: Back
// does not walk through the restore step by step.
func (n *Nav) Restore(route Route, open func(key string) bool) bool {
	ok := true
	n.batch(func() {
		n.PopTo(0)
		for _, key := range route {
			depth := n.Depth()
			if !open(key) || n.Depth() != depth+1 {
				ok = false
				return
			}
		}
	})
	n.ClearHistory()
	return ok
}
