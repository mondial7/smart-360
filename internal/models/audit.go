package models

import "time"

type AuditAction string

const (
	// Round lifecycle
	AuditRoundCreated         AuditAction = "round.created"
	AuditRoundStatusChanged   AuditAction = "round.status_changed"
	AuditRoundSubjectChanged  AuditAction = "round.subject_changed"
	AuditRoundDeadlineChanged AuditAction = "round.deadline_changed"

	// Reviewers
	AuditReviewerAdded   AuditAction = "reviewer.added"
	AuditReviewerRemoved AuditAction = "reviewer.removed"

	// Consolidation
	AuditConsolidationCreated AuditAction = "consolidation.created"
	AuditConsolidationShared  AuditAction = "consolidation.shared"
	AuditConsolidationEdited  AuditAction = "consolidation.edited"

	// Users
	AuditUserRoleChanged AuditAction = "user.role_changed"

	// Teams
	AuditTeamCreated       AuditAction = "team.created"
	AuditTeamUpdated       AuditAction = "team.updated"
	AuditTeamDeleted       AuditAction = "team.deleted"
	AuditTeamAdminChanged  AuditAction = "team.admin_changed"
	AuditTeamMemberAdded   AuditAction = "team.member_added"
	AuditTeamMemberRemoved AuditAction = "team.member_removed"
	AuditTeamRoundCreated  AuditAction = "team_round.created"
)

// AuditLog records who changed what, with display fields cached at write time so
// the log renders without joins even if referenced entities change or vanish.
type AuditLog struct {
	ID           string      `json:"id"`
	Action       AuditAction `json:"action"`
	ActorID      string      `json:"actorId"`
	ActorName    string      `json:"actorName"`
	ActorEmail   string      `json:"actorEmail"`
	RoundID      *string     `json:"roundId,omitempty"`
	RoundSubject string      `json:"roundSubject,omitempty"`
	TeamID       *string     `json:"teamId,omitempty"`
	TeamName     string      `json:"teamName,omitempty"`
	Description  string      `json:"description"`
	OldValue     string      `json:"oldValue,omitempty"`
	NewValue     string      `json:"newValue,omitempty"`
	Metadata     string      `json:"metadata,omitempty"`
	CreatedAt    time.Time   `json:"createdAt"`
}
