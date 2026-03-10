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
	Strengths           []string           `bson:"strengths" json:"strengths"`
	AreasForImprovement []string           `bson:"areas_for_improvement" json:"areasForImprovement"`
	ActionableInsights  []string           `bson:"actionable_insights" json:"actionableInsights"`
	QuestionSummaries   map[string]string  `bson:"question_summaries" json:"questionSummaries"`
	AdminNotes          string             `bson:"admin_notes" json:"adminNotes"`
	SharedAt            *time.Time         `bson:"shared_at,omitempty" json:"sharedAt"`
	CreatedAt           time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updatedAt"`
}
