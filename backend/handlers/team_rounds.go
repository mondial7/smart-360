package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamRoundSubject represents a subject and their deadline for team round creation
type TeamRoundSubject struct {
	SubjectID primitive.ObjectID `json:"subjectId" binding:"required"`
	Deadline  *time.Time         `json:"deadline"`
}

// CreateTeamRoundsRequest is the request body for creating team rounds
type CreateTeamRoundsRequest struct {
	Subjects   []TeamRoundSubject `json:"subjects" binding:"required,min=1"`
	TemplateID primitive.ObjectID `json:"templateId,omitempty"`
}

// CreateTeamRoundsResponse is the response for team round creation
type CreateTeamRoundsResponse struct {
	CreatedRounds  []primitive.ObjectID `json:"createdRounds"`
	SuccessCount   int                  `json:"successCount"`
	FailedSubjects []string             `json:"failedSubjects,omitempty"`
}

// CreateTeamRounds creates multiple feedback rounds for team members
func CreateTeamRounds(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	teamID := c.Param("id")

	var req CreateTeamRoundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Convert team ID
	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	// Verify team exists and get team details
	var team models.Team
	err = db.Collection("teams").FindOne(ctx, bson.M{"_id": teamObjID}).Decode(&team)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Authorization check
	if !canManageTeam(currentUser, teamObjID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to create rounds for this team"})
		return
	}

	// All rounds in a batch share the same template — the wizard picks once.
	templateID, err := resolveTemplateIDForCreate(ctx, req.TemplateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := CreateTeamRoundsResponse{
		CreatedRounds:  make([]primitive.ObjectID, 0),
		FailedSubjects: make([]string, 0),
	}

	// Create a round for each subject
	for _, subject := range req.Subjects {
		// Validate subject is a team member
		if !contains(team.MemberIDs, subject.SubjectID) {
			response.FailedSubjects = append(response.FailedSubjects,
				fmt.Sprintf("%s: not a team member", subject.SubjectID.Hex()))
			continue
		}

		// Cannot create round for self
		if subject.SubjectID == currentUser.ID {
			response.FailedSubjects = append(response.FailedSubjects,
				fmt.Sprintf("%s: cannot create round for yourself", subject.SubjectID.Hex()))
			continue
		}

		// Get subject user
		var subjectUser models.User
		err := db.Collection("users").FindOne(ctx, bson.M{"_id": subject.SubjectID}).Decode(&subjectUser)
		if err != nil {
			response.FailedSubjects = append(response.FailedSubjects,
				fmt.Sprintf("%s: user not found", subject.SubjectID.Hex()))
			continue
		}

		// Create the round
		round := models.FeedbackRound{
			SubjectID:   subject.SubjectID,
			CreatedByID: currentUser.ID,
			TemplateID:  templateID,
			Deadline:    subject.Deadline,
			Status:      models.RoundDraft,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Reviewers:   []models.RoundReviewer{}, // Initialize empty reviewers array
		}

		result, err := db.Collection("feedback_rounds").InsertOne(ctx, round)
		if err != nil {
			response.FailedSubjects = append(response.FailedSubjects,
				fmt.Sprintf("%s: failed to create round - %v", subject.SubjectID.Hex(), err))
			continue
		}

		roundID := result.InsertedID.(primitive.ObjectID)
		response.CreatedRounds = append(response.CreatedRounds, roundID)

		// Auto-assign reviewers: all team members except subject and creator
		reviewerIDs := make([]primitive.ObjectID, 0)
		for _, memberID := range team.MemberIDs {
			if memberID != subject.SubjectID && memberID != currentUser.ID {
				reviewerIDs = append(reviewerIDs, memberID)
			}
		}

		// Add reviewers to the round
		for _, reviewerID := range reviewerIDs {
			reviewer := models.RoundReviewer{
				ID:         primitive.NewObjectID(),
				RoundID:    roundID,
				ReviewerID: reviewerID,
				CreatedAt:  time.Now(),
			}

			_, err := db.Collection("feedback_rounds").UpdateOne(
				ctx,
				bson.M{"_id": roundID},
				bson.M{"$push": bson.M{"reviewers": reviewer}},
			)
			if err != nil {
				log.Printf("Failed to add reviewer %s to round %s: %v", reviewerID.Hex(), roundID.Hex(), err)
			}
		}

		// Create audit log
		createAuditLog(ctx, AuditLogParams{
			Action:       models.AuditTeamRoundCreated,
			ActorID:      currentUser.ID,
			ActorName:    currentUser.Name,
			ActorEmail:   currentUser.Email,
			RoundID:      roundID,
			RoundSubject: subjectUser.Name,
			TeamID:       teamObjID,
			TeamName:     team.Name,
			Description:  fmt.Sprintf("Created team round for %s as part of %s team round", subjectUser.Name, team.Name),
			Metadata:     fmt.Sprintf(`{"reviewerCount": %d}`, len(reviewerIDs)),
		})

		response.SuccessCount++
	}

	c.JSON(http.StatusCreated, response)
}
