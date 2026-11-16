package breadcrumb

import (
	"image/color"
	"image/png"
	"os"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

// lightTheme pins the default theme to its light variant, which GitHub
// shows most READMEs on.
type lightTheme struct{ fyne.Theme }

func (t lightTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(n, theme.VariantLight)
}

// TestDocsImage renders docs/bar.png for the README. It only writes with
// UPDATE_DOCS=1:
//
//	UPDATE_DOCS=1 go test -run TestDocsImage .
func TestDocsImage(t *testing.T) {
	if os.Getenv("UPDATE_DOCS") != "1" {
		t.Skip("set UPDATE_DOCS=1 to render docs/bar.png")
	}
	a := test.NewTempApp(t)
	a.Settings().SetTheme(lightTheme{theme.DefaultTheme()})

	n, _ := plant(t, "Hall 1", "Hall 2", "Hall 3")
	n.Push("Hall 2", view("")).SetChildren([]Entry{{Title: "Press A"}, {Title: "Press B"}, {Title: "Lathe"}})
	n.Push("Press A", view(""))
	n.Push("Order 66340", view(""))
	w := test.NewWindow(n)
	defer w.Close()
	w.Resize(fyne.NewSize(460, 170))
	s := n.bar.steps[1]
	n.showChildren(0, s, fyne.NewPos(s.chevronX(), s.Size().Height))

	if err := os.MkdirAll("docs", 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create("docs/bar.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
