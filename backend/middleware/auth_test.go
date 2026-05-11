package middleware

import (
	"net/http"
	"net/http/httptest"
	"smart360/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newCtx(user *models.User) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	if user != nil {
		c.Set("user", *user)
	}
	return c, w
}

func TestAdminOrTeamAdminRole(t *testing.T) {
	t.Run("admin_passes", func(t *testing.T) {
		c, w := newCtx(&models.User{Role: models.RoleAdmin})
		AdminOrTeamAdminRole()(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, c.IsAborted())
	})
	t.Run("team_admin_passes", func(t *testing.T) {
		c, w := newCtx(&models.User{Role: models.RoleTeamAdmin})
		AdminOrTeamAdminRole()(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, c.IsAborted())
	})
	t.Run("member_is_forbidden", func(t *testing.T) {
		c, w := newCtx(&models.User{Role: models.RoleMember})
		AdminOrTeamAdminRole()(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.True(t, c.IsAborted())
	})
	t.Run("unauthenticated_is_unauthorized", func(t *testing.T) {
		c, w := newCtx(nil)
		AdminOrTeamAdminRole()(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})
}
