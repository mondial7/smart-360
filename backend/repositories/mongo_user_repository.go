package repositories

import (
	"context"
	"smart360/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoUserRepository implements UserRepository using MongoDB
type MongoUserRepository struct {
	db *mongo.Database
}

// NewMongoUserRepository creates a new MongoDB-backed user repository
func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{db: db}
}

// FindByID finds a user by ID
func (r *MongoUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *MongoUserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.Collection("users").InsertOne(ctx, user)
	return err
}

// UpdateRole updates a user's role
func (r *MongoUserRepository) UpdateRole(ctx context.Context, id primitive.ObjectID, role models.UserRole) error {
	_, err := r.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"role": role}},
	)
	return err
}

// FindAll returns all users
func (r *MongoUserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	cursor, err := r.db.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
