package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mondial7/smart-360/internal/models"
)

// Renderer parses all templates into one set and renders either full pages
// (wrapped in the base layout) or standalone fragments (for htmx swaps).
type Renderer struct {
	root *template.Template
}

// PageData is the envelope passed to the base layout. Content names the content
// template to inject; Data is the page-specific payload handed to it.
type PageData struct {
	Title          string
	Active         string // nav key to highlight
	User           *models.User
	CSRF           string
	Flash          string
	Content        string
	Data           any
	ShowOnboarding bool // render the first-login tour overlay
}

// NewRenderer parses templates from the given FS (expects templates/*.html and
// templates/partials/*.html) and wires the template func map.
func NewRenderer(fsys fs.FS) (*Renderer, error) {
	r := &Renderer{}
	funcs := template.FuncMap{
		"partial":   r.partial,
		"radarSVG":  RadarSVG,
		"donutSVG":  DonutSVG,
		"radarAxis": func(label string, value float64) RadarAxis { return RadarAxis{Label: label, Value: value} },
		"donutSlice": func(label string, value int, color string) DonutSlice {
			return DonutSlice{Label: label, Value: value, Color: color}
		},
		"dict":       dict,
		"list":       func(items ...any) []any { return items },
		"lines":      func(xs []string) string { return joinLines(xs) },
		"date":       formatDate,
		"datetime":   formatDateTime,
		"since":      humanizeSince,
		"title":      titleCase,
		"pct":        percent,
		"deref":      derefFloat,
		"hasPrefix":  hasPrefix,
		"add":        func(a, b int) int { return a + b },
		"statusTone": statusTone,
	}
	root, err := template.New("root").Funcs(funcs).ParseFS(fsys,
		"templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, err
	}
	r.root = root
	return r, nil
}

// Page renders a full page: it executes the named content template inside the
// base layout.
func (r *Renderer) Page(w http.ResponseWriter, status int, p PageData) {
	var buf bytes.Buffer
	if err := r.root.ExecuteTemplate(&buf, "base.html", p); err != nil {
		log.Printf("template render error (page %q): %v", p.Content, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// Fragment renders a standalone named template (an htmx partial swap).
func (r *Renderer) Fragment(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := r.root.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("template render error (fragment %q): %v", name, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// partial executes a named template to a string, for composition inside the
// base layout ({{ partial .Content .Data }}).
func (r *Renderer) partial(name string, data any) (template.HTML, error) {
	if name == "" {
		return "", nil
	}
	var buf bytes.Buffer
	if err := r.root.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil // #nosec G203 -- content templates are trusted, data is auto-escaped
}

// ---- template funcs ----

// dict builds a map from alternating key/value pairs, for passing multiple
// values into a partial: {{ template "x" (dict "A" 1 "B" 2) }}.
func dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("dict requires an even number of arguments")
	}
	m := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		m[key] = pairs[i+1]
	}
	return m, nil
}

func joinLines(xs []string) string {
	return strings.Join(xs, "\n")
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("Jan 2, 2006")
}

func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("Jan 2, 2006 15:04")
}

func humanizeSince(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func titleCase(s string) string {
	if s == "" {
		return ""
	}
	return string(s[0]-32) + s[1:]
}

func percent(part, total int) int {
	if total == 0 {
		return 0
	}
	return part * 100 / total
}

func derefFloat(f *float64) string {
	if f == nil {
		return "—"
	}
	return fmt.Sprintf("%.1f", *f)
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// statusTone maps a round status to a CSS modifier for coloured badges.
func statusTone(status string) string {
	switch status {
	case "draft":
		return "muted"
	case "active":
		return "info"
	case "closed":
		return "warn"
	case "shared":
		return "ok"
	}
	return "muted"
}
