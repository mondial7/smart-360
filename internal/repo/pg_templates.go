package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgTemplates struct{ q querier }

const templateColumns = `id, slug, name, description, coaching_persona, questions, competencies, created_at, updated_at`

func scanTemplate(row rowScanner) (*models.Template, error) {
	var (
		t            models.Template
		questions    []byte
		competencies []byte
	)
	if err := row.Scan(&t.ID, &t.Slug, &t.Name, &t.Description, &t.CoachingPersona,
		&questions, &competencies, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, normalizeErr(err)
	}
	if err := decodeJSON(questions, &t.Questions); err != nil {
		return nil, err
	}
	if err := decodeJSON(competencies, &t.Competencies); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *pgTemplates) FindByID(ctx context.Context, id string) (*models.Template, error) {
	return scanTemplate(r.q.QueryRow(ctx, `SELECT `+templateColumns+` FROM templates WHERE id = $1`, id))
}

func (r *pgTemplates) FindBySlug(ctx context.Context, slug string) (*models.Template, error) {
	return scanTemplate(r.q.QueryRow(ctx, `SELECT `+templateColumns+` FROM templates WHERE slug = $1`, slug))
}

func (r *pgTemplates) FindAll(ctx context.Context) ([]models.Template, error) {
	rows, err := r.q.Query(ctx, `SELECT `+templateColumns+` FROM templates ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.Template
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, *t)
	}
	return templates, rows.Err()
}

func (r *pgTemplates) Upsert(ctx context.Context, t *models.Template) error {
	return r.q.QueryRow(ctx, `
		INSERT INTO templates (slug, name, description, coaching_persona, questions, competencies, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (slug) DO UPDATE SET
			name             = EXCLUDED.name,
			description      = EXCLUDED.description,
			coaching_persona = EXCLUDED.coaching_persona,
			questions        = EXCLUDED.questions,
			competencies     = EXCLUDED.competencies,
			updated_at       = now()
		RETURNING id, created_at, updated_at`,
		t.Slug, t.Name, t.Description, t.CoachingPersona,
		mustJSON(t.Questions), mustJSON(t.Competencies),
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}
