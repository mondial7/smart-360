package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOAuthState_isUnpredictable(t *testing.T) {
	a, err := generateOAuthState()
	require.NoError(t, err)
	b, err := generateOAuthState()
	require.NoError(t, err)
	assert.NotEqual(t, a, b)
	assert.GreaterOrEqual(t, len(a), 32, "state should have meaningful entropy")
}

func TestGoogleCallback_rejectsMissingState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/auth/callback?code=any", nil)

	GoogleCallback(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid OAuth state")
}

func TestGoogleCallback_rejectsMismatchedState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/auth/callback?code=any&state=attacker", nil)
	c.Request.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "victim"})

	GoogleCallback(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid OAuth state")
}

func TestGoogleCallback_rejectsEmptyStateEvenWithCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/auth/callback?code=any", nil)
	c.Request.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: ""})

	GoogleCallback(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetGoogleAuthURL_setsStateCookieAndPropagatesItToURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/auth/google", nil)

	// InitOAuthConfig requires no env vars to construct (it just plugs in
	// empty client id/secret), and we never reach the OAuth exchange.
	googleOAuthConfig = nil

	GetGoogleAuthURL(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Cookie set
	setCookie := w.Header().Get("Set-Cookie")
	require.NotEmpty(t, setCookie, "expected oauth_state cookie")
	assert.Contains(t, setCookie, oauthStateCookieName+"=")
	assert.Contains(t, strings.ToLower(setCookie), "httponly")

	// Cookie value matches `state` param in returned URL
	cookieValue := extractCookieValue(setCookie, oauthStateCookieName)
	require.NotEmpty(t, cookieValue)

	// Body shape: {"url":"..."}
	assert.Contains(t, w.Body.String(), "state="+cookieValue)
}

func extractCookieValue(setCookieHeader, name string) string {
	parts := strings.Split(setCookieHeader, ";")
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) == 2 && kv[0] == name {
			return kv[1]
		}
	}
	return ""
}
