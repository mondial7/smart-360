package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Team struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	TeamAdminID primitive.ObjectID   `bson:"team_admin_id" json:"teamAdminId"`
	MemberIDs   []primitive.ObjectID `bson:"member_ids" json:"memberIds"`
	CreatedAt   time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updatedAt"`
}
