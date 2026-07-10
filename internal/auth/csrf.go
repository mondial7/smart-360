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

// StreamToken returns a session-derived token for authorizing a state-changing
// SSE GET (which cannot carry the CSRF header). It is derived from a different
// label than the form CSRF token, so exposure of one (e.g. in a URL or access
// log) cannot be replayed as the other.
func (s *Service) StreamToken(r *http.Request) string {
	id, ok := s.readSessionID(r)
	if !ok {
		return ""
	}
	return s.sign("sse:" + id)
}

// ValidStreamToken reports whether token matches the request's stream token.
func (s *Service) ValidStreamToken(r *http.Request, token string) bool {
	expected := s.StreamToken(r)
	return expected != "" && constantTimeEqual(token, expected)
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
