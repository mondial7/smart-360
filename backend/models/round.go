package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoundStatus string

const (
	RoundDraft  RoundStatus = "draft"
	RoundActive RoundStatus = "active"
	RoundClosed RoundStatus = "closed"
	RoundShared RoundStatus = "shared"
)

type FeedbackRound struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SubjectID   primitive.ObjectID `bson:"subject_id" json:"subjectId"`
	CreatedByID primitive.ObjectID `bson:"created_by_id" json:"createdById"`
	Deadline    *time.Time         `bson:"deadline,omitempty" json:"deadline"`
	Status      RoundStatus        `bson:"status" json:"status"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
	Reviewers   []RoundReviewer    `bson:"reviewers,omitempty" json:"reviewers,omitempty"`
}

type RoundReviewer struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoundID    primitive.ObjectID `bson:"round_id" json:"roundId"`
	ReviewerID primitive.ObjectID `bson:"reviewer_id" json:"reviewerId"`
	CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
}

type Submission struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoundID     primitive.ObjectID `bson:"round_id" json:"roundId"`
	ReviewerID  primitive.ObjectID `bson:"reviewer_id" json:"reviewerId"`
	Responses   string             `bson:"responses" json:"responses"` // JSON string
	SubmittedAt time.Time          `bson:"submitted_at" json:"submittedAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}
