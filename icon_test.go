package breadcrumb

import (
	"testing"

	"fyne.io/fyne/v2/theme"
)

// Ein Icon macht den Schritt breiter, und zwar um genau seine Breite plus Abstand.
func TestSegmentIconWidensStep(t *testing.T) {
	plain := newSegment("Auftrag 65268")
	w0 := plain.MinSize().Width

	lead := newSegment("Auftrag 65268")
	lead.icon, lead.iconAt = theme.InfoIcon(), IconLeading
	if got := lead.MinSize().Width - w0; got-lead.iconWidth() > 0.01 || lead.iconWidth()-got > 0.01 {
		t.Fatalf("Icon vorne: Breite +%v, erwartet +%v", got, lead.iconWidth())
	}
	if lead.iconLead() == 0 || lead.iconTrail() != 0 {
		t.Fatalf("vorne: lead=%v trail=%v", lead.iconLead(), lead.iconTrail())
	}

	trail := newSegment("Auftrag 65268")
	trail.icon, trail.iconAt = theme.InfoIcon(), IconTrailing
	if trail.iconTrail() == 0 || trail.iconLead() != 0 {
		t.Fatalf("hinten: lead=%v trail=%v", trail.iconLead(), trail.iconTrail())
	}
	if trail.MinSize().Width != lead.MinSize().Width {
		t.Fatalf("vorne und hinten gleich breit: %v vs %v", trail.MinSize().Width, lead.MinSize().Width)
	}
}
