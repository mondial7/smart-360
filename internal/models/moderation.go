package models

import "time"

// ModerationLog is the audit record for one moderation pass over one submission,
// so a rewrite can always be explained after the fact.
type ModerationLog struct {
	ID             string    `json:"id"`
	RoundID        string    `json:"roundId"`
	SubmissionID   string    `json:"submissionId"`
	Model          string    `json:"model"`
	Flagged        bool      `json:"flagged"`
	Reasons        []string  `json:"reasons,omitempty"`
	FieldsScrubbed []string  `json:"fieldsScrubbed,omitempty"`
	ModeratedAt    time.Time `json:"moderatedAt"`
}
