package breadcrumb

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// Shortcuts installed by [Nav.BindShortcuts], the same as in web browsers and
// Windows Explorer.
var (
	ShortcutBack    = &desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: fyne.KeyModifierAlt}
	ShortcutForward = &desktop.CustomShortcut{KeyName: fyne.KeyRight, Modifier: fyne.KeyModifierAlt}
	ShortcutUp      = &desktop.CustomShortcut{KeyName: fyne.KeyUp, Modifier: fyne.KeyModifierAlt}
)

// BindShortcuts adds keyboard shortcuts to c: Alt+Left for [Nav.Back],
// Alt+Right for [Nav.Forward] and Alt+Up for [Nav.Pop].
//
// Steps in the bar take the focus without this: Tab reaches them, Space or
// Enter jumps to one, Down opens its siblings, and an open sibling menu takes
// the arrow keys, Enter and Escape.
//
// Fyne hands a shortcut to the focused widget first, and a focused text entry
// keeps it, so these shortcuts work while nothing or a button has the focus.
// An application that needs them while typing binds the same methods to main
// menu items, which Fyne checks before the focused widget.
func (n *Nav) BindShortcuts(c fyne.Canvas) {
	c.AddShortcut(ShortcutBack, func(fyne.Shortcut) { n.Back() })
	c.AddShortcut(ShortcutForward, func(fyne.Shortcut) { n.Forward() })
	c.AddShortcut(ShortcutUp, func(fyne.Shortcut) { n.Pop() })
}
