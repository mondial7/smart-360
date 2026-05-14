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
	// QuestionLabels is a denormalised copy of the template's CardTitle per
	// question key at generation time. We snapshot rather than join so the UI
	// always renders labels consistent with the answers actually collected,
	// even if the template is later edited.
	QuestionLabels    map[string]string  `bson:"question_labels,omitempty" json:"questionLabels,omitempty"`
	SelfVsOthersDelta *SelfVsOthersDelta `bson:"self_vs_others_delta,omitempty" json:"selfVsOthersDelta,omitempty"`
	VoiceBreakdown    *VoiceBreakdown    `bson:"voice_breakdown,omitempty" json:"voiceBreakdown,omitempty"`
	AdminNotes        string             `bson:"admin_notes" json:"adminNotes"`
	SharedAt          *time.Time         `bson:"shared_at,omitempty" json:"sharedAt"`
	CreatedAt         time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updatedAt"`
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

// Voice is one vantage's view of the subject — what someone observing from a
// particular relationship (manager, peer, report, …) consistently sees.
type Voice struct {
	ReviewerCount int      `bson:"reviewer_count" json:"reviewerCount"`
	Summary       string   `bson:"summary" json:"summary"` // 1–2 sentence coaching synthesis
	Themes        []string `bson:"themes" json:"themes"`   // 3–5 short, behaviourally-anchored bullets
}

// VoiceBreakdown separates the consolidation by vantage. A manager seeing
// "ready for the next level" and a peer seeing "easy to collaborate with" are
// different load-bearing signals — surfacing them as distinct streams stops
// the consolidation from averaging them into mush.
type VoiceBreakdown struct {
	ManagerVoice *Voice `bson:"manager_voice,omitempty" json:"managerVoice,omitempty"`
	PeerVoice    *Voice `bson:"peer_voice,omitempty" json:"peerVoice,omitempty"`
	ReportVoice  *Voice `bson:"report_voice,omitempty" json:"reportVoice,omitempty"`
}
