# Peer-Review Template — Analysis & Recommendations

**Date:** 2026-05-13
**Context:** Engineering leadership + team-psychology review of the current 360 peer-review questions and AI consolidation prompt, with recommendations tuned for a scaleup startup.

---

## 1. Current state

The peer-review template is four open-text questions, hard-coded in `frontend/src/views/SubmitFeedbackView.vue:16-21`:

- **a** — "What are this person's key strengths?"
- **b** — "What areas could this person improve?"
- **c** — "What specific behaviors or actions have you observed that stood out?"
- **d** — "What advice would you give to help this person grow?"

The AI consolidation prompt in `backend/handlers/consolidation.go:384-406` mirrors those four buckets and asks Gemini to act as "an expert and caring Engineering Team Lead" producing an executive summary, strengths, areas for improvement, actionable insights, and per-question summaries.

There is no self-assessment, no rating scale, no rubric/competency anchor, no time window, no calibration of the reviewer's vantage point, no values/behaviour framing, and no anonymity affordance.

---

## 2. Assessment

### 2.1 Engineering leadership lens

1. **Not anchored to a competency framework.** For a scaleup, feedback gets weaponised (promo, layoff, PIP) within months of being introduced. If the questions don't map to your career ladder (execution, scope, ownership, collaboration, technical judgement, mentorship), every reviewer applies a private rubric and managers can't calibrate across teams.
2. **No vantage-point metadata.** A reviewer who pairs daily and one who shipped a single ticket are weighted equally by the AI. Add `relationship` (manager / peer / report / cross-functional) and `frequency of interaction` — both for UX and to let the prompt down-weight thin signals.
3. **Generic prompt, eng-only persona.** The hard-coded "Engineering Team Lead" persona biases output for any non-eng team you'll add (PM, design, ops, sales). Make the persona configurable per round template.
4. **No "do more / do less" framing.** "Areas to improve" produces abstractions; "What should this person *do more of starting now*?" produces actions.
5. **No collaboration-multiplier question.** In a scaleup, the difference between a senior IC and a staff IC is leverage. Today's set rewards individual heroics.
6. **No recency anchor.** Without "in the last 3-6 months", reviewers default to the latest sprint or a single memorable incident.
7. **AI consolidation has no "top 1-3 focus areas" forcing function.** It returns parallel arrays of strengths/improvements/insights. Managers in a scaleup don't have bandwidth to triage 8 bullet points per report — force the model to pick.

### 2.2 Team-psychology lens

1. **Question b is deficit-framed.** Buckingham & Goodall ("The Feedback Fallacy", HBR 2019) and broader growth-mindset research show "what could you improve" elicits defensive, lower-quality input and triggers threat response in the recipient. Reframe to growth-oriented: *"What would unlock the most impact for them in the next 6 months?"*
2. **No psychological safety layer.** Reviewers are identified (no anonymity option) and there's no instruction set ("focus on behaviours, not traits; avoid personality labels"). Critical feedback gets self-censored; the resulting consolidation overweights strengths and looks bland.
3. **No self-assessment.** A 360 without a self-component loses its most valuable signal — the **delta** between self-perception and others'. This is also where most insight for the subject comes from.
4. **Question c is the only behaviourally-anchored one and it's optional.** Promote it to be *the* core question, in **Situation–Behaviour–Impact (SBI)** form. SBI produces feedback that's hard to dismiss and easy to act on.
5. **No bias / toxicity guard in the AI step.** Today the consolidation is shared with the subject. You need a moderation pass that strips identity-targeted or personality-attack content before it reaches them.
6. **Manager voice and IC voice not separated.** Aggregating them dilutes both. A manager's "ready for promotion" and a peer's "easy to work with" should be visible as distinct streams in the consolidation.

---

## 3. Recommended redesign (full)

A scaleup-appropriate template — five sections, behaviourally anchored, growth-framed, AI-friendly:

| # | Question | Type | Why |
|---|----------|------|-----|
| 1 | Relationship & interaction frequency | Single-select | Weights signal; calibrates AI |
| 2 | **Continue** — What does this person do that has the biggest positive impact on the team/product? Give one concrete example (Situation, Behaviour, Impact). | Open + SBI scaffold | Strength-spotting, behaviourally anchored |
| 3 | **Amplify** — If they doubled down on one thing in the next 6 months, what should it be, and what would change? | Open | Growth-framed (replaces deficit "improvements") |
| 4 | **Unlock** — What's currently in their way (skill, habit, environment) that, if addressed, would unlock the next level of impact? | Open | Surfaces blockers without judgement |
| 5 | Rate on 1–5 across 4 axes from your career ladder (e.g. Execution, Collaboration, Ownership, Technical judgement) — and one sentence per rating | Likert + one-line justification | Anchored signal for calibration; one-line justification kills lazy 5s |
| 6 | Anything you'd say to them privately but not in a room? *(optional, surfaced anonymised to the manager only, not the subject)* | Open, restricted audience | Recovers candor without weaponising it |

Add a parallel **self-assessment** with the same questions. The consolidation's job becomes: *show the delta*.

