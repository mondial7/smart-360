package testutil

import (
	"smart360/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewTestUser creates a test user with sensible defaults
func NewTestUser(email string, role models.UserRole) *models.User {
	return &models.User{
		ID:        primitive.NewObjectID(),
		Email:     email,
		Name:      "Test User",
		PhotoURL:  "https://example.com/photo.jpg",
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestUserWithID creates a test user with a specific ID
func NewTestUserWithID(id primitive.ObjectID, email string, role models.UserRole) *models.User {
	return &models.User{
		ID:        id,
		Email:     email,
		Name:      "Test User",
		PhotoURL:  "https://example.com/photo.jpg",
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestRound creates a test feedback round with the given status
func NewTestRound(subjectID, createdByID primitive.ObjectID, status models.RoundStatus) *models.FeedbackRound {
	now := time.Now()
	deadline := now.Add(7 * 24 * time.Hour) // 7 days from now

	return &models.FeedbackRound{
		ID:          primitive.NewObjectID(),
		SubjectID:   subjectID,
		CreatedByID: createdByID,
		Status:      status,
		Deadline:    &deadline,
		CreatedAt:   now,
		UpdatedAt:   now,
		Reviewers:   []models.RoundReviewer{},
	}
}

// NewTestRoundWithID creates a test feedback round with a specific ID
func NewTestRoundWithID(id, subjectID, createdByID primitive.ObjectID, status models.RoundStatus) *models.FeedbackRound {
	now := time.Now()
	deadline := now.Add(7 * 24 * time.Hour)

	return &models.FeedbackRound{
		ID:          id,
		SubjectID:   subjectID,
		CreatedByID: createdByID,
		Status:      status,
		Deadline:    &deadline,
		CreatedAt:   now,
		UpdatedAt:   now,
		Reviewers:   []models.RoundReviewer{},
	}
}

// NewTestSubmission creates a test submission with JSON responses
func NewTestSubmission(roundID, reviewerID primitive.ObjectID, responses string) *models.Submission {
	return &models.Submission{
		ID:          primitive.NewObjectID(),
		RoundID:     roundID,
		ReviewerID:  reviewerID,
		Responses:   responses,
		SubmittedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestObjectID generates a new ObjectID for testing
func NewTestObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

// TestObjectIDFromHex creates an ObjectID from a hex string, panics on error (for tests only)
func TestObjectIDFromHex(hex string) primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		panic("invalid ObjectID hex in test: " + hex)
	}
	return id
}
