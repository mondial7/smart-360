package repo

import (
	"context"

	"github.com/mondial7/smart-360/internal/models"
)

type pgSubmissions struct{ q querier }

const submissionColumns = `id, round_id, reviewer_id, responses, is_self, relationship,
	interaction_frequency, ratings, private_notes, submitted_at, updated_at`

func scanSubmission(row rowScanner) (*models.Submission, error) {
	var (
		s         models.Submission
		responses []byte
		ratings   []byte
		rel       *string
		freq      *string
	)
	if err := row.Scan(&s.ID, &s.RoundID, &s.ReviewerID, &responses, &s.IsSelf,
		&rel, &freq, &ratings, &s.PrivateNotes, &s.SubmittedAt, &s.UpdatedAt); err != nil {
		return nil, normalizeErr(err)
	}
	if rel != nil {
		s.Relationship = models.ReviewerRelationship(*rel)
	}
	if freq != nil {
		s.InteractionFrequency = models.InteractionFrequency(*freq)
	}
	if err := decodeJSON(responses, &s.Responses); err != nil {
		return nil, err
	}
	if err := decodeJSON(ratings, &s.Ratings); err != nil {
		return nil, err
	}
	return &s, nil
}

// nullable returns a pointer to s, or nil when s is empty, so empty enum values
// are stored as SQL NULL rather than ”.
func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *pgSubmissions) FindByRoundID(ctx context.Context, roundID string) ([]models.Submission, error) {
	return r.querySubmissions(ctx,
		`SELECT `+submissionColumns+` FROM submissions WHERE round_id = $1 ORDER BY submitted_at`, roundID)
}

func (r *pgSubmissions) FindByReviewerID(ctx context.Context, reviewerID string) ([]models.Submission, error) {
	return r.querySubmissions(ctx,
		`SELECT `+submissionColumns+` FROM submissions WHERE reviewer_id = $1 ORDER BY submitted_at DESC`, reviewerID)
}

func (r *pgSubmissions) querySubmissions(ctx context.Context, sql string, args ...any) ([]models.Submission, error) {
	rows, err := r.q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Submission
	for rows.Next() {
		s, err := scanSubmission(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, *s)
	}
	return subs, rows.Err()
}

func (r *pgSubmissions) FindByRoundAndReviewer(ctx context.Context, roundID, reviewerID string) (*models.Submission, error) {
	return scanSubmission(r.q.QueryRow(ctx,
		`SELECT `+submissionColumns+` FROM submissions WHERE round_id = $1 AND reviewer_id = $2`, roundID, reviewerID))
}

func (r *pgSubmissions) FindByID(ctx context.Context, id string) (*models.Submission, error) {
	return scanSubmission(r.q.QueryRow(ctx,
		`SELECT `+submissionColumns+` FROM submissions WHERE id = $1`, id))
}

func (r *pgSubmissions) CountByRoundAndReviewer(ctx context.Context, roundID, reviewerID string) (int64, error) {
	var n int64
	err := r.q.QueryRow(ctx,
		`SELECT count(*) FROM submissions WHERE round_id = $1 AND reviewer_id = $2`, roundID, reviewerID).Scan(&n)
	return n, err
}

func (r *pgSubmissions) Create(ctx context.Context, s *models.Submission) error {
	if s.Responses == nil {
		s.Responses = map[string]string{}
	}
	return r.q.QueryRow(ctx, `
		INSERT INTO submissions
			(round_id, reviewer_id, responses, is_self, relationship, interaction_frequency, ratings, private_notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, submitted_at, updated_at`,
		s.RoundID, s.ReviewerID, mustJSON(s.Responses), s.IsSelf,
		nullable(string(s.Relationship)), nullable(string(s.InteractionFrequency)),
		mustJSON(s.Ratings), s.PrivateNotes,
	).Scan(&s.ID, &s.SubmittedAt, &s.UpdatedAt)
}

func (r *pgSubmissions) Update(ctx context.Context, s *models.Submission) error {
	_, err := r.q.Exec(ctx, `
		UPDATE submissions SET
			responses = $2, is_self = $3, relationship = $4, interaction_frequency = $5,
			ratings = $6, private_notes = $7, updated_at = now()
		WHERE id = $1`,
		s.ID, mustJSON(s.Responses), s.IsSelf,
		nullable(string(s.Relationship)), nullable(string(s.InteractionFrequency)),
		mustJSON(s.Ratings), s.PrivateNotes)
	return err
}
