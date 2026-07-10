package models

import "time"

type RoundStatus string

const (
	RoundDraft  RoundStatus = "draft"
	RoundActive RoundStatus = "active"
	RoundClosed RoundStatus = "closed"
	RoundShared RoundStatus = "shared"
)

func (s RoundStatus) IsValid() bool {
	switch s {
	case RoundDraft, RoundActive, RoundClosed, RoundShared:
		return true
	}
	return false
}

// FeedbackRound is one 360 cycle for a subject. Reviewers are stored in the
// round_reviewers join table and populated by the repository on read.
type FeedbackRound struct {
	ID          string          `json:"id"`
	SubjectID   string          `json:"subjectId"`
	CreatedByID string          `json:"createdById"`
	TemplateID  *string         `json:"templateId,omitempty"`
	Deadline    *time.Time      `json:"deadline,omitempty"`
	Status      RoundStatus     `json:"status"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	Reviewers   []RoundReviewer `json:"reviewers,omitempty"`
}

type RoundReviewer struct {
	ID         string    `json:"id"`
	RoundID    string    `json:"roundId"`
	ReviewerID string    `json:"reviewerId"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ReviewerRelationship describes how the reviewer relates to the subject, so
// consolidation can weight signals by vantage.
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

// InteractionFrequency is how often the reviewer worked with the subject in the
// window covered by the round.
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

// Submission is one reviewer's (or the subject's own) feedback for a round.
// Responses maps template question keys to answer text; Ratings holds the Likert
// scores. Both are persisted as jsonb.
type Submission struct {
	ID                   string               `json:"id"`
	RoundID              string               `json:"roundId"`
	ReviewerID           string               `json:"reviewerId"`
	Responses            map[string]string    `json:"responses"`
	IsSelf               bool                 `json:"isSelf"`
	Relationship         ReviewerRelationship `json:"relationship,omitempty"`
	InteractionFrequency InteractionFrequency `json:"interactionFrequency,omitempty"`
	Ratings              []CompetencyRating   `json:"ratings,omitempty"`
	// PrivateNotes is content for the manager's eyes only — never surfaced
	// verbatim to the subject. Peer-only.
	PrivateNotes string    `json:"privateNotes,omitempty"`
	SubmittedAt  time.Time `json:"submittedAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CompetencyRating is a single Likert score (1..5) plus a one-line
// justification for one of the template's competencies.
type CompetencyRating struct {
	Key           string `json:"key"`
	Score         int    `json:"score"`
	Justification string `json:"justification"`
}
