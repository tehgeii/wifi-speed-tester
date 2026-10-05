package ui

import (
	"testing"
	"unicode/utf16"
)

func TestDialogFilter(t *testing.T) {
	f := dialogFilter("Text report", "txt")
	want := "Text report (*.txt)\x00*.txt\x00All files (*.*)\x00*.*\x00\x00"
	if got := string(utf16.Decode(f)); got != want {
		t.Fatalf("filter = %q, want %q", got, want)
	}
	if f[len(f)-1] != 0 || f[len(f)-2] != 0 {
		t.Fatal("filter must end with a double NUL")
	}
}
