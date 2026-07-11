package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mondial7/smart-360/internal/models"
)

// Seed ensures every install has at least one usable round template. It is safe
// to run on every boot: templates are upserted on their slug, keeping the
// bundled defaults in sync with the code on each deploy. devMode is accepted for
// future dev-data seeding; template seeding runs unconditionally.
func Seed(ctx context.Context, pool *pgxpool.Pool, devMode bool) error {
	for i := range defaultTemplates {
		t := defaultTemplates[i]
		if err := upsertTemplate(ctx, pool, t); err != nil {
			return fmt.Errorf("seed template %q: %w", t.Slug, err)
		}
	}
	slog.Info("seeded round templates", "count", len(defaultTemplates))
	return nil
}

func upsertTemplate(ctx context.Context, pool *pgxpool.Pool, t models.Template) error {
	questions, err := json.Marshal(t.Questions)
	if err != nil {
		return err
	}
	competencies, err := json.Marshal(t.Competencies)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO templates (slug, name, description, coaching_persona, questions, competencies, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (slug) DO UPDATE SET
			name             = EXCLUDED.name,
			description      = EXCLUDED.description,
			coaching_persona = EXCLUDED.coaching_persona,
			questions        = EXCLUDED.questions,
			competencies     = EXCLUDED.competencies,
			updated_at       = now()`,
		t.Slug, t.Name, t.Description, t.CoachingPersona, string(questions), string(competencies),
	)
	return err
}

// defaultTemplates are the bundled 360 templates seeded on every boot.
var defaultTemplates = []models.Template{
	{
		Slug:        models.DefaultTemplateSlug,
		Name:        "Growth-Framed 360",
		Description: "Continue / Block / Amplify / Experiment — works well for engineering and product teams.",
		CoachingPersona: "a thoughtful coach helping someone grow over the next 6 months. " +
			"Use behavioural, observable language. Avoid trait or personality labels.",
		Questions: []models.TemplateQuestion{
			{Key: "a", CardTitle: "What to continue",
				PeerText: "What does this person do that has the biggest positive impact on the team or product? Where possible, share one concrete example (Situation → Behaviour → Impact).",
				SelfText: "What do you do that has the biggest positive impact on the team or product? Share one concrete example (Situation → Behaviour → Impact)."},
			{Key: "b", CardTitle: "What's blocking growth",
				PeerText: "Looking at the last 3–6 months, what's currently holding this person back from their next level of impact (skill, habit, or environment)?",
				SelfText: "Looking at the last 3–6 months, what's holding you back from your next level of impact (skill, habit, or environment)?"},
			{Key: "c", CardTitle: "Where to double down",
				PeerText: "If this person doubled down on one strength over the next 6 months, what should it be — and what would change for the team?",
				SelfText: "If you doubled down on one of your strengths over the next 6 months, what would it be — and what would change for the team?"},
			{Key: "d", CardTitle: "One experiment to try",
				PeerText: "What's one concrete experiment or focus area you'd suggest they try in the next 30–60 days?",
				SelfText: "What's one concrete experiment or focus area you'd like to try in the next 30–60 days?"},
		},
		Competencies: []models.TemplateCompetency{
			{Key: "execution", Name: "Execution", Description: "Consistently turns plans into shipped, working outcomes."},
			{Key: "collaboration", Name: "Collaboration", Description: "Makes the people around them more effective; communicates honestly and early."},
			{Key: "ownership", Name: "Ownership", Description: "Takes responsibility past the line of their stated scope."},
			{Key: "technical_judgement", Name: "Technical judgement", Description: "Picks the right thing to build and the right way to build it given the trade-offs."},
		},
	},
	{
		Slug:        "engineering-leadership",
		Name:        "Engineering Leadership 360",
		Description: "Tailored for tech leads, EMs, and staff+ engineers. Leans on scope, judgement, and multiplier behaviours.",
		CoachingPersona: "a thoughtful engineering coach helping a tech lead or manager grow over the next 6 months. " +
			"Lean on themes of scope, technical judgement, and how the person makes the team around them better. " +
			"Avoid trait labels; ground everything in observable behaviour.",
		Questions: []models.TemplateQuestion{
			{Key: "a", CardTitle: "Where they multiply",
				PeerText: "Where does this person have the biggest multiplier effect on the team or product? Share one concrete example (Situation → Behaviour → Impact).",
				SelfText: "Where do you feel you have the biggest multiplier effect on your team or product? Share one concrete example."},
			{Key: "b", CardTitle: "Next-level scope",
				PeerText: "What's the next level of scope or judgement this person would need to grow into — and what's keeping them from it today?",
				SelfText: "What's the next level of scope or judgement you want to grow into — and what's keeping you from it today?"},
			{Key: "c", CardTitle: "Lean harder into",
				PeerText: "If they leaned harder into one technical or leadership strength over the next 6 months, which should it be and what would change?",
				SelfText: "If you leaned harder into one technical or leadership strength over the next 6 months, which would it be and what would change?"},
			{Key: "d", CardTitle: "One experiment to try",
				PeerText: "What's one specific habit, ritual, or experiment you'd suggest they try in the next 30–60 days that would unlock impact?",
				SelfText: "What's one specific habit, ritual, or experiment you want to try in the next 30–60 days that would unlock impact?"},
		},
		Competencies: []models.TemplateCompetency{
			{Key: "scope_judgement", Name: "Scope & judgement", Description: "Operates at and slightly beyond the scope expected of their level; picks the right problems."},
			{Key: "multiplier", Name: "Multiplier effect", Description: "Makes the team measurably more effective — not just an individual contributor."},
			{Key: "technical_depth", Name: "Technical depth", Description: "Brings deep technical credibility to design and review decisions."},
			{Key: "coaching", Name: "Coaching others", Description: "Grows the people around them through feedback, pairing, and clarity."},
		},
	},
}