### Round metadata to add to the model

- `time_window` (default last 6 months) — included in the AI prompt.
- `template_id` — so questions are config-driven, not hard-coded, and you can ship different templates per role family (eng / PM / design / sales).
- `reviewer_relationship` per submission.

### AI consolidation changes

- Make the persona configurable per template; default to *"a thoughtful coach using growth-mindset language"*, not "Engineering Team Lead".
- Force the schema to include **`top_focus_areas: max 3`** — this is what the manager actually uses.
- Add a **moderation/rewrite step** that flags personality-trait language, identity-targeted comments, and any direct quote that's identifying, before the subject sees it. The raw reviewer text should never reach the subject verbatim.
- Separate **manager-voice** from **peer-voice** sections in the output.
- Surface the **self vs others delta** explicitly — this is the single highest-leverage thing a 360 can show.

---

## 4. Phase 1 — smallest viable change (this PR)

If you want one PR that gets ~80% of the value without rebuilding the schema:

1. **Rewrite the four question texts** in `frontend/src/views/SubmitFeedbackView.vue:16-21` to the growth-framed Continue / Amplify / Unlock / SBI-example version above.
2. **Update the AI prompt** in `backend/handlers/consolidation.go:384-406`:
   - Drop "Engineering Team Lead" in favour of a coaching persona.
   - Require `top_focus_areas` (max 3).
   - Add an explicit instruction to use behavioural language and avoid trait/personality labels.
   - Add a moderation pass on reviewer text (strip identity-targeted / personality-attack content before it reaches the subject).
3. **Keep the `a / b / c / d` keys** on the wire for backwards compatibility with existing submissions/consolidations — only the visible labels and prompt change.

### Acceptance criteria

- [ ] New question wording is live in the submission form.
- [ ] AI prompt produces a `top_focus_areas` array (max 3 items) on new consolidations.
- [ ] AI prompt instructs the model to (a) use coaching tone, (b) avoid trait labels, (c) flag/strip toxic or identity-targeted content.
- [ ] Existing consolidations still render (backwards compatible).
- [ ] Manual smoke: run one round end-to-end (draft → active → submit → close → consolidate → share) and read the output.

---

## 5. Phase 2 — follow-up

In rough order of value. Status is updated as each ships.

1. ✅ **Self-assessment + delta view** *(shipped)*. Subject fills the same questions in self-flavoured wording; AI consolidation produces `self_vs_others_delta` with `blind_spots`, `hidden_strengths`, `aligned`, and a coaching summary. Rendered in `MyFeedbackView`, `ConsolidationView`, and the PDF.
2. ✅ **Reviewer-relationship metadata** *(shipped)*. Each peer submission now declares `relationship` (manager / report / peer / cross_functional) and `interactionFrequency` (daily / weekly / monthly / rarely). Both are required for peer submissions and skipped for self-assessments. The metadata is injected into the AI prompt per reviewer block and the prompt instructs the model to down-weight thin signals (rare + cross-functional) and to flag single-source themes as hypotheses rather than findings. Surfaced as pill tags in the consolidation per-reviewer breakdown and on the user's own submission view.
3. ✅ **Manager-voice vs peer-voice separation** *(shipped)*. AI consolidation now emits a `voice_breakdown` with three streams: `manager_voice` (from manager-relationship submissions), `peer_voice` (peers + cross-functional), and `report_voice` (people the subject manages). Each voice has a coaching summary and behaviourally-anchored themes, gated by server-side reviewer counts so the AI can't hallucinate a vantage that wasn't represented. Rendered in `MyFeedbackView`, `ConsolidationView`, and the PDF. Non-AI fallback also produces a minimal breakdown from raw answers.
4. ✅ **Configurable templates per role family** *(shipped)*. Round questions, prompt persona, and consolidation card titles are now records in a `templates` collection, not hard-coded. The bundled defaults ("Growth-Framed 360" and "Engineering Leadership 360") are upserted on every boot via `SeedDefaultTemplates`. Rounds carry `template_id`; both wizards (single + team batch) include a template picker on the round-details step. The consolidation persona, per-question prompt labels, and the AI's `question_summaries` schema are driven by the template; per-question display labels are snapshotted onto the consolidation as `questionLabels` so the UI renders them with no extra fetches.
5. **Likert ratings against a competency rubric** (execution / collaboration / ownership / technical judgement, each 1–5 with a one-line justification). Required for cross-team calibration once promo discussions touch this data.
6. **Recency anchor** ("in the last 3–6 months") surfaced in both the UI and the AI prompt.
7. **Private-to-manager channel** for candid input the subject shouldn't see verbatim.
8. **Bias / toxicity moderation as a dedicated step** rather than a prompt instruction (separate LLM call with strict schema and an audit log).

---

## 6. Out of scope (intentionally deferred)

- Re-architecting the consolidation storage to drop the JSON-string-in-Mongo legacy fields (tracked separately).
- Building a full competency-ladder editor in-app — start with a fixed rubric in config.
- Multi-language support for the questions.
