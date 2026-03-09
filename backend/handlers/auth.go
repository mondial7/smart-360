package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var googleOAuthConfig *oauth2.Config

func InitOAuthConfig() {
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/api/auth/callback"
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	googleOAuthConfig = &oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func GetGoogleAuthURL(c *gin.Context) {
	if googleOAuthConfig == nil {
		InitOAuthConfig()
	}
	url := googleOAuthConfig.AuthCodeURL("state")
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func GoogleCallback(c *gin.Context) {
	if googleOAuthConfig == nil {
		InitOAuthConfig()
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not provided"})
		return
	}

	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse user info"})
		return
	}

	db := database.GetDB()
	var user models.User

	err = db.Where("email = ?", googleUser.Email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Check if this is the first user (auto-assign admin)
		var count int64
		db.Model(&models.User{}).Count(&count)

		role := models.RoleMember
		if count == 0 {
			role = models.RoleAdmin
		}

		user = models.User{
			Email:    googleUser.Email,
			Name:     googleUser.Name,
			PhotoURL: googleUser.Picture,
			Role:     role,
		}
		db.Create(&user)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	db.Save(&user)

	// Generate JWT
	jwtToken, err := generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Redirect to frontend with token
	frontendURL := os.Getenv("FRONTEND_URL")
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, jwtToken))
}

func generateJWT(user models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"email":  user.Email,
		"role":   user.Role,
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	return token.SignedString([]byte(secret))
}

func GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DevLogin - bypass Google OAuth for development
func DevLogin(c *gin.Context) {
	db := database.GetDB()
	var user models.User

	// Check if specific email requested, otherwise default to admin
	email := c.Query("email")
	if email == "" {
		email = "dev@example.com"
	}

	// Look for existing user
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		// If dev admin doesn't exist, create it
		if email == "dev@example.com" {
			user = models.User{
				Email:    "dev@example.com",
				Name:     "Dev Admin",
				PhotoURL: "",
				Role:     models.RoleAdmin,
			}
			db.Create(&user)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found. Please seed data first."})
			return
		}
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	db.Save(&user)

	// Generate JWT
	jwtToken, err := generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Redirect to frontend with token
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, jwtToken))
}
