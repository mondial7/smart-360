package repositories

import (
	"context"
	"smart360/models"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FakeSubmissionRepository is an in-memory implementation of SubmissionRepository for testing
type FakeSubmissionRepository struct {
	submissions      map[string]*models.Submission // key: id.Hex()
	roundIndex       map[string][]string           // key: round_id.Hex(), value: []submission_id.Hex()
	reviewerIndex    map[string]string             // key: "round_id:reviewer_id", value: submission_id.Hex()
	mu               sync.RWMutex
}

// NewFakeSubmissionRepository creates a new in-memory submission repository
func NewFakeSubmissionRepository() *FakeSubmissionRepository {
	return &FakeSubmissionRepository{
		submissions:   make(map[string]*models.Submission),
		roundIndex:    make(map[string][]string),
		reviewerIndex: make(map[string]string),
	}
}

// FindByRoundID finds all submissions for a given round
func (r *FakeSubmissionRepository) FindByRoundID(ctx context.Context, roundID primitive.ObjectID) ([]models.Submission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	submissionIDs, exists := r.roundIndex[roundID.Hex()]
	if !exists {
		return []models.Submission{}, nil
	}

	submissions := make([]models.Submission, 0, len(submissionIDs))
	for _, submissionID := range submissionIDs {
		if submission, ok := r.submissions[submissionID]; ok {
			submissions = append(submissions, *submission)
		}
	}

	return submissions, nil
}

// Create creates a new submission
func (r *FakeSubmissionRepository) Create(ctx context.Context, submission *models.Submission) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate (same round + reviewer)
	key := submission.RoundID.Hex() + ":" + submission.ReviewerID.Hex()
	if _, exists := r.reviewerIndex[key]; exists {
		return mongo.WriteException{
			WriteErrors: []mongo.WriteError{
				{Code: 11000, Message: "duplicate submission for round and reviewer"},
			},
		}
	}

	// Generate ID if not set
	if submission.ID.IsZero() {
		submission.ID = primitive.NewObjectID()
	}

	// Store submission
	submissionCopy := *submission
	r.submissions[submission.ID.Hex()] = &submissionCopy

	// Update indexes
	roundHex := submission.RoundID.Hex()
	r.roundIndex[roundHex] = append(r.roundIndex[roundHex], submission.ID.Hex())
	r.reviewerIndex[key] = submission.ID.Hex()

	return nil
}

// Update updates an existing submission
func (r *FakeSubmissionRepository) Update(ctx context.Context, submission *models.Submission) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.submissions[submission.ID.Hex()]
	if !exists {
		return mongo.ErrNoDocuments
	}

	// Update submission
	submissionCopy := *submission
	r.submissions[submission.ID.Hex()] = &submissionCopy

	return nil
}

// FindByRoundAndReviewer finds a submission by round and reviewer
func (r *FakeSubmissionRepository) FindByRoundAndReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) (*models.Submission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := roundID.Hex() + ":" + reviewerID.Hex()
	submissionID, exists := r.reviewerIndex[key]
	if !exists {
		return nil, mongo.ErrNoDocuments
	}

	submission, exists := r.submissions[submissionID]
	if !exists {
		return nil, mongo.ErrNoDocuments
	}

	// Return a copy to prevent external modifications
	submissionCopy := *submission
	return &submissionCopy, nil
}

// CountByRoundAndReviewer counts submissions by round and reviewer
func (r *FakeSubmissionRepository) CountByRoundAndReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := roundID.Hex() + ":" + reviewerID.Hex()
	if _, exists := r.reviewerIndex[key]; exists {
		return 1, nil
	}

	return 0, nil
}

// FindByID finds a submission by ID
func (r *FakeSubmissionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Submission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	submission, exists := r.submissions[id.Hex()]
	if !exists {
		return nil, mongo.ErrNoDocuments
	}

	// Return a copy to prevent external modifications
	submissionCopy := *submission
	return &submissionCopy, nil
}
