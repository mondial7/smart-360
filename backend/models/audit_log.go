package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditAction string

const (
	// Round lifecycle actions
	AuditRoundCreated         AuditAction = "round.created"
	AuditRoundStatusChanged   AuditAction = "round.status_changed"
	AuditRoundSubjectChanged  AuditAction = "round.subject_changed"
	AuditRoundDeadlineChanged AuditAction = "round.deadline_changed"

	// Reviewer actions
	AuditReviewerAdded   AuditAction = "reviewer.added"
	AuditReviewerRemoved AuditAction = "reviewer.removed"

	// Consolidation actions
	AuditConsolidationCreated AuditAction = "consolidation.created"
	AuditConsolidationShared  AuditAction = "consolidation.shared"
	AuditConsolidationEdited  AuditAction = "consolidation.edited"

	// Team actions
	AuditTeamCreated        AuditAction = "team.created"
	AuditTeamUpdated        AuditAction = "team.updated"
	AuditTeamDeleted        AuditAction = "team.deleted"
	AuditTeamAdminChanged   AuditAction = "team.admin_changed"
	AuditTeamMemberAdded    AuditAction = "team.member_added"
	AuditTeamMemberRemoved  AuditAction = "team.member_removed"
	AuditTeamRoundCreated   AuditAction = "team_round.created"
)

type AuditLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Action       AuditAction        `bson:"action" json:"action"`
	ActorID      primitive.ObjectID `bson:"actor_id" json:"actorId"`             // Who made the change
	ActorName    string             `bson:"actor_name" json:"actorName"`         // Cached for display
	ActorEmail   string             `bson:"actor_email" json:"actorEmail"`       // Cached for display
	RoundID      primitive.ObjectID `bson:"round_id,omitempty" json:"roundId"`   // Which round
	RoundSubject string             `bson:"round_subject,omitempty" json:"roundSubject"` // Cached subject name
	TeamID       primitive.ObjectID `bson:"team_id,omitempty" json:"teamId"`     // Which team
	TeamName     string             `bson:"team_name,omitempty" json:"teamName"` // Cached team name
	Description  string             `bson:"description" json:"description"`      // Human-readable description
	OldValue     string             `bson:"old_value,omitempty" json:"oldValue"` // Optional: previous value
	NewValue     string             `bson:"new_value,omitempty" json:"newValue"` // Optional: new value
	Metadata     string             `bson:"metadata,omitempty" json:"metadata"`  // Optional: JSON string for extra data
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
}
