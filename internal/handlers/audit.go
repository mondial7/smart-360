package handlers

import "net/http"

// AuditLogs renders the admin audit trail. The unfiltered view is paginated;
// when an action filter is set, it filters over the most recent 200 entries
// (filtered views are narrow, so a cap is fine there).
func (h *Handlers) AuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	action := r.URL.Query().Get("action")

	data := map[string]any{"Filter": action}
	if action != "" {
		logs, err := h.Repos.Audit.FindAll(ctx, 200)
		if err != nil {
			serverError(w, err)
			return
		}
		filtered := logs[:0]
		for _, l := range logs {
			if hasPrefix(string(l.Action), action) {
				filtered = append(filtered, l)
			}
		}
		data["Logs"] = filtered
	} else {
		page := pageParam(r)
		rows, err := h.Repos.Audit.FindPaged(ctx, pageSize+1, (page-1)*pageSize)
		if err != nil {
			serverError(w, err)
			return
		}
		logs, hasNext := paginate(rows)
		data["Logs"] = logs
		data["Nav"] = buildPageNav(r, page, hasNext)
	}
	h.View.Page(w, http.StatusOK, h.page(r, "Audit log", "audit", "audit_content", data))
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
