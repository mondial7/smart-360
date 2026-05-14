package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ModerationLog is the audit record for one moderation pass against one
// submission. We persist these so the trust story is auditable: if a subject
// or admin ever asks "why was this rewritten?", we can answer.
type ModerationLog struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoundID        primitive.ObjectID `bson:"round_id" json:"roundId"`
	SubmissionID   primitive.ObjectID `bson:"submission_id" json:"submissionId"`
	Model          string             `bson:"model" json:"model"`
	Flagged        bool               `bson:"flagged" json:"flagged"`
	Reasons        []string           `bson:"reasons,omitempty" json:"reasons,omitempty"`
	FieldsScrubbed []string           `bson:"fields_scrubbed,omitempty" json:"fieldsScrubbed,omitempty"`
	ModeratedAt    time.Time          `bson:"moderated_at" json:"moderatedAt"`
}
