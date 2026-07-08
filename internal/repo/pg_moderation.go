package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgModeration struct{ q querier }

func (r *pgModeration) Create(ctx context.Context, m *models.ModerationLog) error {
	return r.q.QueryRow(ctx, `
		INSERT INTO moderation_logs (round_id, submission_id, model, flagged, reasons, fields_scrubbed)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, moderated_at`,
		m.RoundID, m.SubmissionID, m.Model, m.Flagged, mustJSON(m.Reasons), mustJSON(m.FieldsScrubbed),
	).Scan(&m.ID, &m.ModeratedAt)
}

func (r *pgModeration) FindByRoundID(ctx context.Context, roundID string) ([]models.ModerationLog, error) {
	rows, err := r.q.Query(ctx, `
		SELECT id, round_id, submission_id, model, flagged, reasons, fields_scrubbed, moderated_at
		FROM moderation_logs WHERE round_id = $1 ORDER BY moderated_at DESC`, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ModerationLog
	for rows.Next() {
		var (
			m        models.ModerationLog
			reasons  []byte
			scrubbed []byte
		)
		if err := rows.Scan(&m.ID, &m.RoundID, &m.SubmissionID, &m.Model, &m.Flagged,
			&reasons, &scrubbed, &m.ModeratedAt); err != nil {
			return nil, err
		}
		if err := decodeJSON(reasons, &m.Reasons); err != nil {
			return nil, err
		}
		if err := decodeJSON(scrubbed, &m.FieldsScrubbed); err != nil {
			return nil, err
		}
		logs = append(logs, m)
	}
	return logs, rows.Err()
}
