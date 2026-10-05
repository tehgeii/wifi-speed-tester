//go:build !windows

package main

import (
	"fmt"
	"os"
)

func prepareConsole() {}

func runGUI(string) {
	fmt.Fprintln(os.Stderr, "The desktop window is only available on Windows. Use --cli, or --serve 8080 to open the UI in a browser.")
	os.Exit(2)
}
