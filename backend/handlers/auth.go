package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const oauthStateCookieName = "oauth_state"
const oauthStateMaxAgeSeconds = 600

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

	state, err := generateOAuthState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize login"})
		return
	}

	// Bind the state to the user's browser via an HttpOnly cookie. The
	// callback compares the cookie to the state echoed back by Google;
	// without this, login-CSRF lets an attacker force a victim into the
	// attacker's account.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookieName, state, oauthStateMaxAgeSeconds, "/", "", isSecureCookie(), true)

	url := googleOAuthConfig.AuthCodeURL(state)
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

	state := c.Query("state")
	cookieState, err := c.Cookie(oauthStateCookieName)
	// Always clear the state cookie, regardless of outcome.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookieName, "", -1, "/", "", isSecureCookie(), true)
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(cookieState)) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
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
	ctx := context.Background()
	var user models.User

	// Look for existing user
	err = db.Collection("users").FindOne(ctx, bson.M{"email": googleUser.Email}).Decode(&user)
	if err != nil && err == mongo.ErrNoDocuments {
		// Check if this is the first user (auto-assign admin)
		count, countErr := db.Collection("users").CountDocuments(ctx, bson.M{})
		if countErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
			return
		}

		role := models.RoleMember
		if count == 0 {
			role = models.RoleAdmin
		}

		newUser := models.User{
			Email:     googleUser.Email,
			Name:      googleUser.Name,
			PhotoURL:  googleUser.Picture,
			Role:      role,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = db.Collection("users").InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		// Get the user back from MongoDB to get the real ObjectID
		err = db.Collection("users").FindOne(ctx, bson.M{"email": googleUser.Email}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created user"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now

	updateFields := bson.M{
		"last_login": user.LastLogin,
		"updated_at": user.UpdatedAt,
	}

	_, err = db.Collection("users").UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": updateFields})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

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

func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func isSecureCookie() bool {
	return os.Getenv("DEV_MODE") != "true"
}

func generateJWT(user models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not configured")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID.Hex(),
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
	ctx := context.Background()

	// Check if specific email requested, otherwise default to admin
	email := c.Query("email")
	if email == "" {
		email = "dev@example.com"
	}

	// Look for existing user
	var user models.User
	err := db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil && err == mongo.ErrNoDocuments {
		// If dev admin doesn't exist, create it
		if email == "dev@example.com" {
			newUser := models.User{
				Email:     "dev@example.com",
				Name:      "Dev Admin",
				PhotoURL:  "",
				Role:      models.RoleAdmin,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_, err := db.Collection("users").InsertOne(ctx, newUser)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dev user"})
				return
			}
			// Get the user back from MongoDB to get the real ObjectID
			err = db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
			if err != nil {
				log.Printf("Failed to retrieve created dev user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created user"})
				return
			}
			log.Printf("Retrieved dev user with ID: %s", user.ID.Hex())
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found. Please seed data first."})
			return
		}
	}

	// Update last login - Dev Login
	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now

	// Debug: Print user ID for dev login
	log.Printf("Dev login - attempting to update user with ID: %s", user.ID.Hex())

	updateFields := bson.M{
		"last_login": user.LastLogin,
		"updated_at": user.UpdatedAt,
	}

	result, err := db.Collection("users").UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": updateFields})
	if err != nil {
		log.Printf("Dev login update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	log.Printf("Dev login update successful: matched %d, modified %d", result.MatchedCount, result.ModifiedCount)

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
