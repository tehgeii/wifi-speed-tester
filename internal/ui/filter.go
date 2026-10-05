package ui

import (
	"fmt"
	"unicode/utf16"
)

// dialogFilter builds the lpstrFilter value for the Windows Save dialog:
// pairs of "label\0pattern\0", ended by an extra \0. It must keep the
// embedded NULs, so it cannot use windows.StringToUTF16 (which panics on
// NUL and closed the app on every export).
func dialogFilter(name, ext string) []uint16 {
	s := fmt.Sprintf("%s (*.%s)\x00*.%s\x00All files (*.*)\x00*.*\x00\x00", name, ext, ext)
	return utf16.Encode([]rune(s))
}
