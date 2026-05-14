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
	TemplateID  primitive.ObjectID `bson:"template_id,omitempty" json:"templateId,omitempty"`
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

// ReviewerRelationship describes how the reviewer relates to the subject. It
// lets the consolidation weight signals: a manager's view and a daily peer's
// view carry different evidentiary weight than a cross-functional acquaintance.
type ReviewerRelationship string

const (
	RelationshipManager         ReviewerRelationship = "manager"
	RelationshipReport          ReviewerRelationship = "report"
	RelationshipPeer            ReviewerRelationship = "peer"
	RelationshipCrossFunctional ReviewerRelationship = "cross_functional"
)

func (r ReviewerRelationship) IsValid() bool {
	switch r {
	case RelationshipManager, RelationshipReport, RelationshipPeer, RelationshipCrossFunctional:
		return true
	}
	return false
}

// InteractionFrequency is how often the reviewer worked with the subject in
// the window covered by the round.
type InteractionFrequency string

const (
	InteractionDaily   InteractionFrequency = "daily"
	InteractionWeekly  InteractionFrequency = "weekly"
	InteractionMonthly InteractionFrequency = "monthly"
	InteractionRarely  InteractionFrequency = "rarely"
)

func (f InteractionFrequency) IsValid() bool {
	switch f {
	case InteractionDaily, InteractionWeekly, InteractionMonthly, InteractionRarely:
		return true
	}
	return false
}

type Submission struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	RoundID              primitive.ObjectID   `bson:"round_id" json:"roundId"`
	ReviewerID           primitive.ObjectID   `bson:"reviewer_id" json:"reviewerId"`
	Responses            string               `bson:"responses" json:"responses"` // JSON string
	IsSelf               bool                 `bson:"is_self,omitempty" json:"isSelf,omitempty"`
	Relationship         ReviewerRelationship `bson:"relationship,omitempty" json:"relationship,omitempty"`
	InteractionFrequency InteractionFrequency `bson:"interaction_frequency,omitempty" json:"interactionFrequency,omitempty"`
	SubmittedAt          time.Time            `bson:"submitted_at" json:"submittedAt"`
	UpdatedAt            time.Time            `bson:"updated_at" json:"updatedAt"`
}
