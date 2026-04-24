package repositories

import (
	"context"
	"smart360/models"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FakeRoundRepository is an in-memory implementation of RoundRepository for testing
type FakeRoundRepository struct {
	rounds         map[string]*models.FeedbackRound // key: id.Hex()
	subjectIndex   map[string][]string              // key: subject_id.Hex(), value: []round_id.Hex()
	mu             sync.RWMutex
}

// NewFakeRoundRepository creates a new in-memory round repository
func NewFakeRoundRepository() *FakeRoundRepository {
	return &FakeRoundRepository{
		rounds:       make(map[string]*models.FeedbackRound),
		subjectIndex: make(map[string][]string),
	}
}

// FindByID finds a round by ID
func (r *FakeRoundRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.FeedbackRound, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	round, exists := r.rounds[id.Hex()]
	if !exists {
		return nil, mongo.ErrNoDocuments
	}

	// Return a copy to prevent external modifications
	roundCopy := *round
	return &roundCopy, nil
}

// Create creates a new round
func (r *FakeRoundRepository) Create(ctx context.Context, round *models.FeedbackRound) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate ID if not set
	if round.ID.IsZero() {
		round.ID = primitive.NewObjectID()
	}

	// Initialize reviewers array if nil
	if round.Reviewers == nil {
		round.Reviewers = []models.RoundReviewer{}
	}

	// Store round
	roundCopy := *round
	r.rounds[round.ID.Hex()] = &roundCopy

	// Update subject index
	subjectHex := round.SubjectID.Hex()
	r.subjectIndex[subjectHex] = append(r.subjectIndex[subjectHex], round.ID.Hex())

	return nil
}

// Update updates an existing round
func (r *FakeRoundRepository) Update(ctx context.Context, round *models.FeedbackRound) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.rounds[round.ID.Hex()]
	if !exists {
		return mongo.ErrNoDocuments
	}

	// Update round
	roundCopy := *round
	r.rounds[round.ID.Hex()] = &roundCopy

	return nil
}

// UpdateStatus updates a round's status
func (r *FakeRoundRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.RoundStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	round, exists := r.rounds[id.Hex()]
	if !exists {
		return mongo.ErrNoDocuments
	}

	round.Status = status
	return nil
}

// FindBySubjectID finds all rounds for a given subject
func (r *FakeRoundRepository) FindBySubjectID(ctx context.Context, subjectID primitive.ObjectID) ([]models.FeedbackRound, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roundIDs, exists := r.subjectIndex[subjectID.Hex()]
	if !exists {
		return []models.FeedbackRound{}, nil
	}

	rounds := make([]models.FeedbackRound, 0, len(roundIDs))
	for _, roundID := range roundIDs {
		if round, ok := r.rounds[roundID]; ok {
			rounds = append(rounds, *round)
		}
	}

	return rounds, nil
}

// FindAll returns all rounds
func (r *FakeRoundRepository) FindAll(ctx context.Context) ([]models.FeedbackRound, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rounds := make([]models.FeedbackRound, 0, len(r.rounds))
	for _, round := range r.rounds {
		rounds = append(rounds, *round)
	}

	return rounds, nil
}

// AddReviewer adds a reviewer to a round
func (r *FakeRoundRepository) AddReviewer(ctx context.Context, roundID primitive.ObjectID, reviewer models.RoundReviewer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	round, exists := r.rounds[roundID.Hex()]
	if !exists {
		return mongo.ErrNoDocuments
	}

	// Check if reviewer already exists
	for _, existingReviewer := range round.Reviewers {
		if existingReviewer.ReviewerID == reviewer.ReviewerID {
			return nil // Already exists, silently succeed
		}
	}

	// Set round ID and generate ID if needed
	reviewer.RoundID = roundID
	if reviewer.ID.IsZero() {
		reviewer.ID = primitive.NewObjectID()
	}

	round.Reviewers = append(round.Reviewers, reviewer)
	return nil
}

// RemoveReviewer removes a reviewer from a round
func (r *FakeRoundRepository) RemoveReviewer(ctx context.Context, roundID, reviewerID primitive.ObjectID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	round, exists := r.rounds[roundID.Hex()]
	if !exists {
		return mongo.ErrNoDocuments
	}

	// Filter out the reviewer
	newReviewers := make([]models.RoundReviewer, 0, len(round.Reviewers))
	for _, reviewer := range round.Reviewers {
		if reviewer.ReviewerID != reviewerID {
			newReviewers = append(newReviewers, reviewer)
		}
	}

	round.Reviewers = newReviewers
	return nil
}

// GetReviewers returns all reviewers for a round
func (r *FakeRoundRepository) GetReviewers(ctx context.Context, roundID primitive.ObjectID) ([]models.RoundReviewer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	round, exists := r.rounds[roundID.Hex()]
	if !exists {
		return nil, mongo.ErrNoDocuments
	}

	// Return a copy of the reviewers slice
	reviewers := make([]models.RoundReviewer, len(round.Reviewers))
	copy(reviewers, round.Reviewers)

	return reviewers, nil
}
