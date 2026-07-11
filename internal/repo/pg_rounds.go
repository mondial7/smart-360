package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgRounds struct{ q querier }

const roundColumns = `id, subject_id, created_by_id, template_id, deadline, status, created_at, updated_at`

func scanRound(row rowScanner) (*models.FeedbackRound, error) {
	var r models.FeedbackRound
	if err := row.Scan(&r.ID, &r.SubjectID, &r.CreatedByID, &r.TemplateID,
		&r.Deadline, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, normalizeErr(err)
	}
	return &r, nil
}

func (r *pgRounds) FindByID(ctx context.Context, id string) (*models.FeedbackRound, error) {
	round, err := scanRound(r.q.QueryRow(ctx, `SELECT `+roundColumns+` FROM feedback_rounds WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	reviewers, err := r.GetReviewers(ctx, id)
	if err != nil {
		return nil, err
	}
	round.Reviewers = reviewers
	return round, nil
}

// queryRounds runs a rounds query and populates each round's reviewers.
func (r *pgRounds) queryRounds(ctx context.Context, sql string, args ...any) ([]models.FeedbackRound, error) {
	rows, err := r.q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rounds []models.FeedbackRound
	for rows.Next() {
		round, err := scanRound(rows)
		if err != nil {
			return nil, err
		}
		rounds = append(rounds, *round)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range rounds {
		reviewers, err := r.GetReviewers(ctx, rounds[i].ID)
		if err != nil {
			return nil, err
		}
		rounds[i].Reviewers = reviewers
	}
	return rounds, nil
}

func (r *pgRounds) FindBySubjectID(ctx context.Context, subjectID string) ([]models.FeedbackRound, error) {
	return r.queryRounds(ctx,
		`SELECT `+roundColumns+` FROM feedback_rounds WHERE subject_id = $1 ORDER BY created_at DESC`, subjectID)
}

func (r *pgRounds) FindByCreatedByID(ctx context.Context, creatorID string) ([]models.FeedbackRound, error) {
	return r.queryRounds(ctx,
		`SELECT `+roundColumns+` FROM feedback_rounds WHERE created_by_id = $1 ORDER BY created_at DESC`, creatorID)
}

func (r *pgRounds) FindByReviewerID(ctx context.Context, reviewerID string) ([]models.FeedbackRound, error) {
	return r.queryRounds(ctx, `
		SELECT `+prefixed("fr", roundColumns)+`
		FROM feedback_rounds fr
		JOIN round_reviewers rr ON rr.round_id = fr.id
		WHERE rr.reviewer_id = $1
		ORDER BY fr.created_at DESC`, reviewerID)
}

func (r *pgRounds) FindAll(ctx context.Context) ([]models.FeedbackRound, error) {
	return r.queryRounds(ctx, `SELECT `+roundColumns+` FROM feedback_rounds ORDER BY created_at DESC, id DESC`)
}

func (r *pgRounds) FindPaged(ctx context.Context, limit, offset int) ([]models.FeedbackRound, error) {
	return r.queryRounds(ctx,
		`SELECT `+roundColumns+` FROM feedback_rounds ORDER BY created_at DESC, id DESC LIMIT $1 OFFSET $2`, limit, offset)
}

func (r *pgRounds) Create(ctx context.Context, round *models.FeedbackRound) error {
	if round.Status == "" {
		round.Status = models.RoundDraft
	}
	return r.q.QueryRow(ctx, `
		INSERT INTO feedback_rounds (subject_id, created_by_id, template_id, deadline, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		round.SubjectID, round.CreatedByID, round.TemplateID, round.Deadline, round.Status,
	).Scan(&round.ID, &round.CreatedAt, &round.UpdatedAt)
}

func (r *pgRounds) Update(ctx context.Context, round *models.FeedbackRound) error {
	_, err := r.q.Exec(ctx, `
		UPDATE feedback_rounds
		SET subject_id = $2, template_id = $3, deadline = $4, status = $5, updated_at = now()
		WHERE id = $1`,
		round.ID, round.SubjectID, round.TemplateID, round.Deadline, round.Status)
	return err
}

func (r *pgRounds) UpdateStatus(ctx context.Context, id string, status models.RoundStatus) error {
	_, err := r.q.Exec(ctx,
		`UPDATE feedback_rounds SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	return err
}

func (r *pgRounds) AddReviewer(ctx context.Context, roundID string, reviewer models.RoundReviewer) error {
	_, err := r.q.Exec(ctx, `
		INSERT INTO round_reviewers (round_id, reviewer_id) VALUES ($1, $2)
		ON CONFLICT (round_id, reviewer_id) DO NOTHING`,
		roundID, reviewer.ReviewerID)
	return err
}

func (r *pgRounds) RemoveReviewer(ctx context.Context, roundID, reviewerID string) error {
	_, err := r.q.Exec(ctx,
		`DELETE FROM round_reviewers WHERE round_id = $1 AND reviewer_id = $2`, roundID, reviewerID)
	return err
}

func (r *pgRounds) GetReviewers(ctx context.Context, roundID string) ([]models.RoundReviewer, error) {
	rows, err := r.q.Query(ctx,
		`SELECT id, round_id, reviewer_id, created_at FROM round_reviewers WHERE round_id = $1 ORDER BY created_at`, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []models.RoundReviewer
	for rows.Next() {
		var rev models.RoundReviewer
		if err := rows.Scan(&rev.ID, &rev.RoundID, &rev.ReviewerID, &rev.CreatedAt); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, rev)
	}
	return reviewers, rows.Err()
}
