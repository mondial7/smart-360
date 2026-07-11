package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgAudit struct{ q querier }

const auditColumns = `id, action, actor_id, actor_name, actor_email, round_id, round_subject,
	team_id, team_name, description, old_value, new_value, metadata, created_at`

func scanAudit(row rowScanner) (*models.AuditLog, error) {
	var a models.AuditLog
	if err := row.Scan(&a.ID, &a.Action, &a.ActorID, &a.ActorName, &a.ActorEmail, &a.RoundID,
		&a.RoundSubject, &a.TeamID, &a.TeamName, &a.Description, &a.OldValue, &a.NewValue,
		&a.Metadata, &a.CreatedAt); err != nil {
		return nil, normalizeErr(err)
	}
	return &a, nil
}

func (r *pgAudit) Create(ctx context.Context, a *models.AuditLog) error {
	return r.q.QueryRow(ctx, `
		INSERT INTO audit_logs
			(action, actor_id, actor_name, actor_email, round_id, round_subject,
			 team_id, team_name, description, old_value, new_value, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at`,
		a.Action, a.ActorID, a.ActorName, a.ActorEmail, a.RoundID, a.RoundSubject,
		a.TeamID, a.TeamName, a.Description, a.OldValue, a.NewValue, a.Metadata,
	).Scan(&a.ID, &a.CreatedAt)
}

func (r *pgAudit) FindAll(ctx context.Context, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 200
	}
	return r.queryAudit(ctx,
		`SELECT `+auditColumns+` FROM audit_logs ORDER BY created_at DESC, id DESC LIMIT $1`, limit)
}

func (r *pgAudit) FindPaged(ctx context.Context, limit, offset int) ([]models.AuditLog, error) {
	return r.queryAudit(ctx,
		`SELECT `+auditColumns+` FROM audit_logs ORDER BY created_at DESC, id DESC LIMIT $1 OFFSET $2`, limit, offset)
}

func (r *pgAudit) FindByRoundID(ctx context.Context, roundID string) ([]models.AuditLog, error) {
	return r.queryAudit(ctx,
		`SELECT `+auditColumns+` FROM audit_logs WHERE round_id = $1 ORDER BY created_at DESC`, roundID)
}

func (r *pgAudit) queryAudit(ctx context.Context, sql string, args ...any) ([]models.AuditLog, error) {
	rows, err := r.q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		a, err := scanAudit(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *a)
	}
	return logs, rows.Err()
}
