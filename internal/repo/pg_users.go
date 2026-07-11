package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgUsers struct{ q querier }

const userColumns = `id, email, name, photo_url, role, team_id, created_at, updated_at, last_login, onboarded_at`

func scanUser(row rowScanner) (*models.User, error) {
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PhotoURL, &u.Role, &u.TeamID,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLogin, &u.OnboardedAt); err != nil {
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

// FindByIDs fetches the users for the given ids in one query (order unspecified).
func (r *pgUsers) FindByIDs(ctx context.Context, ids []string) ([]models.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return r.queryUsers(ctx, `SELECT `+userColumns+` FROM users WHERE id = ANY($1::uuid[])`, ids)
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

func (r *pgUsers) MarkOnboarded(ctx context.Context, id string) error {
	_, err := r.q.Exec(ctx,
		`UPDATE users SET onboarded_at = now(), updated_at = now() WHERE id = $1 AND onboarded_at IS NULL`, id)
	return err
}

func (r *pgUsers) SetTeam(ctx context.Context, userID string, teamID *string) error {
	_, err := r.q.Exec(ctx, `UPDATE users SET team_id = $2, updated_at = now() WHERE id = $1`, userID, teamID)
	return err
}

func (r *pgUsers) FindAll(ctx context.Context) ([]models.User, error) {
	return r.queryUsers(ctx, `SELECT `+userColumns+` FROM users ORDER BY name, email`)
}

func (r *pgUsers) FindPaged(ctx context.Context, limit, offset int) ([]models.User, error) {
	return r.queryUsers(ctx,
		`SELECT `+userColumns+` FROM users ORDER BY name, email LIMIT $1 OFFSET $2`, limit, offset)
}

func (r *pgUsers) queryUsers(ctx context.Context, sql string, args ...any) ([]models.User, error) {
	rows, err := r.q.Query(ctx, sql, args...)
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
