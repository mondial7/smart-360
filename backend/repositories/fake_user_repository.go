package repositories

import (
	"context"
	"smart360/models"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FakeUserRepository is an in-memory implementation of UserRepository for testing
type FakeUserRepository struct {
	users      map[string]*models.User // key: id.Hex()
	emailIndex map[string]string       // key: email, value: id.Hex()
	mu         sync.RWMutex
}

// NewFakeUserRepository creates a new in-memory user repository
func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{
		users:      make(map[string]*models.User),
		emailIndex: make(map[string]string),
	}
}

// FindByID finds a user by ID
func (r *FakeUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id.Hex()]
	if !exists {
		return nil, mongo.ErrNoDocuments
	}

	// Return a copy to prevent external modifications
	userCopy := *user
	return &userCopy, nil
}

// FindByEmail finds a user by email
func (r *FakeUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	idHex, exists := r.emailIndex[email]
	if !exists {
		return nil, mongo.ErrNoDocuments
	}

	user, exists := r.users[idHex]
	if !exists {
		return nil, mongo.ErrNoDocuments
	}

	// Return a copy to prevent external modifications
	userCopy := *user
	return &userCopy, nil
}

// Create creates a new user
func (r *FakeUserRepository) Create(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	if _, exists := r.emailIndex[user.Email]; exists {
		return mongo.WriteException{
			WriteErrors: []mongo.WriteError{
				{Code: 11000, Message: "duplicate key error"},
			},
		}
	}

	// Generate ID if not set
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	// Store user
	userCopy := *user
	r.users[user.ID.Hex()] = &userCopy
	r.emailIndex[user.Email] = user.ID.Hex()

	return nil
}

// UpdateRole updates a user's role
func (r *FakeUserRepository) UpdateRole(ctx context.Context, id primitive.ObjectID, role models.UserRole) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id.Hex()]
	if !exists {
		return mongo.ErrNoDocuments
	}

	user.Role = role
	return nil
}

// FindAll returns all users
func (r *FakeUserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, *user)
	}

	return users, nil
}
