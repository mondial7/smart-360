package ai

import (
	"context"
	"testing"

	"github.com/mondial7/smart-360/internal/models"
)

func testTemplate() *models.Template {
	return &models.Template{
		Slug: "default",
		Questions: []models.TemplateQuestion{
			{Key: "a", CardTitle: "Continue"},
			{Key: "b", CardTitle: "Blocking"},
		},
		Competencies: []models.TemplateCompetency{
			{Key: "execution", Name: "Execution"},
			{Key: "collaboration", Name: "Collaboration"},
		},
	}
}

func TestValidateRatings(t *testing.T) {
	tmpl := testTemplate()

	valid := []models.CompetencyRating{
		{Key: "execution", Score: 4, Justification: "ships"},
		{Key: "collaboration", Score: 5, Justification: "helps"},
	}
	if err := ValidateRatings(valid, tmpl); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}

	missing := valid[:1]
	if err := ValidateRatings(missing, tmpl); err == nil {
		t.Fatal("expected error for missing competency")
	}

	outOfRange := []models.CompetencyRating{
		{Key: "execution", Score: 9, Justification: "x"},
		{Key: "collaboration", Score: 3, Justification: "y"},
	}
	if err := ValidateRatings(outOfRange, tmpl); err == nil {
		t.Fatal("expected error for out-of-range score")
	}

	// A template with no competencies must reject any ratings.
	noComp := &models.Template{Slug: "bare"}
	if err := ValidateRatings(valid, noComp); err == nil {
		t.Fatal("expected error: ratings on a rubric-less template")
	}
	if err := ValidateRatings(nil, noComp); err != nil {
		t.Fatalf("expected no error for empty ratings on bare template, got %v", err)
	}
}

func TestAggregateCompetencyRatings(t *testing.T) {
	tmpl := testTemplate()
	subs := []models.Submission{
		{IsSelf: true, Ratings: []models.CompetencyRating{{Key: "execution", Score: 5}}},
		{Relationship: models.RelationshipManager, Ratings: []models.CompetencyRating{{Key: "execution", Score: 3}}},
		{Relationship: models.RelationshipPeer, Ratings: []models.CompetencyRating{{Key: "execution", Score: 4}}},
	}

	aggs := AggregateCompetencyRatings(subs, tmpl)
	if len(aggs) != 2 {
		t.Fatalf("expected 2 aggregates (template order), got %d", len(aggs))
	}
	exec := aggs[0]
	if exec.Key != "execution" {
		t.Fatalf("expected execution first, got %q", exec.Key)
	}
	if exec.SelfScore == nil || *exec.SelfScore != 5 {
		t.Fatalf("self score wrong: %+v", exec.SelfScore)
	}
	if exec.ManagerAverage == nil || *exec.ManagerAverage != 3 {
		t.Fatalf("manager avg wrong: %+v", exec.ManagerAverage)
	}
	if exec.OthersCount != 2 {
		t.Fatalf("expected 2 non-self ratings, got %d", exec.OthersCount)
	}
	if exec.OthersAverage == nil || *exec.OthersAverage != 3.5 {
		t.Fatalf("others avg wrong: %+v", exec.OthersAverage)
	}
	if exec.Spread != 1 { // max 4 - min 3
		t.Fatalf("expected spread 1, got %v", exec.Spread)
	}
}

func TestConsolidate_FallbackWithoutAPIKey(t *testing.T) {
	tmpl := testTemplate()
	subs := []models.Submission{
		{
			IsSelf:    true,
			Responses: map[string]string{"a": "I mentor juniors", "b": "spread thin"},
			Ratings:   []models.CompetencyRating{{Key: "execution", Score: 4, Justification: "s"}, {Key: "collaboration", Score: 4, Justification: "c"}},
		},
		{
			Relationship: models.RelationshipManager,
			Responses:    map[string]string{"a": "drives clarity", "b": "could delegate more"},
			PrivateNotes: "watch for burnout",
			Ratings:      []models.CompetencyRating{{Key: "execution", Score: 5, Justification: "s"}, {Key: "collaboration", Score: 4, Justification: "c"}},
		},
	}

	c, logs, err := Consolidate(context.Background(), Options{
		Submissions:   subs,
		Template:      tmpl,
		RoundID:       "round-1",
		GeneratedByID: "admin-1",
		APIKey:        "", // fallback path — no network
	})
	if err != nil {
		t.Fatalf("consolidate: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("expected no moderation logs without API key, got %d", len(logs))
	}
	if c.RoundID != "round-1" {
		t.Fatalf("round id not set: %q", c.RoundID)
	}
	if len(c.Strengths) == 0 {
		t.Fatal("expected fallback to collect 'continue' answers as strengths")
	}
	// Competency aggregates are computed deterministically regardless of AI.
	if len(c.CompetencyRatings) != 2 {
		t.Fatalf("expected 2 competency aggregates, got %d", len(c.CompetencyRatings))
	}
	// Question labels are snapshotted from the template.
	if c.QuestionLabels["a"] != "Continue" {
		t.Fatalf("expected snapshotted label, got %q", c.QuestionLabels["a"])
	}
	// Self was submitted → delta reflects it.
	if c.SelfVsOthersDelta == nil || !c.SelfVsOthersDelta.SelfSubmitted {
		t.Fatalf("expected self-submitted delta, got %+v", c.SelfVsOthersDelta)
	}
	// A private note from a non-self reviewer surfaces in the manager-only channel.
	if c.ManagerOnlyChannel == nil || c.ManagerOnlyChannel.NoteCount != 1 {
		t.Fatalf("expected 1 private note in manager channel, got %+v", c.ManagerOnlyChannel)
	}
}
