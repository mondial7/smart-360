// Package web embeds the HTML templates and static assets into the binary so
// the server ships as a single file with no external asset dependencies.
package web

import "embed"

// TemplatesFS holds the html/template sources under templates/.
//
//go:embed templates/*.html templates/partials/*.html
var TemplatesFS embed.FS

// StaticFS holds JS/CSS/image assets served under /static/.
//
//go:embed static
var StaticFS embed.FS
