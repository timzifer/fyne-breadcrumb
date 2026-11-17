package breadcrumb

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

// Ein Icon macht den Schritt breiter, und zwar um genau seine Breite plus Abstand.
func TestSegmentIconWidensStep(t *testing.T) {
	plain := newSegment("Auftrag 65268")
	w0 := plain.MinSize().Width
	h := plain.MinSize().Height

	lead := newSegment("Auftrag 65268")
	lead.icon, lead.iconAt = theme.InfoIcon(), IconLeading
	if got := lead.MinSize().Width - w0; got-lead.iconWidth(h) > 0.01 || lead.iconWidth(h)-got > 0.01 {
		t.Fatalf("Icon vorne: Breite +%v, erwartet +%v", got, lead.iconWidth(h))
	}
	if lead.iconLead(h) == 0 || lead.iconTrail(h) != 0 {
		t.Fatalf("vorne: lead=%v trail=%v", lead.iconLead(h), lead.iconTrail(h))
	}

	trail := newSegment("Auftrag 65268")
	trail.icon, trail.iconAt = theme.InfoIcon(), IconTrailing
	if trail.iconTrail(h) == 0 || trail.iconLead(h) != 0 {
		t.Fatalf("hinten: lead=%v trail=%v", trail.iconLead(h), trail.iconTrail(h))
	}
	if trail.MinSize().Width != lead.MinSize().Width {
		t.Fatalf("vorne und hinten gleich breit: %v vs %v", trail.MinSize().Width, lead.MinSize().Width)
	}
}

// Ein Icon-Objekt hat die Größe, die es selbst meldet, und steht mittig in der Zeile.
func TestSegmentIconObjectKeepsItsOwnSize(t *testing.T) {
	seg := newSegment("Bereich")
	box := canvas.NewRectangle(nil)
	box.SetMinSize(fyne.NewSize(34, 12))
	seg.iconObj, seg.iconAt = box, IconTrailing
	if got := seg.iconObjSize(); got != fyne.NewSize(34, 12) {
		t.Fatalf("Größe %v, erwartet 34x12", got)
	}
	if seg.iconTrail(24) == 0 {
		t.Fatal("Icon-Objekt belegt keinen Platz")
	}
}
