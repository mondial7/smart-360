package handlers

import (
	"context"
	"net/http"
	"smart360/database"
	"smart360/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetAuditLogs returns all audit logs with pagination (admin only)
func GetAuditLogs(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Only admins can view audit logs
	if currentUser.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Parse pagination parameters
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	skip := (page - 1) * limit

	// Get total count
	total, err := db.Collection("audit_logs").CountDocuments(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count audit logs"})
		return
	}

	// Fetch audit logs sorted by created_at descending
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := db.Collection("audit_logs").Find(ctx, bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}
	defer cursor.Close(ctx)

	var logs []models.AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetRoundAuditLogs returns audit logs for a specific round (admin only)
func GetRoundAuditLogs(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Only admins can view audit logs
	if currentUser.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	roundID := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Fetch audit logs for this round, sorted chronologically
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := db.Collection("audit_logs").Find(ctx, bson.M{"round_id": roundObjID}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}
	defer cursor.Close(ctx)

	var logs []models.AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode audit logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}
