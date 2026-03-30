package middleware

import (
	"context"
	"net/http"
	"os"
	"smart360/database"
	"smart360/models"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "your-secret-key-change-in-production"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		// Get user ID from JWT claims (stored as string for ObjectID)
		userIDStr, ok := claims["userId"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		var user models.User
		ctx := context.Background()
		err = database.GetDB().Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		u := user.(models.User)
		if u.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func TeamAdminOrGlobalAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		u := user.(models.User)

		// Global admin has access to everything
		if u.Role == models.RoleAdmin {
			c.Next()
			return
		}

		// Team admin needs to manage their own team
		if u.Role == models.RoleTeamAdmin {
			// Get team ID from route parameter
			teamID := c.Param("id")
			if teamID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Team ID required"})
				c.Abort()
				return
			}

			teamObjID, err := primitive.ObjectIDFromHex(teamID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
				c.Abort()
				return
			}

			// Verify user's team matches the route team
			if u.TeamID == nil || *u.TeamID != teamObjID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to manage this team"})
				c.Abort()
				return
			}

			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Team admin or global admin access required"})
		c.Abort()
	}
}
