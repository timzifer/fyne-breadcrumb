package breadcrumb

import (
	"image/color"
	"testing"
)

// Text on a dark segment is white, on a light one black.
func TestContrastOn(t *testing.T) {
	if contrastOn(color.RGBA{R: 0, G: 60, B: 0, A: 255}) != color.White {
		t.Fatal("dunkler Grund: Schrift muss weiss sein")
	}
	if contrastOn(color.RGBA{R: 250, G: 240, B: 40, A: 255}) != color.Black {
		t.Fatal("heller Grund: Schrift muss schwarz sein")
	}
}
