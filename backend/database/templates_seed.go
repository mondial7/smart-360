package database

import (
	"context"
	"log"
	"smart360/models"
	"smart360/repositories"
)

// SeedDefaultTemplates ensures every install has at least one usable 360
// template. It is safe to run on every boot — the upsert keys on `slug`, so
// the bundled defaults are kept in sync with the current code on each deploy.
func SeedDefaultTemplates() {
	repo := repositories.NewMongoTemplateRepository(GetDB())
	ctx := context.Background()

	templates := []models.Template{
		{
			Slug:        models.DefaultTemplateSlug,
			Name:        "Growth-Framed 360",
			Description: "Continue / Block / Amplify / Experiment — works well for engineering and product teams.",
			CoachingPersona: "a thoughtful coach helping someone grow over the next 6 months. " +
				"Use behavioural, observable language. Avoid trait or personality labels.",
			Questions: []models.TemplateQuestion{
				{
					Key:       "a",
					PeerText:  "What does this person do that has the biggest positive impact on the team or product? Where possible, share one concrete example (Situation → Behaviour → Impact).",
					SelfText:  "What do you do that has the biggest positive impact on the team or product? Share one concrete example (Situation → Behaviour → Impact).",
					CardTitle: "What to continue",
				},
				{
					Key:       "b",
					PeerText:  "Looking at the last 3–6 months, what's currently holding this person back from their next level of impact (skill, habit, or environment)?",
					SelfText:  "Looking at the last 3–6 months, what's holding you back from your next level of impact (skill, habit, or environment)?",
					CardTitle: "What's blocking growth",
				},
				{
					Key:       "c",
					PeerText:  "If this person doubled down on one strength over the next 6 months, what should it be — and what would change for the team?",
					SelfText:  "If you doubled down on one of your strengths over the next 6 months, what would it be — and what would change for the team?",
					CardTitle: "Where to double down",
				},
				{
					Key:       "d",
					PeerText:  "What's one concrete experiment or focus area you'd suggest they try in the next 30–60 days?",
					SelfText:  "What's one concrete experiment or focus area you'd like to try in the next 30–60 days?",
					CardTitle: "One experiment to try",
				},
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
				{
					Key:       "a",
					PeerText:  "Where does this person have the biggest multiplier effect on the team or product? Share one concrete example (Situation → Behaviour → Impact).",
					SelfText:  "Where do you feel you have the biggest multiplier effect on your team or product? Share one concrete example.",
					CardTitle: "Where they multiply",
				},
				{
					Key:       "b",
					PeerText:  "What's the next level of scope or judgement this person would need to grow into — and what's keeping them from it today?",
					SelfText:  "What's the next level of scope or judgement you want to grow into — and what's keeping you from it today?",
					CardTitle: "Next-level scope",
				},
				{
					Key:       "c",
					PeerText:  "If they leaned harder into one technical or leadership strength over the next 6 months, which should it be and what would change?",
					SelfText:  "If you leaned harder into one technical or leadership strength over the next 6 months, which would it be and what would change?",
					CardTitle: "Lean harder into",
				},
				{
					Key:       "d",
					PeerText:  "What's one specific habit, ritual, or experiment you'd suggest they try in the next 30–60 days that would unlock impact?",
					SelfText:  "What's one specific habit, ritual, or experiment you want to try in the next 30–60 days that would unlock impact?",
					CardTitle: "One experiment to try",
				},
			},
			Competencies: []models.TemplateCompetency{
				{Key: "scope_judgement", Name: "Scope & judgement", Description: "Operates at and slightly beyond the scope expected of their level; picks the right problems."},
				{Key: "multiplier", Name: "Multiplier effect", Description: "Makes the team measurably more effective — not just an individual contributor."},
				{Key: "technical_depth", Name: "Technical depth", Description: "Brings deep technical credibility to design and review decisions."},
				{Key: "coaching", Name: "Coaching others", Description: "Grows the people around them through feedback, pairing, and clarity."},
			},
		},
	}

	for i := range templates {
		t := templates[i]
		if err := repo.Upsert(ctx, &t); err != nil {
			log.Printf("⚠️  failed to seed template %q: %v", t.Slug, err)
			continue
		}
	}
	log.Printf("✅ Seeded %d round template(s)", len(templates))
}
