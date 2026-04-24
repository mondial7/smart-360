package handlers

import (
	"context"
	"fmt"
	"smart360/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ValidationError represents a validation error with a message and code
type ValidationError struct {
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// validateStatusTransition enforces forward-only status flow
func validateStatusTransition(current, new models.RoundStatus) error {
	switch current {
	case models.RoundDraft:
		if new != models.RoundActive {
			return &ValidationError{
				Message: fmt.Sprintf("Cannot transition from draft to %s. Draft rounds can only be activated.", new),
				Code:    "INVALID_STATUS_TRANSITION",
			}
		}
	case models.RoundActive:
		if new != models.RoundClosed {
			return &ValidationError{
				Message: fmt.Sprintf("Cannot transition from active to %s. Active rounds can only be closed.", new),
				Code:    "INVALID_STATUS_TRANSITION",
			}
		}
	case models.RoundClosed:
		if new != models.RoundShared {
			return &ValidationError{
				Message: fmt.Sprintf("Cannot transition from closed to %s. Closed rounds can only be shared.", new),
				Code:    "INVALID_STATUS_TRANSITION",
			}
		}
	case models.RoundShared:
		return &ValidationError{
			Message: "Cannot change status of shared rounds. Shared rounds are final.",
			Code:    "INVALID_STATUS_TRANSITION",
		}
	}
	return nil
}

// validateSubjectChange ensures subject can only be changed in Draft status
func validateSubjectChange(status models.RoundStatus) error {
	if status != models.RoundDraft {
		return &ValidationError{
			Message: fmt.Sprintf("Cannot change subject after round is activated. Current status: %s", status),
			Code:    "INVALID_SUBJECT_CHANGE",
		}
	}
	return nil
}

// hasReviewerSubmitted checks if a reviewer has already submitted feedback
func hasReviewerSubmitted(ctx context.Context, db *mongo.Database, roundID, reviewerID primitive.ObjectID) (bool, error) {
	count, err := db.Collection("submissions").CountDocuments(ctx, bson.M{
		"round_id":    roundID,
		"reviewer_id": reviewerID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// validateReviewerRemoval ensures reviewers who have submitted cannot be removed
func validateReviewerRemoval(ctx context.Context, db *mongo.Database, roundID, reviewerID primitive.ObjectID) error {
	hasSubmitted, err := hasReviewerSubmitted(ctx, db, roundID, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to check submission status: %w", err)
	}

	if hasSubmitted {
		return &ValidationError{
			Message: "Cannot remove reviewer who has already submitted feedback",
			Code:    "REVIEWER_HAS_SUBMITTED",
		}
	}

	return nil
}
