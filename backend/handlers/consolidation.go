package handlers

import (
	"context"
	"net/http"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Simple MongoDB-based consolidation handlers
func ConsolidateFeedback(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	roundID := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	// Simple mock consolidation for now
	consolidation := models.Consolidation{
		RoundID:              roundObjID,
		GeneratedByID:        currentUser.ID,
		ExecutiveSummary:     "This is a mock executive summary for development purposes.",
		Strengths:            `["Good communication", "Team collaboration", "Technical skills"]`,
		AreasForImprovement:  `["Time management", "Documentation", "Code reviews"]`,
		ActionableInsights:   `["Focus on prioritization", "Improve documentation practices", "Implement regular code reviews"]`,
		QuestionSummaries:    `{"question1": "Summary of responses for question 1", "question2": "Summary of responses for question 2"}`,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	_, err = db.Collection("consolidations").InsertOne(ctx, consolidation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create consolidation"})
		return
	}

	c.JSON(http.StatusCreated, consolidation)
}

func GetConsolidation(c *gin.Context) {
	roundID := c.Param("roundId")
	db := database.GetDB()
	ctx := context.Background()

	// Convert roundID string to ObjectID
	roundObjID, err := primitive.ObjectIDFromHex(roundID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round ID"})
		return
	}

	var consolidation models.Consolidation
	err = db.Collection("consolidations").FindOne(ctx, bson.M{"round_id": roundObjID}).Decode(&consolidation)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consolidation not found"})
		return
	}

	c.JSON(http.StatusOK, consolidation)
}

func UpdateConsolidationNotes(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	var req struct {
		AdminNotes string `json:"adminNotes" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert id string to ObjectID
	consolidationObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consolidation ID"})
		return
	}

	update := bson.M{"$set": bson.M{
		"admin_notes": req.AdminNotes,
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("consolidations").UpdateOne(ctx, bson.M{"_id": consolidationObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update consolidation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consolidation notes updated successfully"})
}

func ShareConsolidation(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	ctx := context.Background()

	// Convert id string to ObjectID
	consolidationObjID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consolidation ID"})
		return
	}

	update := bson.M{"$set": bson.M{
		"shared_at": time.Now(),
		"updated_at": time.Now(),
	}}

	_, err = db.Collection("consolidations").UpdateOne(ctx, bson.M{"_id": consolidationObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share consolidation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consolidation shared successfully"})
}

func GetMyConsolidatedFeedback(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	// Find consolidations where user is the subject
	cursor, err := db.Collection("consolidations").Find(ctx, bson.M{
		"generated_by_id": currentUser.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consolidations"})
		return
	}
	defer cursor.Close(ctx)

	var consolidations []models.Consolidation
	if err = cursor.All(ctx, &consolidations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode consolidations"})
		return
	}

	c.JSON(http.StatusOK, consolidations)
}
