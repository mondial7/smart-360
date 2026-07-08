package models

import "time"

// DefaultTemplateSlug is the slug every round falls back to when no explicit
// template is set.
const DefaultTemplateSlug = "default"

// Template defines the question set, coaching persona, and presentation labels
// for a round. Questions and Competencies are persisted as jsonb.
type Template struct {
	ID              string               `json:"id"`
	Slug            string               `json:"slug"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	CoachingPersona string               `json:"coachingPersona"`
	Questions       []TemplateQuestion   `json:"questions"`
	Competencies    []TemplateCompetency `json:"competencies,omitempty"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

// TemplateQuestion is one prompt slot. Key is the storage key on the submission
// Responses map; PeerText/SelfText are the prompts shown to reviewers and the
// subject; CardTitle is the short label used on consolidation cards and the PDF.
type TemplateQuestion struct {
	Key       string `json:"key"`
	PeerText  string `json:"peerText"`
	SelfText  string `json:"selfText"`
	CardTitle string `json:"cardTitle"`
}

// TemplateCompetency is one axis of the Likert rubric. Templates with no
// competencies skip the rubric entirely.
type TemplateCompetency struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
