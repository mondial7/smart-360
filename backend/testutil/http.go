package testutil

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"smart360/models"

	"github.com/gin-gonic/gin"
)

// NewTestGinContext creates a Gin context for testing with an optional authenticated user
func NewTestGinContext(user *models.User) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user in context if provided (simulates authentication middleware)
	if user != nil {
		c.Set("user", *user)
	}

	return c, w
}

// SetJSONBody sets a JSON body on the Gin context for POST/PUT requests
func SetJSONBody(c *gin.Context, body interface{}) error {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	return nil
}

// SetParam sets a URL parameter on the Gin context
func SetParam(c *gin.Context, key, value string) {
	c.Params = append(c.Params, gin.Param{Key: key, Value: value})
}

// SetQueryParam sets a query parameter on the Gin context
func SetQueryParam(c *gin.Context, key, value string) {
	if c.Request == nil {
		c.Request = httptest.NewRequest("GET", "/?"+key+"="+value, nil)
	} else {
		q := c.Request.URL.Query()
		q.Add(key, value)
		c.Request.URL.RawQuery = q.Encode()
	}
}

// ParseJSONResponse parses the JSON response from a response recorder
func ParseJSONResponse(w *httptest.ResponseRecorder, v interface{}) error {
	return json.Unmarshal(w.Body.Bytes(), v)
}
