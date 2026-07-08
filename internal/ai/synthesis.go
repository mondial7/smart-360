package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/mondial7/smart-360/internal/models"
)

// Options configures a consolidation run.
type Options struct {
	Submissions   []models.Submission
	Template      *models.Template
	RoundID       string
	GeneratedByID string
	APIKey        string   // empty → non-AI fallback and no moderation
	Progress      Progress // optional; may be nil
}

// Consolidate runs the full pipeline: a per-submission moderation scrub, then
// synthesis (Gemini when an API key is present, deterministic fallback
// otherwise), then the deterministic label snapshot and competency aggregates.
// It returns the consolidation and the moderation audit logs; persistence is
// the caller's responsibility.
func Consolidate(ctx context.Context, opts Options) (models.Consolidation, []models.ModerationLog, error) {
	submissions, logs := moderateSubmissions(ctx, opts.Submissions, opts.RoundID, opts.APIKey, opts.Progress)

	var (
		consolidation models.Consolidation
		err           error
	)
	if opts.APIKey != "" {
		opts.Progress.emit(ProgressEvent{Stage: "synthesizing", Message: "Synthesizing feedback with AI"})
		consolidation, err = generateGeminiConsolidation(ctx, submissions, opts.RoundID, opts.GeneratedByID, opts.APIKey, opts.Template)
		if err != nil {
			return models.Consolidation{}, logs, fmt.Errorf("gemini consolidation: %s", sanitiseErr(err))
		}
	} else {
		opts.Progress.emit(ProgressEvent{Stage: "synthesizing", Message: "Combining feedback"})
		consolidation = combineFeedbackSubmissions(submissions, opts.RoundID, opts.GeneratedByID)
	}

	opts.Progress.emit(ProgressEvent{Stage: "aggregating", Message: "Computing rating aggregates"})
	if consolidation.QuestionLabels == nil {
		consolidation.QuestionLabels = snapshotQuestionLabels(opts.Template)
	}
	if consolidation.CompetencyRatings == nil {
		consolidation.CompetencyRatings = AggregateCompetencyRatings(submissions, opts.Template)
	}

	opts.Progress.emit(ProgressEvent{Stage: "done", Message: "Consolidation ready"})
	return consolidation, logs, nil
}

const consolidationModelName = "gemini-flash-latest"

