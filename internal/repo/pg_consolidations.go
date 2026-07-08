package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgConsolidations struct{ q querier }

const consolidationColumns = `id, round_id, generated_by_id, executive_summary, strengths,
	areas_for_improvement, actionable_insights, question_summaries, question_labels,
	self_vs_others_delta, voice_breakdown, competency_ratings, manager_only_channel,
	admin_notes, shared_at, created_at, updated_at`

func scanConsolidation(row rowScanner) (*models.Consolidation, error) {
	var (
		c           models.Consolidation
		strengths   []byte
		areas       []byte
		insights    []byte
		qSummaries  []byte
		qLabels     []byte
		delta       []byte
		voice       []byte
		competency  []byte
		managerOnly []byte
	)
	if err := row.Scan(&c.ID, &c.RoundID, &c.GeneratedByID, &c.ExecutiveSummary, &strengths,
		&areas, &insights, &qSummaries, &qLabels, &delta, &voice, &competency, &managerOnly,
		&c.AdminNotes, &c.SharedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, normalizeErr(err)
	}
	for _, d := range []struct {
		src []byte
		dst any
	}{
		{strengths, &c.Strengths}, {areas, &c.AreasForImprovement}, {insights, &c.ActionableInsights},
		{qSummaries, &c.QuestionSummaries}, {qLabels, &c.QuestionLabels}, {delta, &c.SelfVsOthersDelta},
		{voice, &c.VoiceBreakdown}, {competency, &c.CompetencyRatings}, {managerOnly, &c.ManagerOnlyChannel},
	} {
		if err := decodeJSON(d.src, d.dst); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

func (r *pgConsolidations) FindByRoundID(ctx context.Context, roundID string) (*models.Consolidation, error) {
	return scanConsolidation(r.q.QueryRow(ctx,
		`SELECT `+consolidationColumns+` FROM consolidations WHERE round_id = $1`, roundID))
}

func (r *pgConsolidations) FindByID(ctx context.Context, id string) (*models.Consolidation, error) {
	return scanConsolidation(r.q.QueryRow(ctx,
		`SELECT `+consolidationColumns+` FROM consolidations WHERE id = $1`, id))
}

func (r *pgConsolidations) FindSharedBySubjectID(ctx context.Context, subjectID string) ([]models.Consolidation, error) {
	rows, err := r.q.Query(ctx, `
		SELECT `+prefixed("c", consolidationColumns)+`
		FROM consolidations c
		JOIN feedback_rounds fr ON fr.id = c.round_id
		WHERE fr.subject_id = $1 AND c.shared_at IS NOT NULL
		ORDER BY c.shared_at DESC`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Consolidation
	for rows.Next() {
		c, err := scanConsolidation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (r *pgConsolidations) Create(ctx context.Context, c *models.Consolidation) error {
	return r.q.QueryRow(ctx, `
		INSERT INTO consolidations
			(round_id, generated_by_id, executive_summary, strengths, areas_for_improvement,
			 actionable_insights, question_summaries, question_labels, self_vs_others_delta,
			 voice_breakdown, competency_ratings, manager_only_channel, admin_notes, shared_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`,
		c.RoundID, c.GeneratedByID, c.ExecutiveSummary, mustJSON(c.Strengths), mustJSON(c.AreasForImprovement),
		mustJSON(c.ActionableInsights), mustJSON(c.QuestionSummaries), mustJSON(c.QuestionLabels),
		mustJSON(c.SelfVsOthersDelta), mustJSON(c.VoiceBreakdown), mustJSON(c.CompetencyRatings),
		mustJSON(c.ManagerOnlyChannel), c.AdminNotes, c.SharedAt,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *pgConsolidations) Update(ctx context.Context, c *models.Consolidation) error {
	_, err := r.q.Exec(ctx, `
		UPDATE consolidations SET
			executive_summary = $2, strengths = $3, areas_for_improvement = $4, actionable_insights = $5,
			question_summaries = $6, question_labels = $7, self_vs_others_delta = $8, voice_breakdown = $9,
			competency_ratings = $10, manager_only_channel = $11, admin_notes = $12, shared_at = $13,
			updated_at = now()
		WHERE id = $1`,
		c.ID, c.ExecutiveSummary, mustJSON(c.Strengths), mustJSON(c.AreasForImprovement), mustJSON(c.ActionableInsights),
		mustJSON(c.QuestionSummaries), mustJSON(c.QuestionLabels), mustJSON(c.SelfVsOthersDelta), mustJSON(c.VoiceBreakdown),
		mustJSON(c.CompetencyRatings), mustJSON(c.ManagerOnlyChannel), c.AdminNotes, c.SharedAt)
	return err
}

func (r *pgConsolidations) UpdateNotes(ctx context.Context, id string, notes string) error {
	_, err := r.q.Exec(ctx,
		`UPDATE consolidations SET admin_notes = $2, updated_at = now() WHERE id = $1`, id, notes)
	return err
}
