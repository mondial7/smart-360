package handlers

import (
	"smart360/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func templateWithCompetencies() *models.Template {
	return &models.Template{
		Competencies: []models.TemplateCompetency{
			{Key: "execution", Name: "Execution"},
			{Key: "collaboration", Name: "Collaboration"},
		},
	}
}

func TestValidateRatings(t *testing.T) {
	t.Run("nil_template_or_no_competencies_rejects_ratings", func(t *testing.T) {
		err := validateRatings([]models.CompetencyRating{
			{Key: "execution", Score: 4, Justification: "ships"},
		}, &models.Template{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not collect Likert")
	})

	t.Run("template_without_competencies_allows_empty_ratings", func(t *testing.T) {
		assert.NoError(t, validateRatings(nil, &models.Template{}))
	})

	t.Run("must_cover_every_template_competency", func(t *testing.T) {
		tmpl := templateWithCompetencies()
		err := validateRatings([]models.CompetencyRating{
			{Key: "execution", Score: 4, Justification: "ships consistently"},
		}, tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing rating for competency \"collaboration\"")
	})

	t.Run("rejects_unknown_competency_key", func(t *testing.T) {
		tmpl := templateWithCompetencies()
		err := validateRatings([]models.CompetencyRating{
			{Key: "execution", Score: 4, Justification: "ships"},
			{Key: "collaboration", Score: 5, Justification: "great pair"},
			{Key: "unrelated", Score: 3, Justification: "n/a"},
		}, tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown competency key \"unrelated\"")
	})

	t.Run("rejects_duplicate_keys", func(t *testing.T) {
		tmpl := templateWithCompetencies()
		err := validateRatings([]models.CompetencyRating{
			{Key: "execution", Score: 4, Justification: "ships"},
			{Key: "execution", Score: 5, Justification: "also ships"},
			{Key: "collaboration", Score: 4, Justification: "good"},
		}, tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate rating")
	})

	t.Run("rejects_score_out_of_range", func(t *testing.T) {
		tmpl := templateWithCompetencies()
		err := validateRatings([]models.CompetencyRating{
			{Key: "execution", Score: 0, Justification: "x"},
			{Key: "collaboration", Score: 4, Justification: "good"},
		}, tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "between 1 and 5")
	})

	t.Run("rejects_empty_justification", func(t *testing.T) {
		tmpl := templateWithCompetencies()
		err := validateRatings([]models.CompetencyRating{
			{Key: "execution", Score: 4, Justification: "   "},
			{Key: "collaboration", Score: 4, Justification: "good"},
		}, tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "one-line justification")
	})

	t.Run("happy_path", func(t *testing.T) {
		tmpl := templateWithCompetencies()
		assert.NoError(t, validateRatings([]models.CompetencyRating{
			{Key: "execution", Score: 4, Justification: "ships consistently"},
			{Key: "collaboration", Score: 5, Justification: "great pair partner"},
		}, tmpl))
	})
}

func TestAggregateCompetencyRatings(t *testing.T) {
	tmpl := templateWithCompetencies()

	submissions := []models.Submission{
		{
			IsSelf:  true,
			Ratings: []models.CompetencyRating{{Key: "execution", Score: 3}, {Key: "collaboration", Score: 4}},
		},
		{
			Relationship: models.RelationshipManager,
			Ratings:      []models.CompetencyRating{{Key: "execution", Score: 5}, {Key: "collaboration", Score: 4}},
		},
		{
			Relationship: models.RelationshipPeer,
			Ratings:      []models.CompetencyRating{{Key: "execution", Score: 4}, {Key: "collaboration", Score: 5}},
		},
		{
			Relationship: models.RelationshipCrossFunctional, // bucketed under peer
			Ratings:      []models.CompetencyRating{{Key: "execution", Score: 2}, {Key: "collaboration", Score: 3}},
		},
		{
			Relationship: models.RelationshipReport,
			Ratings:      []models.CompetencyRating{{Key: "execution", Score: 5}, {Key: "collaboration", Score: 5}},
		},
	}

	aggregates := aggregateCompetencyRatings(submissions, tmpl)
	require.Len(t, aggregates, 2)

	exec := aggregates[0]
	assert.Equal(t, "execution", exec.Key)
	require.NotNil(t, exec.SelfScore)
	assert.InDelta(t, 3.0, *exec.SelfScore, 0.001)
	require.NotNil(t, exec.ManagerAverage)
	assert.InDelta(t, 5.0, *exec.ManagerAverage, 0.001)
	require.NotNil(t, exec.PeerAverage)
	assert.InDelta(t, 3.0, *exec.PeerAverage, 0.001, "peer voice averages peer + cross_functional")
	require.NotNil(t, exec.ReportAverage)
	assert.InDelta(t, 5.0, *exec.ReportAverage, 0.001)
	require.NotNil(t, exec.OthersAverage)
	assert.InDelta(t, 4.0, *exec.OthersAverage, 0.001, "others average is mean of all non-self ratings")
	assert.Equal(t, 4, exec.OthersCount)
	assert.InDelta(t, 3.0, exec.Spread, 0.001, "spread is max(5) - min(2) across non-self")
}

func TestAggregateCompetencyRatings_NoCompetencies(t *testing.T) {
	assert.Nil(t, aggregateCompetencyRatings(nil, nil))
	assert.Nil(t, aggregateCompetencyRatings(nil, &models.Template{}))
}
