package handlers

import (
	"fmt"
	"net/http"
)

// LogsPage renders the admin live-log viewer. It streams the application log
// over SSE into the browser (see logs.html), and shows the stream endpoint so
// an admin can point their own tail/consumer at it.
func (h *Handlers) LogsPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Token":     h.Auth.StreamToken(r),
		"LogFormat": h.Cfg.LogFormat,
	}
	h.View.Page(w, http.StatusOK, h.page(r, "Logs", "logs", "logs_content", data))
}

// LogsStream streams application logs as Server-Sent Events: the recent buffer
// first, then live lines. Admin only, guarded by the session-derived stream
// token (it's a GET, so the CSRF middleware doesn't cover it).
func (h *Handlers) LogsStream(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.ValidStreamToken(r, r.URL.Query().Get("t")) {
		forbidden(w)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok || h.Logs == nil {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Backfill the recent buffer, then subscribe to live lines.
	for _, line := range h.Logs.Recent() {
		writeLogSSE(w, flusher, line)
	}
	sub, cancel := h.Logs.Subscribe()
	defer cancel()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-sub:
			if !ok {
				return
			}
			writeLogSSE(w, flusher, line)
		}
	}
}

// writeLogSSE emits one log line as an SSE "line" event whose data is an
// HTML-escaped element appended to the viewer (hx-swap="beforeend").
func writeLogSSE(w http.ResponseWriter, f http.Flusher, line string) {
	fmt.Fprintf(w, "event: line\ndata: <div class=\"log-line\">%s</div>\n\n", htmlEscape(line))
	f.Flush()
}
