package handlers

import (
	"errors"
	"fmt"
	"smart360/models"
	"strings"
)

// validateRatings enforces three rules against the round's template:
//   - if the template defines no competencies, ratings must be empty (we don't
//     accept ratings on rubric-less rounds — they have nowhere to land);
//   - every template competency must be rated exactly once;
//   - scores are 1..5 inclusive and justifications are non-empty.
//
// Returns a user-facing error or nil.
func validateRatings(ratings []models.CompetencyRating, template *models.Template) error {
	if template == nil || len(template.Competencies) == 0 {
		if len(ratings) > 0 {
			return errors.New("this template does not collect Likert ratings; remove the ratings field")
		}
		return nil
	}

	expected := make(map[string]bool, len(template.Competencies))
	for _, c := range template.Competencies {
		expected[c.Key] = false
	}

	for _, r := range ratings {
		seen, known := expected[r.Key]
		if !known {
			return fmt.Errorf("unknown competency key %q", r.Key)
		}
		if seen {
			return fmt.Errorf("duplicate rating for competency %q", r.Key)
		}
		if r.Score < 1 || r.Score > 5 {
			return fmt.Errorf("score for %q must be between 1 and 5", r.Key)
		}
		if strings.TrimSpace(r.Justification) == "" {
			return fmt.Errorf("a one-line justification is required for %q", r.Key)
		}
		expected[r.Key] = true
	}

	for key, seen := range expected {
		if !seen {
			return fmt.Errorf("missing rating for competency %q", key)
		}
	}
	return nil
}

// aggregateCompetencyRatings produces a deterministic per-competency view of
// the Likert ratings across submissions: self score, averages per voice, and
// the spread (max minus min across non-self ratings). The template provides
// the canonical order and human-readable name/description.
func aggregateCompetencyRatings(submissions []models.Submission, template *models.Template) []models.CompetencyRatingAggregate {
	if template == nil || len(template.Competencies) == 0 {
		return nil
	}

	type bucket struct {
		selfScore   *float64
		peerScores  []float64
		mgrScores   []float64
		repScores   []float64
		allOthers   []float64
	}
	by := make(map[string]*bucket, len(template.Competencies))
	for _, c := range template.Competencies {
		by[c.Key] = &bucket{}
	}

	for _, s := range submissions {
		for _, r := range s.Ratings {
			b, ok := by[r.Key]
			if !ok {
				continue // rating for a competency not in this template; ignore
			}
			score := float64(r.Score)
			if s.IsSelf {
				v := score
				b.selfScore = &v
				continue
			}
			b.allOthers = append(b.allOthers, score)
			switch s.Relationship {
			case models.RelationshipManager:
				b.mgrScores = append(b.mgrScores, score)
			case models.RelationshipReport:
				b.repScores = append(b.repScores, score)
			case models.RelationshipPeer, models.RelationshipCrossFunctional:
				b.peerScores = append(b.peerScores, score)
			}
		}
	}

	out := make([]models.CompetencyRatingAggregate, 0, len(template.Competencies))
	for _, c := range template.Competencies {
		b := by[c.Key]
		agg := models.CompetencyRatingAggregate{
			Key:         c.Key,
			Name:        c.Name,
			Description: c.Description,
			SelfScore:   b.selfScore,
			OthersCount: len(b.allOthers),
		}
		if avg, ok := mean(b.peerScores); ok {
			agg.PeerAverage = &avg
		}
		if avg, ok := mean(b.mgrScores); ok {
			agg.ManagerAverage = &avg
		}
		if avg, ok := mean(b.repScores); ok {
			agg.ReportAverage = &avg
		}
		if avg, ok := mean(b.allOthers); ok {
			agg.OthersAverage = &avg
			agg.Spread = maxFloat(b.allOthers) - minFloat(b.allOthers)
		}
		out = append(out, agg)
	}
	return out
}

func mean(xs []float64) (float64, bool) {
	if len(xs) == 0 {
		return 0, false
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs)), true
}

func maxFloat(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

func minFloat(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs[1:] {
		if x < m {
			m = x
		}
	}
	return m
}
