package services

import (
	"strings"
	"testing"
)

func TestSplitMessage(t *testing.T) {
	if got := splitMessage("salom", 10); len(got) != 1 || got[0] != "salom" {
		t.Fatalf("short text split: %q", got)
	}

	line := strings.Repeat("ш", 30) + "\n" // multi-byte runes
	text := strings.Repeat(line, 10)       // 310 runes
	chunks := splitMessage(text, 100)
	if strings.Join(chunks, "") != text {
		t.Fatal("chunks do not reassemble the text")
	}
	for _, c := range chunks {
		if n := len([]rune(c)); n > 100 {
			t.Errorf("chunk has %d runes", n)
		}
		if !strings.HasSuffix(c, "\n") {
			t.Errorf("chunk not cut on a line boundary: %q", c)
		}
	}

	long := strings.Repeat("a", 250)
	if chunks := splitMessage(long, 100); len(chunks) != 3 || strings.Join(chunks, "") != long {
		t.Errorf("long line split into %d chunks", len(chunks))
	}
}
