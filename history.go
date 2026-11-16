package breadcrumb

import "slices"

// DefaultHistoryLimit is the number of navigations [Nav.Back] can undo, see
// [Nav.SetHistoryLimit].
const DefaultHistoryLimit = 50

// history is the browser-like list of stacks the navigation went through.
// It holds the layers themselves, so a view revisited through Back or
// Forward still has its state.
type history struct {
	states [][]*Layer
	pos    int
	limit  int
}

func newHistory(stack []*Layer) *history {
	return &history{states: [][]*Layer{slices.Clone(stack)}, limit: DefaultHistoryLimit}
}

// record appends stack as the newest state, dropping any forward states.
func (h *history) record(stack []*Layer) {
	if slices.Equal(h.states[h.pos], stack) {
		return
	}
	h.states = append(h.states[:h.pos+1], slices.Clone(stack))
	h.pos++
	h.trim()
}

func (h *history) trim() {
	if extra := len(h.states) - h.limit; extra > 0 {
		clear(h.states[:extra])
		h.states = h.states[extra:]
		h.pos = max(h.pos-extra, 0)
	}
}

func (h *history) reset(stack []*Layer) {
	h.states = [][]*Layer{slices.Clone(stack)}
	h.pos = 0
}

// CanBack reports whether [Nav.Back] has a state to return to.
func (n *Nav) CanBack() bool { return n.history.pos > 0 }

// CanForward reports whether [Nav.Forward] has a state to return to.
func (n *Nav) CanForward() bool { return n.history.pos < len(n.history.states)-1 }

// Back returns to the stack before the last navigation, like a browser's back
// button: after a Push it drops the layer again, after a jump to the root it
// brings the dropped layers back, with their views as they were. It reports
// whether there was a state to return to.
func (n *Nav) Back() bool { return n.travel(-1) }

// Forward undoes a [Nav.Back]. A new navigation after Back discards the
// forward states, as in a browser.
func (n *Nav) Forward() bool { return n.travel(+1) }

func (n *Nav) travel(step int) bool {
	h := n.history
	to := h.pos + step
	if to < 0 || to >= len(h.states) {
		return false
	}
	h.pos = to
	n.stack = slices.Clone(h.states[to])
	n.commit(false)
	return true
}

// SetHistoryLimit sets how many states the history keeps, the current one
// included; the oldest go first. A limit below 1 is taken as 1, which turns
// Back and Forward off. The default is [DefaultHistoryLimit].
func (n *Nav) SetHistoryLimit(limit int) {
	n.history.limit = max(limit, 1)
	n.history.trim()
	n.bar.rebuild()
}

// ClearHistory forgets every state but the current one.
func (n *Nav) ClearHistory() {
	n.history.reset(n.stack)
	n.bar.rebuild()
}
