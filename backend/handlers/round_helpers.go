package handlers

import (
	"context"
	"fmt"
	"smart360/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Response structure for populated round data
type PopulatedRound struct {
	ID          primitive.ObjectID       `bson:"_id,omitempty" json:"id"`
	SubjectID   primitive.ObjectID       `bson:"subject_id" json:"subjectId"`
	Subject     *models.User             `json:"subject,omitempty"`
	CreatedByID primitive.ObjectID       `bson:"created_by_id" json:"createdById"`
	CreatedBy   *models.User             `json:"createdBy,omitempty"`
	Deadline    *time.Time               `bson:"deadline,omitempty" json:"deadline"`
	Status      models.RoundStatus       `bson:"status" json:"status"`
	CreatedAt   time.Time                `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time                `bson:"updated_at" json:"updatedAt"`
	Reviewers   []PopulatedRoundReviewer `bson:"reviewers,omitempty" json:"reviewers,omitempty"`
}

type PopulatedRoundReviewer struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoundID    primitive.ObjectID `bson:"round_id" json:"roundId"`
	ReviewerID primitive.ObjectID `bson:"reviewer_id" json:"reviewerId"`
	Reviewer   *models.User       `json:"reviewer,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
}

// Helper function to get populated round data
func getPopulatedRound(ctx context.Context, db *mongo.Database, roundID primitive.ObjectID) (*PopulatedRound, error) {
	// Get the basic round
	var round models.FeedbackRound
	err := db.Collection("feedback_rounds").FindOne(ctx, bson.M{"_id": roundID}).Decode(&round)
	if err != nil {
		return nil, err
	}

	// Create populated round
	populatedRound := &PopulatedRound{
		ID:          round.ID,
		SubjectID:   round.SubjectID,
		CreatedByID: round.CreatedByID,
		Deadline:    round.Deadline,
		Status:      round.Status,
		CreatedAt:   round.CreatedAt,
		UpdatedAt:   round.UpdatedAt,
		Reviewers:   make([]PopulatedRoundReviewer, 0),
	}

	// Populate subject
	if round.SubjectID != primitive.NilObjectID {
		var subject models.User
		err := db.Collection("users").FindOne(ctx, bson.M{"_id": round.SubjectID}).Decode(&subject)
		if err == nil {
			populatedRound.Subject = &subject
		}
	}

	// Populate created by
	if round.CreatedByID != primitive.NilObjectID {
		var createdBy models.User
		err := db.Collection("users").FindOne(ctx, bson.M{"_id": round.CreatedByID}).Decode(&createdBy)
		if err == nil {
			populatedRound.CreatedBy = &createdBy
		}
	}

	// Populate reviewers - fetch from separate collection
	cursor, err := db.Collection("round_reviewers").Find(ctx, bson.M{"round_id": roundID})
	if err != nil {
		fmt.Printf("Error fetching reviewers for round %s: %v\n", roundID.Hex(), err)
	} else {
		defer cursor.Close(ctx)

		var reviewers []models.RoundReviewer
		if err = cursor.All(ctx, &reviewers); err != nil {
			fmt.Printf("Error decoding reviewers for round %s: %v\n", roundID.Hex(), err)
		} else {
			fmt.Printf("Found %d reviewers for round %s\n", len(reviewers), roundID.Hex())
			for _, reviewer := range reviewers {
				fmt.Printf("  - Reviewer ID: %s\n", reviewer.ReviewerID.Hex())
				populatedReviewer := PopulatedRoundReviewer{
					ID:         reviewer.ID,
					RoundID:    reviewer.RoundID,
					ReviewerID: reviewer.ReviewerID,
					CreatedAt:  reviewer.CreatedAt,
				}

				// Populate reviewer user data
				var reviewerUser models.User
				err := db.Collection("users").FindOne(ctx, bson.M{"_id": reviewer.ReviewerID}).Decode(&reviewerUser)
				if err == nil {
					populatedReviewer.Reviewer = &reviewerUser
					fmt.Printf("    - User: %s\n", reviewerUser.Name)
				} else {
					fmt.Printf("    - Error finding user: %v\n", err)
				}

				populatedRound.Reviewers = append(populatedRound.Reviewers, populatedReviewer)
			}
		}
	}

	return populatedRound, nil
}
