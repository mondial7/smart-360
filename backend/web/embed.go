// Package web exposes the embedded frontend bundle.
//
// During development the `dist/` directory contains only a `.gitkeep`
// placeholder, so HasIndex() returns false and main.go skips SPA mounting
// (the user runs `npm run dev` separately). For production / Homebrew
// builds, the top-level Makefile copies `frontend/dist/*` into this folder
// before `go build`, and the resulting binary serves the SPA itself.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// Assets returns a filesystem rooted at the embedded dist directory.
func Assets() (fs.FS, error) {
	return fs.Sub(embedded, "dist")
}

// HasIndex reports whether the embedded bundle includes a built SPA
// (i.e. an index.html). Used by main.go to decide whether to mount the
// static / SPA-fallback handlers.
func HasIndex() bool {
	sub, err := Assets()
	if err != nil {
		return false
	}
	_, err = fs.Stat(sub, "index.html")
	return err == nil
}
