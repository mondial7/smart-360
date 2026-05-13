package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Consolidation struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoundID             primitive.ObjectID `bson:"round_id" json:"roundId"`
	GeneratedByID       primitive.ObjectID `bson:"generated_by_id" json:"generatedById"`
	ExecutiveSummary    string             `bson:"executive_summary" json:"executiveSummary"`
	Strengths           string             `bson:"strengths" json:"strengths"`                       // JSON array
	AreasForImprovement string             `bson:"areas_for_improvement" json:"areasForImprovement"` // JSON array
	ActionableInsights  string             `bson:"actionable_insights" json:"actionableInsights"`    // JSON array
	QuestionSummaries   string             `bson:"question_summaries" json:"questionSummaries"`      // JSON object
	SelfVsOthersDelta   *SelfVsOthersDelta `bson:"self_vs_others_delta,omitempty" json:"selfVsOthersDelta,omitempty"`
	AdminNotes          string             `bson:"admin_notes" json:"adminNotes"`
	SharedAt            *time.Time         `bson:"shared_at,omitempty" json:"sharedAt"`
	CreatedAt           time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updatedAt"`
}

// SelfVsOthersDelta captures how the subject's self-assessment compares to peer
// feedback. It is the single highest-leverage signal a 360 round produces.
type SelfVsOthersDelta struct {
	SelfSubmitted   bool     `bson:"self_submitted" json:"selfSubmitted"`
	BlindSpots      []string `bson:"blind_spots" json:"blindSpots"`           // peers see, self doesn't
	HiddenStrengths []string `bson:"hidden_strengths" json:"hiddenStrengths"` // self underestimates
	Aligned         []string `bson:"aligned" json:"aligned"`                  // both agree
	Summary         string   `bson:"summary" json:"summary"`                  // 1–2 sentence coaching framing of the gap
}
