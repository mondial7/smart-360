// Package repo defines data-access interfaces and their Postgres (pgx)
// implementations, plus in-memory fakes used by handler tests. IDs are UUID
// strings throughout.
package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	UpdateRole(ctx context.Context, id string, role models.UserRole) error
	UpdateLastLogin(ctx context.Context, id string) error
	MarkOnboarded(ctx context.Context, id string) error
	SetTeam(ctx context.Context, userID string, teamID *string) error
	FindAll(ctx context.Context) ([]models.User, error)
	FindPaged(ctx context.Context, limit, offset int) ([]models.User, error)
}

type TeamRepository interface {
	FindByID(ctx context.Context, id string) (*models.Team, error)
	FindAll(ctx context.Context) ([]models.Team, error)
	Create(ctx context.Context, team *models.Team) error
	Update(ctx context.Context, team *models.Team) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, teamID, userID string) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	GetMemberIDs(ctx context.Context, teamID string) ([]string, error)
}

type RoundRepository interface {
	FindByID(ctx context.Context, id string) (*models.FeedbackRound, error)
	Create(ctx context.Context, round *models.FeedbackRound) error
	Update(ctx context.Context, round *models.FeedbackRound) error
	UpdateStatus(ctx context.Context, id string, status models.RoundStatus) error
	FindBySubjectID(ctx context.Context, subjectID string) ([]models.FeedbackRound, error)
	FindByCreatedByID(ctx context.Context, creatorID string) ([]models.FeedbackRound, error)
	FindByReviewerID(ctx context.Context, reviewerID string) ([]models.FeedbackRound, error)
	FindAll(ctx context.Context) ([]models.FeedbackRound, error)
	FindPaged(ctx context.Context, limit, offset int) ([]models.FeedbackRound, error)
	AddReviewer(ctx context.Context, roundID string, reviewer models.RoundReviewer) error
	RemoveReviewer(ctx context.Context, roundID, reviewerID string) error
	GetReviewers(ctx context.Context, roundID string) ([]models.RoundReviewer, error)
}

type SubmissionRepository interface {
	FindByRoundID(ctx context.Context, roundID string) ([]models.Submission, error)
	FindByReviewerID(ctx context.Context, reviewerID string) ([]models.Submission, error)
	Create(ctx context.Context, submission *models.Submission) error
	Update(ctx context.Context, submission *models.Submission) error
	FindByRoundAndReviewer(ctx context.Context, roundID, reviewerID string) (*models.Submission, error)
	CountByRoundAndReviewer(ctx context.Context, roundID, reviewerID string) (int64, error)
	FindByID(ctx context.Context, id string) (*models.Submission, error)
}

type TemplateRepository interface {
	FindByID(ctx context.Context, id string) (*models.Template, error)
	FindBySlug(ctx context.Context, slug string) (*models.Template, error)
	FindAll(ctx context.Context) ([]models.Template, error)
	Upsert(ctx context.Context, template *models.Template) error
}

type ConsolidationRepository interface {
	FindByRoundID(ctx context.Context, roundID string) (*models.Consolidation, error)
	FindByID(ctx context.Context, id string) (*models.Consolidation, error)
	FindSharedBySubjectID(ctx context.Context, subjectID string) ([]models.Consolidation, error)
	Create(ctx context.Context, consolidation *models.Consolidation) error
	Update(ctx context.Context, consolidation *models.Consolidation) error
	UpdateNotes(ctx context.Context, id string, notes string) error
}

type AuditRepository interface {
	Create(ctx context.Context, entry *models.AuditLog) error
	FindAll(ctx context.Context, limit int) ([]models.AuditLog, error)
	FindPaged(ctx context.Context, limit, offset int) ([]models.AuditLog, error)
	FindByRoundID(ctx context.Context, roundID string) ([]models.AuditLog, error)
}

type ModerationRepository interface {
	Create(ctx context.Context, entry *models.ModerationLog) error
	FindByRoundID(ctx context.Context, roundID string) ([]models.ModerationLog, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	FindByID(ctx context.Context, id string) (*models.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

// Repositories bundles every repository so wiring passes a single value.
type Repositories struct {
	Users          UserRepository
	Teams          TeamRepository
	Rounds         RoundRepository
	Submissions    SubmissionRepository
	Templates      TemplateRepository
	Consolidations ConsolidationRepository
	Audit          AuditRepository
	Moderation     ModerationRepository
	Sessions       SessionRepository
}
