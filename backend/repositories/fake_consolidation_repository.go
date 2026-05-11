package repositories

import (
	"context"
	"smart360/models"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FakeConsolidationRepository is an in-memory implementation of ConsolidationRepository for testing.
type FakeConsolidationRepository struct {
	consolidations map[string]*models.Consolidation // key: id.Hex()
	roundIndex     map[string]string                // key: round_id.Hex(), value: consolidation id.Hex()
	mu             sync.RWMutex
}

func NewFakeConsolidationRepository() *FakeConsolidationRepository {
	return &FakeConsolidationRepository{
		consolidations: make(map[string]*models.Consolidation),
		roundIndex:     make(map[string]string),
	}
}

func (r *FakeConsolidationRepository) FindByRoundID(ctx context.Context, roundID primitive.ObjectID) (*models.Consolidation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.roundIndex[roundID.Hex()]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	c, ok := r.consolidations[id]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	cp := *c
	return &cp, nil
}

func (r *FakeConsolidationRepository) Create(ctx context.Context, c *models.Consolidation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c.ID.IsZero() {
		c.ID = primitive.NewObjectID()
	}
	cp := *c
	r.consolidations[c.ID.Hex()] = &cp
	r.roundIndex[c.RoundID.Hex()] = c.ID.Hex()
	return nil
}

func (r *FakeConsolidationRepository) Update(ctx context.Context, c *models.Consolidation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.consolidations[c.ID.Hex()]; !ok {
		return mongo.ErrNoDocuments
	}
	cp := *c
	r.consolidations[c.ID.Hex()] = &cp
	return nil
}

func (r *FakeConsolidationRepository) UpdateNotes(ctx context.Context, id primitive.ObjectID, notes string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.consolidations[id.Hex()]
	if !ok {
		return mongo.ErrNoDocuments
	}
	c.AdminNotes = notes
	return nil
}

func (r *FakeConsolidationRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Consolidation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.consolidations[id.Hex()]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	cp := *c
	return &cp, nil
}