func generateGeminiConsolidation(ctx context.Context, submissions []models.Submission, roundID, generatedByID, apiKey string, template *models.Template) (models.Consolidation, error) {
	peerTexts, selfText, hasSelf := buildFeedbackPrompts(submissions, template)
	privateNotes := collectPrivateNotes(submissions)

	selfSection := "No self-assessment submitted — return self_vs_others_delta with self_submitted=false and empty arrays."
	if hasSelf {
		selfSection = "Self-assessment from the subject:\n" + selfText
	}

	managerOnlySection := "No private manager-only notes were submitted — return manager_only_channel with note_count=0, empty synthesis, and empty themes."
	if len(privateNotes) > 0 {
		managerOnlySection = "Private notes from reviewers, addressed to the manager only (the subject will NEVER see these):\n" + strings.Join(privateNotes, "\n")
	}

	mgrCount, peerCount, reportCount := countByVoice(submissions)
	voiceContext := fmt.Sprintf(
		"Voice reviewer counts: manager=%d, peer=%d (includes cross-functional), report=%d. "+
			"Produce a summary and themes for every voice; if a voice has zero reviewers, "+
			"return an empty summary string and an empty themes array for that voice.",
		mgrCount, peerCount, reportCount,
	)

	persona := "a thoughtful coach helping someone grow over the next 6 months"
	if template != nil && strings.TrimSpace(template.CoachingPersona) != "" {
		persona = strings.TrimSpace(template.CoachingPersona)
	}

	prompt := fmt.Sprintf(`You are %s.
You are reviewing 360 feedback from multiple peers AND a self-assessment from the subject, and synthesising it for them.

Apply these guidelines strictly:
- Use behavioural, observable language. Avoid trait or personality labels (do not say "they ARE …"). Prefer "they often DO X, which leads to Y".
- Use growth-oriented framing. Replace deficit language ("weakness", "bad at", "lacks") with forward-looking framing ("opportunity to amplify", "would unlock impact by", "next-level habit to build").
- Never reproduce direct quotes that could identify a specific reviewer. Synthesise across reviewers.
- Be specific. Vague compliments ("good communicator", "team player") are useless — ground every point in observable behaviour or impact.
- Anchor your analysis on this person's last 3–6 months.
- Content has already been scrubbed by a separate moderation pass; you should not see identity-targeted or personality-attack language. If you do, drop the offending content silently and proceed.

Weight reviewer signals by relationship and interaction frequency:
- Daily peers and the subject's manager have the richest signal — give their themes more weight, especially for execution and collaboration behaviours.
- Direct reports (subjects who manage them) carry distinct, high-value signal for leadership and feedback behaviours — weight them heavily for those themes.
- Cross-functional collaborators with rare interaction provide thin signal — only surface their themes if they appear in at least one other reviewer's input, otherwise treat them as a hypothesis rather than a finding.
- If a theme appears in only one rarely-interacting reviewer's input, frame it cautiously ("one cross-functional partner observed …") instead of stating it as fact.

Use the Likert ratings (when present) as quantitative anchors for your synthesis:
- A wide spread (e.g., one reviewer at 2 and another at 5 on the same competency) is itself a finding — surface it as a calibration gap to investigate.
- Don't restate the average numbers in the executive summary — the UI shows them. Instead, name the underlying *behaviours* the scores point at.

For the self-vs-others delta:
- blind_spots: things peers consistently flagged that the self-assessment does not acknowledge. Frame as opportunities, not accusations.
- hidden_strengths: things peers value highly that the self-assessment underplays or omits.
- aligned: themes where the self-assessment and peer feedback clearly agree.
- summary: 1–2 sentences in a coaching tone naming the most important gap and why closing it matters.
- If no self-assessment was submitted, set self_submitted=false, return empty arrays, and summary="".

For the voice_breakdown — separate views by vantage so distinct signals do not get averaged into mush:
- manager_voice: synthesise only the feedback from reviewers tagged "manager (manages the subject)". Lean into themes a manager is uniquely placed to see (scope, growth trajectory, readiness, judgement).
- peer_voice: synthesise feedback from reviewers tagged as peers OR cross-functional collaborators. Lean into themes about day-to-day collaboration, execution, and how they affect the people around them.
- report_voice: synthesise feedback from reviewers tagged "direct report (the subject manages them)". Lean into themes a report is uniquely placed to see (clarity, support, feedback they give, psychological safety).
- For each voice, summary is 1–2 sentences in coaching tone, and themes is at most 5 behaviourally-anchored bullets. Do not duplicate the top-level executive summary verbatim — each voice should add what's distinctive about that vantage.
%s

For manager_only_channel — the private notes reviewers addressed to the manager and NOT to the subject:
- note_count is the number of private notes received (we will overwrite this server-side; you can pass back whatever).
- synthesis: 1–2 sentences naming the pattern across the private notes, written in a coaching tone for the manager. Empty string if there are no private notes.
- themes: at most 5 short bullets distilling what reviewers want the manager to know privately. Empty array if no notes.
- This block is for the manager only — it will never be shown to the subject. Speak frankly; you don't need the same diplomatic framing as the subject-facing output.
%s

Feedback data from peer reviewers:
%s

%s

Return a single minified JSON object with this exact shape:
{
  "executive_summary": "2–3 sentences in a coaching tone. Forward-looking. No verdicts.",
  "strengths": ["Behaviourally anchored strengths (at most 5)"],
  "areas_for_improvement": ["Growth-oriented opportunities, NOT deficits (at most 5)"],
  "actionable_insights": ["The TOP 1–3 focus areas the subject should act on first. Never more than 3. Each should be a concrete next step, not advice in the abstract."],
  "question_summaries": {
%s
  },
  "self_vs_others_delta": {
    "self_submitted": true,
    "blind_spots": ["Themes peers see that the self-assessment misses (at most 5)"],
    "hidden_strengths": ["Themes peers value that the self-assessment underplays (at most 5)"],
    "aligned": ["Themes where self and peers clearly agree (at most 5)"],
    "summary": "1–2 sentence coaching framing of the most important gap"
  },
  "voice_breakdown": {
    "manager_voice": {"summary": "1–2 sentences from the manager's vantage, or empty string if no manager reviewer", "themes": ["At most 5 themes, or empty array"]},
    "peer_voice":    {"summary": "1–2 sentences from peer/cross-functional vantage, or empty string if no peer reviewer", "themes": ["At most 5 themes, or empty array"]},
    "report_voice":  {"summary": "1–2 sentences from a direct report's vantage, or empty string if no report reviewer", "themes": ["At most 5 themes, or empty array"]}
  },
  "manager_only_channel": {
    "synthesis": "1–2 sentences for the manager only; empty string if no private notes were submitted",
    "themes": ["Frank, manager-only bullets (at most 5); empty array if no private notes"]
  }
}

Return ONLY the minified JSON object. No code fences, no markdown, no prose.`, persona, voiceContext, managerOnlySection, strings.Join(peerTexts, "\n\n"), selfSection, questionSummariesBlock(template))

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return models.Consolidation{}, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(consolidationModelName)
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = consolidationSchema()

	genCtx, genCancel := context.WithTimeout(ctx, 60*time.Second)
	defer genCancel()
	resp, err := model.GenerateContent(genCtx, genai.Text(prompt))
	if err != nil {
		return models.Consolidation{}, fmt.Errorf("failed to generate content: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return models.Consolidation{}, fmt.Errorf("no response from Gemini")
	}

	var b strings.Builder
	for _, p := range resp.Candidates[0].Content.Parts {
		if tt, ok := p.(genai.Text); ok {
			b.WriteString(string(tt))
		}
	}
	aiResponse := parseAIPayload(b.String(), hasSelf)

	return models.Consolidation{
		RoundID:             roundID,
		GeneratedByID:       generatedByID,
		ExecutiveSummary:    aiResponse.ExecutiveSummary,
		Strengths:           orEmpty(aiResponse.Strengths),
		AreasForImprovement: orEmpty(aiResponse.AreasForImprovement),
		ActionableInsights:  orEmpty(aiResponse.ActionableInsights),
		QuestionSummaries:   aiResponse.QuestionSummaries,
		QuestionLabels:      snapshotQuestionLabels(template),
		SelfVsOthersDelta: &models.SelfVsOthersDelta{
			SelfSubmitted:   aiResponse.SelfVsOthersDelta.SelfSubmitted,
			BlindSpots:      aiResponse.SelfVsOthersDelta.BlindSpots,
			HiddenStrengths: aiResponse.SelfVsOthersDelta.HiddenStrengths,
			Aligned:         aiResponse.SelfVsOthersDelta.Aligned,
			Summary:         aiResponse.SelfVsOthersDelta.Summary,
		},
		VoiceBreakdown:     buildVoiceBreakdown(aiResponse.VoiceBreakdown, mgrCount, peerCount, reportCount),
		ManagerOnlyChannel: buildManagerOnlyChannel(aiResponse.ManagerOnlyChannel, privateNotes),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}, nil
}

func orEmpty(xs []string) []string {
	if xs == nil {
		return []string{}
	}
	return xs
}

// collectPrivateNotes returns each non-empty private note from peer submissions,
// tagged with the relationship label so the manager has context without knowing
// who wrote what. Self submissions never carry private notes.
func collectPrivateNotes(submissions []models.Submission) []string {
	var out []string
	for _, s := range submissions {
		if s.IsSelf {
			continue
		}
		note := strings.TrimSpace(s.PrivateNotes)
		if note == "" {
			continue
		}
		out = append(out, fmt.Sprintf("[%s] %s", relationshipLabel(s.Relationship), note))
	}
	return out
}

func buildManagerOnlyChannel(p aiManagerOnlyPayload, rawNotes []string) *models.ManagerOnlyChannel {
	if len(rawNotes) == 0 {
		return nil
	}
	return &models.ManagerOnlyChannel{
		NoteCount: len(rawNotes),
		Synthesis: p.Synthesis,
		Themes:    p.Themes,
		RawNotes:  rawNotes,
	}
}

// snapshotQuestionLabels copies the template's CardTitle per question key into a
// flat map for persisting on the Consolidation. Returns nil for a nil template.
func snapshotQuestionLabels(template *models.Template) map[string]string {
	if template == nil || len(template.Questions) == 0 {
		return nil
	}
	labels := make(map[string]string, len(template.Questions))
	for _, q := range template.Questions {
		labels[q.Key] = q.CardTitle
	}
	return labels
}

func countByVoice(submissions []models.Submission) (manager, peer, report int) {
	for _, s := range submissions {
		if s.IsSelf {
			continue
		}
		switch s.Relationship {
		case models.RelationshipManager:
			manager++
		case models.RelationshipReport:
			report++
		case models.RelationshipPeer, models.RelationshipCrossFunctional:
			peer++
		}
	}
	return manager, peer, report
}

func buildVoiceBreakdown(p aiVoiceBreakdownPayload, mgrCount, peerCount, reportCount int) *models.VoiceBreakdown {
	if mgrCount == 0 && peerCount == 0 && reportCount == 0 {
		return nil
	}
	vb := &models.VoiceBreakdown{}
	if mgrCount > 0 {
		vb.ManagerVoice = &models.Voice{ReviewerCount: mgrCount, Summary: p.ManagerVoice.Summary, Themes: p.ManagerVoice.Themes}
	}
	if peerCount > 0 {
		vb.PeerVoice = &models.Voice{ReviewerCount: peerCount, Summary: p.PeerVoice.Summary, Themes: p.PeerVoice.Themes}
	}
	if reportCount > 0 {
		vb.ReportVoice = &models.Voice{ReviewerCount: reportCount, Summary: p.ReportVoice.Summary, Themes: p.ReportVoice.Themes}
	}
	return vb
}

// buildFeedbackPrompts splits submissions into peer feedback blocks and the
// subject's self-assessment block, tagging peers with relationship + frequency
// so the model can weight thin vs rich signals.
func buildFeedbackPrompts(submissions []models.Submission, template *models.Template) (peerTexts []string, selfText string, hasSelf bool) {
	labels := questionLabelsFromTemplate(template)
	for _, submission := range submissions {
		responses := submission.Responses

		header := "Feedback from peer reviewer:"
		if submission.IsSelf {
			header = "Self-assessment from the subject:"
		}
		block := header + "\n"
		if !submission.IsSelf {
			block += fmt.Sprintf("Relationship to subject: %s\n", relationshipLabel(submission.Relationship))
			block += fmt.Sprintf("Interaction frequency: %s\n", frequencyLabel(submission.InteractionFrequency))
		}
		if template != nil && len(template.Questions) > 0 {
			for _, q := range template.Questions {
				if ans, ok := responses[q.Key]; ok && ans != "" {
					block += fmt.Sprintf("%s: %s\n", labels[q.Key], ans)
				}
			}
		} else {
			for _, key := range []string{"a", "b", "c", "d"} {
				if ans, ok := responses[key]; ok && ans != "" {
					block += fmt.Sprintf("%s: %s\n", labels[key], ans)
				}
			}
		}

		if len(submission.Ratings) > 0 {
			block += "Ratings (1–5):\n"
			names := competencyNamesByKey(template)
			for _, r := range submission.Ratings {
				name := names[r.Key]
				if name == "" {
					name = r.Key
				}
				block += fmt.Sprintf("  • %s — %d. %s\n", name, r.Score, strings.TrimSpace(r.Justification))
			}
		}

		if submission.IsSelf {
			selfText = block
			hasSelf = true
		} else {
			peerTexts = append(peerTexts, block)
		}
	}
	return peerTexts, selfText, hasSelf
}

func competencyNamesByKey(template *models.Template) map[string]string {
	if template == nil {
		return nil
	}
	out := make(map[string]string, len(template.Competencies))
	for _, c := range template.Competencies {
		out[c.Key] = c.Name
	}
	return out
}

func questionLabelsFromTemplate(template *models.Template) map[string]string {
	labels := map[string]string{
		"a": "What to continue (biggest positive impact, with example)",
		"b": "What's blocking growth (last 3–6 months)",
		"c": "Where to double down (one strength to amplify)",
		"d": "Suggested experiment (next 30–60 days)",
	}
	if template == nil {
		return labels
	}
	for _, q := range template.Questions {
		if q.CardTitle != "" {
			labels[q.Key] = q.CardTitle
		}
	}
	return labels
}

func questionSummariesBlock(template *models.Template) string {
	keys := []string{"a", "b", "c", "d"}
	if template != nil && len(template.Questions) > 0 {
		keys = nil
		for _, q := range template.Questions {
			keys = append(keys, q.Key)
		}
	}
	labels := questionLabelsFromTemplate(template)

	var lines []string
	for i, key := range keys {
		comma := ","
		if i == len(keys)-1 {
			comma = ""
		}
		lines = append(lines, fmt.Sprintf(`    "%s": "Synthesis across reviewers about: %s"%s`, key, labels[key], comma))
	}
	return strings.Join(lines, "\n")
}

func relationshipLabel(r models.ReviewerRelationship) string {
	switch r {
	case models.RelationshipManager:
		return "manager (manages the subject)"
	case models.RelationshipReport:
		return "direct report (the subject manages them)"
	case models.RelationshipPeer:
		return "peer (direct teammate)"
	case models.RelationshipCrossFunctional:
		return "cross-functional collaborator (different team)"
	}
	return "unspecified"
}

func frequencyLabel(f models.InteractionFrequency) string {
	switch f {
	case models.InteractionDaily:
		return "daily — works together most days"
	case models.InteractionWeekly:
		return "weekly — syncs at least once a week"
	case models.InteractionMonthly:
		return "monthly — connects occasionally"
	case models.InteractionRarely:
		return "rarely — limited direct interaction"
	}
	return "unspecified"
}

// combineFeedbackSubmissions is the non-AI fallback used when no Gemini key is
// configured: it stitches the raw answers into a readable, honest consolidation
// that tells the manager AI synthesis was unavailable.
func combineFeedbackSubmissions(submissions []models.Submission, roundID, generatedByID string) models.Consolidation {
	var allContinue, allBlockers, allAmplify, allExperiments []string
	var selfResponses map[string]string
	hasSelf := false
	peerCount := 0
	questionSummaries := make(map[string]string)

	type voiceAcc struct {
		count    int
		themes   []string
		blockers []string
	}
	mgrAcc, peerAcc, reportAcc := voiceAcc{}, voiceAcc{}, voiceAcc{}

	for _, submission := range submissions {
		responses := submission.Responses
		if submission.IsSelf {
			selfResponses = responses
			hasSelf = true
			continue
		}
		peerCount++

		var acc *voiceAcc
		switch submission.Relationship {
		case models.RelationshipManager:
			acc = &mgrAcc
		case models.RelationshipReport:
			acc = &reportAcc
		case models.RelationshipPeer, models.RelationshipCrossFunctional:
			acc = &peerAcc
		}
		if acc != nil {
			acc.count++
			if t := responses["a"]; t != "" {
				acc.themes = append(acc.themes, t)
			}
			if t := responses["b"]; t != "" {
				acc.blockers = append(acc.blockers, t)
			}
		}

		if v := responses["a"]; v != "" {
			allContinue = append(allContinue, v)
		}
		if v := responses["b"]; v != "" {
			allBlockers = append(allBlockers, v)
		}
		if v := responses["c"]; v != "" {
			allAmplify = append(allAmplify, v)
		}
		if v := responses["d"]; v != "" {
			allExperiments = append(allExperiments, v)
		}
	}

	executiveSummary := fmt.Sprintf("Consolidated feedback from %d peer reviewers. ", peerCount)
	if hasSelf {
		executiveSummary += "Includes a self-assessment from the subject for delta analysis. "
	}
	if len(allContinue) > 0 {
		executiveSummary += "Reviewers called out concrete behaviours worth continuing. "
	}
	if len(allBlockers) > 0 || len(allExperiments) > 0 {
		executiveSummary += "There are clear opportunities to unlock the next level of impact over the coming months."
	}

	if len(allContinue) > 0 {
		questionSummaries["a"] = "What to continue: " + strings.Join(allContinue, "; ")
	}
	if len(allBlockers) > 0 {
		questionSummaries["b"] = "What's blocking growth: " + strings.Join(allBlockers, "; ")
	}
	if len(allAmplify) > 0 {
		questionSummaries["c"] = "Where to double down: " + strings.Join(allAmplify, "; ")
	}
	if len(allExperiments) > 0 {
		questionSummaries["d"] = "Suggested experiments (next 30–60 days): " + strings.Join(allExperiments, "; ")
	}

	var actionableInsights []string
	for _, experiment := range allExperiments {
		if len(actionableInsights) >= 3 {
			break
		}
		if len(experiment) > 10 {
			actionableInsights = append(actionableInsights, experiment)
		}
	}

	delta := &models.SelfVsOthersDelta{
		SelfSubmitted:   hasSelf,
		BlindSpots:      []string{},
		HiddenStrengths: []string{},
		Aligned:         []string{},
	}
	if hasSelf {
		delta.Summary = "Self-assessment captured — AI-assisted delta unavailable without GEMINI_API_KEY. Compare manually."
		if v := selfResponses["a"]; v != "" {
			delta.Aligned = append(delta.Aligned, "Self — what to continue: "+v)
		}
		if v := selfResponses["b"]; v != "" {
			delta.BlindSpots = append(delta.BlindSpots, "Self — what's blocking growth: "+v)
		}
		if v := selfResponses["c"]; v != "" {
			delta.HiddenStrengths = append(delta.HiddenStrengths, "Self — where to double down: "+v)
		}
	}

	toVoice := func(label string, acc voiceAcc) *models.Voice {
		if acc.count == 0 {
			return nil
		}
		v := &models.Voice{
			ReviewerCount: acc.count,
			Summary:       fmt.Sprintf("%s feedback captured from %d reviewer(s). AI synthesis unavailable without GEMINI_API_KEY — themes below are raw answers, not summaries.", label, acc.count),
			Themes:        []string{},
		}
		if len(acc.themes) > 0 {
			v.Themes = append(v.Themes, "What to continue: "+strings.Join(acc.themes, "; "))
		}
		if len(acc.blockers) > 0 {
			v.Themes = append(v.Themes, "What's blocking growth: "+strings.Join(acc.blockers, "; "))
		}
		return v
	}

	var voiceBreakdown *models.VoiceBreakdown
	if mgrAcc.count+peerAcc.count+reportAcc.count > 0 {
		voiceBreakdown = &models.VoiceBreakdown{
			ManagerVoice: toVoice("Manager", mgrAcc),
			PeerVoice:    toVoice("Peer", peerAcc),
			ReportVoice:  toVoice("Direct report", reportAcc),
		}
	}

	var managerOnly *models.ManagerOnlyChannel
	if rawNotes := collectPrivateNotes(submissions); len(rawNotes) > 0 {
		managerOnly = &models.ManagerOnlyChannel{
			NoteCount: len(rawNotes),
			Synthesis: "AI synthesis unavailable without GEMINI_API_KEY. Read the raw private notes below.",
			Themes:    []string{},
			RawNotes:  rawNotes,
		}
	}

	return models.Consolidation{
		RoundID:             roundID,
		GeneratedByID:       generatedByID,
		ExecutiveSummary:    executiveSummary,
		Strengths:           orEmpty(allContinue),
		AreasForImprovement: orEmpty(allBlockers),
		ActionableInsights:  orEmpty(actionableInsights),
		QuestionSummaries:   questionSummaries,
		SelfVsOthersDelta:   delta,
		VoiceBreakdown:      voiceBreakdown,
		ManagerOnlyChannel:  managerOnly,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}
