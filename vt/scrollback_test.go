package vt

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestScrollback(t *testing.T) {
	t.Run("basic push and len", func(t *testing.T) {
		sb := NewScrollback(100)
		if sb.Len() != 0 {
			t.Errorf("expected len 0, got %d", sb.Len())
		}
		if sb.MaxLines() != 100 {
			t.Errorf("expected max 100, got %d", sb.MaxLines())
		}
	})

	t.Run("scrollback in emulator", func(t *testing.T) {
		// Create a small terminal
		e := NewEmulator(10, 5)

		// Fill the screen with numbered lines and force scrolling
		for i := 0; i < 10; i++ {
			e.WriteString("\r\n") // Scroll up
		}

		// Check scrollback has captured some lines
		sbLen := e.ScrollbackLen()
		t.Logf("scrollback length after 10 newlines: %d", sbLen)

		if sbLen == 0 {
			t.Error("expected scrollback to have captured lines, got 0")
		}
	})

	t.Run("scrollback with content", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write content that will scroll
		for i := 0; i < 10; i++ {
			e.WriteString("line\r\n")
		}

		// Verify scrollback captured the scrolled content
		sb := e.Scrollback()
		if sb == nil {
			t.Fatal("scrollback is nil")
		}

		// Should have captured lines (at least 5, since screen is 5 tall and we wrote 10 lines)
		if sb.Len() < 5 {
			t.Errorf("expected at least 5 lines in scrollback, got %d", sb.Len())
		}
	})

	t.Run("scrollback max lines", func(t *testing.T) {
		sb := NewScrollback(5)

		// Push more lines than max
		for i := 0; i < 10; i++ {
			sb.Push(nil)
		}

		if sb.Len() != 5 {
			t.Errorf("expected len 5 after overflow, got %d", sb.Len())
		}
	})

	t.Run("clear scrollback", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write content that will scroll
		for i := 0; i < 10; i++ {
			e.WriteString("line\r\n")
		}

		// Verify we have scrollback
		if e.ScrollbackLen() == 0 {
			t.Error("expected scrollback before clear")
		}

		// Clear it
		e.ClearScrollback()

		if e.ScrollbackLen() != 0 {
			t.Errorf("expected empty scrollback after clear, got %d", e.ScrollbackLen())
		}
	})

	t.Run("alt screen does not have scrollback", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write some content to main screen
		for i := 0; i < 10; i++ {
			e.WriteString("line\r\n")
		}

		mainScrollbackLen := e.ScrollbackLen()
		if mainScrollbackLen == 0 {
			t.Error("expected scrollback on main screen")
		}

		// Enter alt screen
		e.WriteString("\x1b[?1049h") // DECSET alt screen

		// Scrollback should still be from main screen
		if e.ScrollbackLen() != mainScrollbackLen {
			t.Errorf("expected scrollback len %d in alt screen, got %d",
				mainScrollbackLen, e.ScrollbackLen())
		}

		// Write to alt screen - should not affect main scrollback
		for i := 0; i < 10; i++ {
			e.WriteString("alt\r\n")
		}

		// Main screen scrollback should be unchanged
		if e.ScrollbackLen() != mainScrollbackLen {
			t.Errorf("expected scrollback len %d after alt screen writes, got %d",
				mainScrollbackLen, e.ScrollbackLen())
		}
	})

	t.Run("ED 2 saves to scrollback", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write some content (not enough to scroll)
		e.WriteString("line 1\r\n")
		e.WriteString("line 2\r\n")
		e.WriteString("line 3\r\n")

		// Should have no scrollback yet (didn't scroll)
		initialLen := e.ScrollbackLen()

		// Clear screen with ED 2 (ESC[2J)
		e.WriteString("\x1b[2J")

		// Should have saved lines to scrollback
		newLen := e.ScrollbackLen()
		if newLen <= initialLen {
			t.Errorf("expected scrollback to grow after ED 2, was %d now %d", initialLen, newLen)
		}
		t.Logf("scrollback after ED 2: %d lines", newLen)
	})

	t.Run("ED 3 clears scrollback", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write content that will scroll
		for i := 0; i < 10; i++ {
			e.WriteString("line\r\n")
		}

		// Verify we have scrollback
		if e.ScrollbackLen() == 0 {
			t.Error("expected scrollback before ED 3")
		}

		// ED 3 (ESC[3J) should clear scrollback
		e.WriteString("\x1b[3J")

		if e.ScrollbackLen() != 0 {
			t.Errorf("expected empty scrollback after ED 3, got %d", e.ScrollbackLen())
		}
	})

	t.Run("ring buffer order and lines across wrap", func(t *testing.T) {
		sb := NewScrollback(5)

		makeLine := func(val rune) uv.Line {
			l := make(uv.Line, 3)
			l[0].Content = string(val)
			l[0].Width = 1
			l[1].Content = string(val)
			l[1].Width = 1
			l[2].Content = string(val)
			l[2].Width = 1
			return l
		}

		for i := 0; i < 12; i++ {
			sb.Push(makeLine(rune('A' + i)))
		}

		if sb.Len() != 5 {
			t.Fatalf("expected len 5, got %d", sb.Len())
		}

		// Oldest should be 'H' (index 7), newest should be 'L' (index 11)
		for i := 0; i < 5; i++ {
			want := string(rune('H' + i))
			line := sb.Line(i)
			if line == nil || line[0].Content != want {
				t.Errorf("line %d: expected content %s, got %v", i, want, line)
			}
			cell := sb.CellAt(0, i)
			if cell == nil || cell.Content != want {
				t.Errorf("cell %d: expected content %s, got %v", i, want, cell)
			}
		}

		all := sb.Lines()
		if len(all) != 5 {
			t.Fatalf("expected Lines() len 5, got %d", len(all))
		}
		for i := 0; i < 5; i++ {
			want := string(rune('H' + i))
			if all[i][0].Content != want {
				t.Errorf("Lines()[%d]: expected content %s, got %s", i, want, all[i][0].Content)
			}
		}

		// Test SetMaxLines preserves order
		sb.SetMaxLines(3)
		if sb.Len() != 3 {
			t.Fatalf("expected len 3 after resize, got %d", sb.Len())
		}
		for i := 0; i < 3; i++ {
			want := string(rune('J' + i))
			line := sb.Line(i)
			if line == nil || line[0].Content != want {
				t.Errorf("after resize line %d: expected content %s, got %v", i, want, line)
			}
		}
	})

	t.Run("SetMaxLines growth preserves order after wrapping", func(t *testing.T) {
		sb := NewScrollback(3)
		makeLine := func(r rune) uv.Line {
			l := make(uv.Line, 1)
			l[0].Content = string(r)
			l[0].Width = 1
			return l
		}

		// Push A, B, C, D -> ring wraps: head = 1, logical [B, C, D]
		sb.Push(makeLine('A'))
		sb.Push(makeLine('B'))
		sb.Push(makeLine('C'))
		sb.Push(makeLine('D'))

		// Increase capacity to 5
		sb.SetMaxLines(5)

		// Push E -> should append, yielding logical [B, C, D, E]
		sb.Push(makeLine('E'))

		if sb.Len() != 4 {
			t.Fatalf("expected len 4, got %d", sb.Len())
		}
		expected := []string{"B", "C", "D", "E"}
		for i, want := range expected {
			line := sb.Line(i)
			if line == nil || line[0].Content != want {
				t.Errorf("line %d: expected %s, got %v", i, want, line)
			}
		}
		all := sb.Lines()
		for i, want := range expected {
			if all[i][0].Content != want {
				t.Errorf("Lines()[%d]: expected %s, got %s", i, want, all[i][0].Content)
			}
		}
	})

	t.Run("buffer reuse on eviction", func(t *testing.T) {
		sb := NewScrollback(3)
		makeLine := func(n int, r rune) uv.Line {
			l := make(uv.Line, n)
			for i := range l {
				l[i].Content = string(r)
				l[i].Width = 1
			}
			return l
		}

		sb.Push(makeLine(10, 'A'))
		sb.Push(makeLine(10, 'B'))
		sb.Push(makeLine(10, 'C'))

		// Oldest line is A; grab its backing array pointer
		oldPtr := &sb.Line(0)[0]

		// Push 4th line with smaller length (can reuse capacity 10)
		sb.Push(makeLine(6, 'D'))

		// Newest line is D (index 2)
		newPtr := &sb.Line(2)[0]
		if oldPtr != newPtr {
			t.Errorf("expected evicted buffer to be reused (old ptr %p, new ptr %p)", oldPtr, newPtr)
		}
	})
}

func BenchmarkScrollbackPushFull(b *testing.B) {
	sb := NewScrollback(DefaultScrollbackSize)
	line := make(uv.Line, 80)
	for i := range line {
		line[i].Content = "x"
		line[i].Width = 1
	}
	// Fill scrollback to capacity
	for i := 0; i < DefaultScrollbackSize; i++ {
		sb.Push(line)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sb.Push(line)
	}
}
