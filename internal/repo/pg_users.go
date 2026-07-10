package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgUsers struct{ q querier }

const userColumns = `id, email, name, photo_url, role, team_id, created_at, updated_at, last_login`

func scanUser(row rowScanner) (*models.User, error) {
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PhotoURL, &u.Role, &u.TeamID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLogin); err != nil {
		return nil, normalizeErr(err)
	}
	return &u, nil
}

func (r *pgUsers) FindByID(ctx context.Context, id string) (*models.User, error) {
	return scanUser(r.q.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

func (r *pgUsers) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return scanUser(r.q.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email))
}

func (r *pgUsers) Create(ctx context.Context, u *models.User) error {
	if u.Role == "" {
		u.Role = models.RoleMember
	}
	return r.q.QueryRow(ctx, `
		INSERT INTO users (email, name, photo_url, role, team_id, last_login)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`,
		u.Email, u.Name, u.PhotoURL, u.Role, u.TeamID, u.LastLogin,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *pgUsers) UpdateRole(ctx context.Context, id string, role models.UserRole) error {
	_, err := r.q.Exec(ctx, `UPDATE users SET role = $2, updated_at = now() WHERE id = $1`, id, role)
	return err
}

func (r *pgUsers) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.q.Exec(ctx, `UPDATE users SET last_login = now(), updated_at = now() WHERE id = $1`, id)
	return err
}

func (r *pgUsers) SetTeam(ctx context.Context, userID string, teamID *string) error {
	_, err := r.q.Exec(ctx, `UPDATE users SET team_id = $2, updated_at = now() WHERE id = $1`, userID, teamID)
	return err
}

func (r *pgUsers) FindAll(ctx context.Context) ([]models.User, error) {
	rows, err := r.q.Query(ctx, `SELECT `+userColumns+` FROM users ORDER BY name, email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}
