package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgSessions struct{ q querier }

func (r *pgSessions) Create(ctx context.Context, s *models.Session) error {
	return r.q.QueryRow(ctx, `
		INSERT INTO sessions (user_id, expires_at) VALUES ($1, $2)
		RETURNING id, created_at`,
		s.UserID, s.ExpiresAt,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *pgSessions) FindByID(ctx context.Context, id string) (*models.Session, error) {
	var s models.Session
	if err := r.q.QueryRow(ctx,
		`SELECT id, user_id, created_at, expires_at FROM sessions WHERE id = $1`, id,
	).Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt); err != nil {
		return nil, normalizeErr(err)
	}
	return &s, nil
}

func (r *pgSessions) Delete(ctx context.Context, id string) error {
	_, err := r.q.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (r *pgSessions) DeleteExpired(ctx context.Context) error {
	_, err := r.q.Exec(ctx, `DELETE FROM sessions WHERE expires_at < now()`)
	return err
}
