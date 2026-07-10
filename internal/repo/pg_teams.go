package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgTeams struct{ q querier }

func scanTeamRow(row rowScanner) (*models.Team, error) {
	var t models.Team
	if err := row.Scan(&t.ID, &t.Name, &t.TeamAdminID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, normalizeErr(err)
	}
	return &t, nil
}

func (r *pgTeams) FindByID(ctx context.Context, id string) (*models.Team, error) {
	t, err := scanTeamRow(r.q.QueryRow(ctx,
		`SELECT id, name, team_admin_id, created_at, updated_at FROM teams WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	members, err := r.GetMemberIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	t.MemberIDs = members
	return t, nil
}

func (r *pgTeams) FindAll(ctx context.Context) ([]models.Team, error) {
	rows, err := r.q.Query(ctx,
		`SELECT id, name, team_admin_id, created_at, updated_at FROM teams ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		t, err := scanTeamRow(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Populate members after iterating (avoids nested queries on the same conn).
	for i := range teams {
		members, err := r.GetMemberIDs(ctx, teams[i].ID)
		if err != nil {
			return nil, err
		}
		teams[i].MemberIDs = members
	}
	return teams, nil
}

func (r *pgTeams) Create(ctx context.Context, t *models.Team) error {
	return r.q.QueryRow(ctx, `
		INSERT INTO teams (name, team_admin_id)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`,
		t.Name, t.TeamAdminID,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *pgTeams) Update(ctx context.Context, t *models.Team) error {
	_, err := r.q.Exec(ctx, `
		UPDATE teams SET name = $2, team_admin_id = $3, updated_at = now() WHERE id = $1`,
		t.ID, t.Name, t.TeamAdminID)
	return err
}

func (r *pgTeams) Delete(ctx context.Context, id string) error {
	// team_members rows cascade; clear the denormalized pointer on users first.
	if _, err := r.q.Exec(ctx, `UPDATE users SET team_id = NULL WHERE team_id = $1`, id); err != nil {
		return err
	}
	_, err := r.q.Exec(ctx, `DELETE FROM teams WHERE id = $1`, id)
	return err
}

func (r *pgTeams) AddMember(ctx context.Context, teamID, userID string) error {
	if _, err := r.q.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)
		ON CONFLICT (team_id, user_id) DO NOTHING`, teamID, userID); err != nil {
		return err
	}
	// Maintain the denormalized single-team pointer used by role checks.
	_, err := r.q.Exec(ctx, `UPDATE users SET team_id = $2, updated_at = now() WHERE id = $1`, userID, teamID)
	return err
}

func (r *pgTeams) RemoveMember(ctx context.Context, teamID, userID string) error {
	if _, err := r.q.Exec(ctx,
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID); err != nil {
		return err
	}
	_, err := r.q.Exec(ctx,
		`UPDATE users SET team_id = NULL, updated_at = now() WHERE id = $1 AND team_id = $2`, userID, teamID)
	return err
}

func (r *pgTeams) GetMemberIDs(ctx context.Context, teamID string) ([]string, error) {
	rows, err := r.q.Query(ctx,
		`SELECT user_id FROM team_members WHERE team_id = $1 ORDER BY created_at`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
