package repositories

import (
	"context"
	"smart360/models"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FakeTemplateRepository is an in-memory implementation of TemplateRepository
// for unit and integration tests.
type FakeTemplateRepository struct {
	templates map[string]*models.Template // key: id.Hex()
	slugIndex map[string]string           // key: slug, value: id.Hex()
	mu        sync.RWMutex
}

func NewFakeTemplateRepository() *FakeTemplateRepository {
	return &FakeTemplateRepository{
		templates: make(map[string]*models.Template),
		slugIndex: make(map[string]string),
	}
}

func (r *FakeTemplateRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.templates[id.Hex()]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	cp := *t
	return &cp, nil
}

func (r *FakeTemplateRepository) FindBySlug(ctx context.Context, slug string) (*models.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hex, ok := r.slugIndex[slug]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	cp := *r.templates[hex]
	return &cp, nil
}

func (r *FakeTemplateRepository) FindAll(ctx context.Context) ([]models.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]models.Template, 0, len(r.templates))
	for _, t := range r.templates {
		out = append(out, *t)
	}
	return out, nil
}

func (r *FakeTemplateRepository) Upsert(ctx context.Context, template *models.Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	template.UpdatedAt = now

	if existingHex, ok := r.slugIndex[template.Slug]; ok {
		// Preserve the existing ID across upserts so callers can hold on to it.
		if id, err := primitive.ObjectIDFromHex(existingHex); err == nil {
			template.ID = id
		}
	}
	if template.ID.IsZero() {
		template.ID = primitive.NewObjectID()
	}

	cp := *template
	r.templates[cp.ID.Hex()] = &cp
	r.slugIndex[cp.Slug] = cp.ID.Hex()
	return nil
}
