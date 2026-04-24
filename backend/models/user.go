package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleTeamAdmin UserRole = "team_admin"
	RoleMember    UserRole = "member"
)

type User struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Email     string              `bson:"email" json:"email"`
	Name      string              `bson:"name" json:"name"`
	PhotoURL  string              `bson:"photo_url" json:"photoUrl"`
	Role      UserRole            `bson:"role" json:"role"`
	TeamID    *primitive.ObjectID `bson:"team_id,omitempty" json:"teamId,omitempty"`
	CreatedAt time.Time           `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time           `bson:"updated_at" json:"updatedAt"`
	LastLogin *time.Time          `bson:"last_login,omitempty" json:"lastLogin"`
}
