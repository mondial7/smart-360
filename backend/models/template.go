package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DefaultTemplateSlug is the slug of the template every round falls back to
// when no explicit template_id is set.
const DefaultTemplateSlug = "default"

// Template defines the question set, prompt persona, and presentation labels
// used by a 360 round. Storing this as a record (rather than hard-coding in
// the frontend and prompt) lets us tailor rounds per role family — engineering
// vs product vs design — without changing code.
type Template struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Slug            string              `bson:"slug" json:"slug"`
	Name            string              `bson:"name" json:"name"`
	Description     string              `bson:"description" json:"description"`
	CoachingPersona string              `bson:"coaching_persona" json:"coachingPersona"`
	Questions       []TemplateQuestion  `bson:"questions" json:"questions"`
	CreatedAt       time.Time           `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time           `bson:"updated_at" json:"updatedAt"`
}

// TemplateQuestion is one slot in a 360 round. The Key is the storage key on
// the submission JSON (a/b/c/d for the current 4-question shape). PeerText and
// SelfText are the prompts shown to reviewers and the subject respectively;
// CardTitle is the short label used on consolidation summary cards and the PDF.
type TemplateQuestion struct {
	Key       string `bson:"key" json:"key"`
	PeerText  string `bson:"peer_text" json:"peerText"`
	SelfText  string `bson:"self_text" json:"selfText"`
	CardTitle string `bson:"card_title" json:"cardTitle"`
}
