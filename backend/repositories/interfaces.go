package repositories

import (
	"context"
	"smart360/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	UpdateRole(ctx context.Context, id primitive.ObjectID, role models.UserRole) error
	FindAll(ctx context.Context) ([]models.User, error)
}

// RoundRepository defines the interface for feedback round data access
type RoundRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.FeedbackRound, error)
	Create(ctx context.Context, round *models.FeedbackRound) error
	Update(ctx context.Context, round *models.FeedbackRound) error
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.RoundStatus) error
	FindBySubjectID(ctx context.Context, subjectID primitive.ObjectID) ([]models.FeedbackRound, error)
	FindAll(ctx context.Context) ([]models.FeedbackRound, error)
	AddReviewer(ctx context.Context, roundID primitive.ObjectID, reviewer models.RoundReviewer) error
	RemoveReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) error
	GetReviewers(ctx context.Context, roundID primitive.ObjectID) ([]models.RoundReviewer, error)
}

// SubmissionRepository defines the interface for feedback submission data access
type SubmissionRepository interface {
	FindByRoundID(ctx context.Context, roundID primitive.ObjectID) ([]models.Submission, error)
	Create(ctx context.Context, submission *models.Submission) error
	Update(ctx context.Context, submission *models.Submission) error
	FindByRoundAndReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) (*models.Submission, error)
	CountByRoundAndReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) (int64, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Submission, error)
}

// ConsolidationRepository defines the interface for consolidation data access
type ConsolidationRepository interface {
	FindByRoundID(ctx context.Context, roundID primitive.ObjectID) (*models.Consolidation, error)
	Create(ctx context.Context, consolidation *models.Consolidation) error
	Update(ctx context.Context, consolidation *models.Consolidation) error
	UpdateNotes(ctx context.Context, id primitive.ObjectID, notes string) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Consolidation, error)
}
