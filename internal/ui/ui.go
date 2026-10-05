// Package ui hosts the embedded web front-end: in a WebView2 window on
// Windows, or over a local HTTP server for development.
package ui

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed web
var webFS embed.FS

// Assets returns the web folder.
func Assets() fs.FS {
	sub, _ := fs.Sub(webFS, "web")
	return sub
}

// InlinePage returns index.html with the stylesheet and script inlined, so
// the page can be loaded from a string without any local web server.
func InlinePage() string {
	read := func(name string) string {
		b, _ := fs.ReadFile(Assets(), name)
		return string(b)
	}
	page := read("index.html")
	page = strings.Replace(page, `<link rel="stylesheet" href="style.css">`, "<style>\n"+read("style.css")+"\n</style>", 1)
	// Escape "</" so the script cannot close its own tag early.
	js := strings.ReplaceAll(read("app.js"), "</", `<\/`)
	page = strings.Replace(page, `<script src="app.js"></script>`, "<script>\n"+js+"\n</script>", 1)
	return page
}
