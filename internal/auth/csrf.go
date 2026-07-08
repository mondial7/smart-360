package auth

import "net/http"

// CSRF protection uses a signed token derived from the session ID. Because the
// session cookie is HttpOnly, a cross-site attacker cannot read the session ID
// and therefore cannot compute a valid token. The token is embedded in forms
// (hidden field `csrf_token`) and sent by htmx via the `X-CSRF-Token` header.

const csrfHeader = "X-CSRF-Token"
const csrfField = "csrf_token"

// CSRFToken returns the token for the request's session, or "" if unauthenticated.
func (s *Service) CSRFToken(r *http.Request) string {
	id, ok := s.readSessionID(r)
	if !ok {
		return ""
	}
	return s.sign("csrf:" + id)
}

// ProtectCSRF rejects unsafe requests whose CSRF token doesn't match the
// session. Safe methods (GET/HEAD/OPTIONS) pass through untouched.
func (s *Service) ProtectCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			next.ServeHTTP(w, r)
			return
		}
		expected := s.CSRFToken(r)
		if expected == "" {
			forbidden(w, "No session for CSRF validation")
			return
		}
		got := r.Header.Get(csrfHeader)
		if got == "" {
			got = r.FormValue(csrfField)
		}
		if !constantTimeEqual(got, expected) {
			forbidden(w, "Invalid CSRF token")
			return
		}
		next.ServeHTTP(w, r)
	})
}
