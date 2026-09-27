package vt

import (
	"image/color"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func TestScrollRegion(t *testing.T) {
	term := newTestTerminal(t, 80, 24)
	if got := term.ScrollRegion(); got.Min.Y != 0 || got.Max.Y != 24 {
		t.Fatalf("default ScrollRegion = %v, want rows 0-24", got)
	}
	_, _ = term.Write([]byte("\x1b[5;20r"))
	if got := term.ScrollRegion(); got.Min.Y != 4 || got.Max.Y != 20 {
		t.Fatalf("ScrollRegion after CSI 5;20 r = %v, want rows 4-20", got)
	}
	se := NewSafeEmulator(80, 24)
	_, _ = se.Write([]byte("\x1b[2;10r"))
	if got := se.ScrollRegion(); got.Min.Y != 1 || got.Max.Y != 10 {
		t.Fatalf("SafeEmulator ScrollRegion = %v, want rows 1-10", got)
	}
}

func TestCursorPen(t *testing.T) {
	term := newTestTerminal(t, 80, 24)
	if pen := term.CursorPen(); !pen.IsZero() {
		t.Fatalf("default CursorPen = %+v, want zero", pen)
	}
	_, _ = term.Write([]byte("\x1b[1;31m"))
	pen := term.CursorPen()
	if pen.Attrs&uv.AttrBold == 0 {
		t.Errorf("CursorPen after CSI 1;31m lacks bold: %+v", pen)
	}
	if pen.Fg != color.Color(ansi.Red) {
		t.Errorf("CursorPen fg = %v, want ansi red", pen.Fg)
	}
	_, _ = term.Write([]byte("\x1b[0m"))
	if pen := term.CursorPen(); !pen.IsZero() {
		t.Errorf("CursorPen after CSI 0m = %+v, want zero", pen)
	}
	se := NewSafeEmulator(80, 24)
	_, _ = se.Write([]byte("\x1b[4m"))
	if pen := se.CursorPen(); pen.Underline == 0 {
		t.Errorf("SafeEmulator CursorPen after CSI 4m = %+v, want underline", pen)
	}
}
