package handlers

import "net/http"

// AuditLogs renders the admin audit trail, optionally filtered by action prefix.
func (h *Handlers) AuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logs, err := h.Repos.Audit.FindAll(ctx, 200)
	if err != nil {
		serverError(w, err)
		return
	}

	action := r.URL.Query().Get("action")
	if action != "" {
		filtered := logs[:0]
		for _, l := range logs {
			if hasPrefix(string(l.Action), action) {
				filtered = append(filtered, l)
			}
		}
		logs = filtered
	}

	data := map[string]any{"Logs": logs, "Filter": action}
	h.View.Page(w, http.StatusOK, h.page(r, "Audit log", "audit", "audit_content", data))
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
