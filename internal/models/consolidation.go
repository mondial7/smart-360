package models

import "time"

// Consolidation is the synthesized output of a round. The list/map fields that
// were historically stored as JSON strings in Mongo are now proper typed values
// persisted as jsonb columns.
type Consolidation struct {
	ID                  string            `json:"id"`
	RoundID             string            `json:"roundId"`
	GeneratedByID       string            `json:"generatedById"`
	ExecutiveSummary    string            `json:"executiveSummary"`
	Strengths           []string          `json:"strengths"`
	AreasForImprovement []string          `json:"areasForImprovement"`
	ActionableInsights  []string          `json:"actionableInsights"`
	QuestionSummaries   map[string]string `json:"questionSummaries"`
	// QuestionLabels snapshots the template's CardTitle per question key at
	// generation time, so labels always match the answers collected.
	QuestionLabels    map[string]string           `json:"questionLabels,omitempty"`
	SelfVsOthersDelta *SelfVsOthersDelta          `json:"selfVsOthersDelta,omitempty"`
	VoiceBreakdown    *VoiceBreakdown             `json:"voiceBreakdown,omitempty"`
	CompetencyRatings []CompetencyRatingAggregate `json:"competencyRatings,omitempty"`
	// ManagerOnlyChannel synthesizes reviewers' private notes. Stripped from any
	// response or PDF the subject sees.
	ManagerOnlyChannel *ManagerOnlyChannel `json:"managerOnlyChannel,omitempty"`
	AdminNotes         string              `json:"adminNotes"`
	SharedAt           *time.Time          `json:"sharedAt,omitempty"`
	CreatedAt          time.Time           `json:"createdAt"`
	UpdatedAt          time.Time           `json:"updatedAt"`
}

// SelfVsOthersDelta compares the subject's self-assessment to peer feedback.
type SelfVsOthersDelta struct {
	SelfSubmitted   bool     `json:"selfSubmitted"`
	BlindSpots      []string `json:"blindSpots"`
	HiddenStrengths []string `json:"hiddenStrengths"`
	Aligned         []string `json:"aligned"`
	Summary         string   `json:"summary"`
}

// Voice is one vantage's view of the subject.
type Voice struct {
	ReviewerCount int      `json:"reviewerCount"`
	Summary       string   `json:"summary"`
	Themes        []string `json:"themes"`
}

// VoiceBreakdown separates the consolidation by vantage (manager / peer /
// report) so distinct signals aren't averaged into mush.
type VoiceBreakdown struct {
	ManagerVoice *Voice `json:"managerVoice,omitempty"`
	PeerVoice    *Voice `json:"peerVoice,omitempty"`
	ReportVoice  *Voice `json:"reportVoice,omitempty"`
}

// ManagerOnlyChannel synthesizes private reviewer notes for the manager only.
// Raw notes are kept anonymised (relationship-tagged, no reviewer identity).
type ManagerOnlyChannel struct {
	NoteCount int      `json:"noteCount"`
	Synthesis string   `json:"synthesis"`
	Themes    []string `json:"themes"`
	RawNotes  []string `json:"rawNotes,omitempty"`
}

// CompetencyRatingAggregate is the deterministic, server-computed view of Likert
// ratings across submissions for one competency on a round.
type CompetencyRatingAggregate struct {
	Key            string   `json:"key"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	SelfScore      *float64 `json:"selfScore,omitempty"`
	PeerAverage    *float64 `json:"peerAverage,omitempty"`
	ManagerAverage *float64 `json:"managerAverage,omitempty"`
	ReportAverage  *float64 `json:"reportAverage,omitempty"`
	OthersAverage  *float64 `json:"othersAverage,omitempty"`
	OthersCount    int      `json:"othersCount"`
	Spread         float64  `json:"spread"`
}
