package repo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
)

func TestUsers_CreateAndLookup(t *testing.T) {
	r := gateway(t)
	ctx := context.Background()

	u := &models.User{Email: "alice@example.com", Name: "Alice"}
	if err := r.Users.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == "" {
		t.Fatal("expected generated id")
	}
	if u.Role != models.RoleMember {
		t.Fatalf("expected default role member, got %q", u.Role)
	}

	got, err := r.Users.FindByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}
	if got.ID != u.ID {
		t.Fatalf("id mismatch: %s vs %s", got.ID, u.ID)
	}

	if _, err := r.Users.FindByEmail(ctx, "nobody@example.com"); !errors.Is(err, repo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUsers_FindByIDs(t *testing.T) {
	r := gateway(t)
	ctx := context.Background()
	a := makeUser(t, r, "a")
	b := makeUser(t, r, "b")
	_ = makeUser(t, r, "c") // not requested

	got, err := r.Users.FindByIDs(ctx, []string{a, b, "" + a}) // dup a is fine
	if err != nil {
		t.Fatalf("find by ids: %v", err)
	}
	ids := map[string]bool{}
	for _, u := range got {
		ids[u.ID] = true
	}
	if !ids[a] || !ids[b] || len(ids) != 2 {
		t.Fatalf("expected exactly {a,b}, got %v", ids)
	}

	// Empty input is a no-op (no query).
	if out, err := r.Users.FindByIDs(ctx, nil); err != nil || len(out) != 0 {
		t.Fatalf("expected empty result for nil ids, got %v (err %v)", out, err)
	}
}

func TestTeams_MembershipJoinTable(t *testing.T) {
	r := gateway(t)
	ctx := context.Background()

	adminID := makeUser(t, r, "admin")
	memberID := makeUser(t, r, "member")

	team := &models.Team{Name: "Platform", TeamAdminID: adminID}
	if err := r.Teams.Create(ctx, team); err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := r.Teams.AddMember(ctx, team.ID, memberID); err != nil {
		t.Fatalf("add member: %v", err)
	}
	// Adding twice is idempotent.
	if err := r.Teams.AddMember(ctx, team.ID, memberID); err != nil {
		t.Fatalf("add member again: %v", err)
	}

	got, err := r.Teams.FindByID(ctx, team.ID)
	if err != nil {
		t.Fatalf("find team: %v", err)
	}
	if len(got.MemberIDs) != 1 || got.MemberIDs[0] != memberID {
		t.Fatalf("expected one member %s, got %v", memberID, got.MemberIDs)
	}

	// AddMember maintains the denormalized users.team_id pointer.
	u, _ := r.Users.FindByID(ctx, memberID)
	if u.TeamID == nil || *u.TeamID != team.ID {
		t.Fatalf("expected user's team_id to be %s, got %v", team.ID, u.TeamID)
	}

	if err := r.Teams.RemoveMember(ctx, team.ID, memberID); err != nil {
		t.Fatalf("remove member: %v", err)
	}
	got, _ = r.Teams.FindByID(ctx, team.ID)
	if len(got.MemberIDs) != 0 {
		t.Fatalf("expected no members, got %v", got.MemberIDs)
	}
}

func TestRounds_ReviewersAndReverseLookup(t *testing.T) {
	r := gateway(t)
	ctx := context.Background()

	subjectID := makeUser(t, r, "subject")
	creatorID := makeUser(t, r, "creator")
	reviewerID := makeUser(t, r, "reviewer")

	round := &models.FeedbackRound{SubjectID: subjectID, CreatedByID: creatorID}
	if err := r.Rounds.Create(ctx, round); err != nil {
		t.Fatalf("create round: %v", err)
	}
	if round.Status != models.RoundDraft {
		t.Fatalf("expected draft, got %q", round.Status)
	}
	if err := r.Rounds.AddReviewer(ctx, round.ID, models.RoundReviewer{ReviewerID: reviewerID}); err != nil {
		t.Fatalf("add reviewer: %v", err)
	}

	got, err := r.Rounds.FindByID(ctx, round.ID)
	if err != nil {
		t.Fatalf("find round: %v", err)
	}
	if len(got.Reviewers) != 1 || got.Reviewers[0].ReviewerID != reviewerID {
		t.Fatalf("expected reviewer %s, got %+v", reviewerID, got.Reviewers)
	}

	// The reviewer can find the round via the reverse lookup.
	forReviewer, err := r.Rounds.FindByReviewerID(ctx, reviewerID)
	if err != nil {
		t.Fatalf("find by reviewer: %v", err)
	}
	if len(forReviewer) != 1 || forReviewer[0].ID != round.ID {
		t.Fatalf("expected round %s in reviewer's list, got %+v", round.ID, forReviewer)
	}
}

func TestSubmissions_UniquePerRoundReviewer(t *testing.T) {
	r := gateway(t)
	ctx := context.Background()

	subjectID := makeUser(t, r, "subject")
	reviewerID := makeUser(t, r, "reviewer")
	round := &models.FeedbackRound{SubjectID: subjectID, CreatedByID: subjectID}
	if err := r.Rounds.Create(ctx, round); err != nil {
		t.Fatalf("create round: %v", err)
	}

	s := &models.Submission{
		RoundID:      round.ID,
		ReviewerID:   reviewerID,
		Responses:    map[string]string{"a": "answer", "b": "another"},
		Relationship: models.RelationshipPeer,
		Ratings:      []models.CompetencyRating{{Key: "execution", Score: 4, Justification: "ships"}},
	}
	if err := r.Submissions.Create(ctx, s); err != nil {
		t.Fatalf("create submission: %v", err)
	}

	// jsonb round-trips both the map and the ratings slice.
	got, err := r.Submissions.FindByRoundAndReviewer(ctx, round.ID, reviewerID)
	if err != nil {
		t.Fatalf("find submission: %v", err)
	}
	if got.Responses["a"] != "answer" || len(got.Ratings) != 1 || got.Ratings[0].Score != 4 {
		t.Fatalf("jsonb round-trip mismatch: %+v", got)
	}
	if got.Relationship != models.RelationshipPeer {
		t.Fatalf("relationship mismatch: %q", got.Relationship)
	}

	// A second submission by the same reviewer on the same round is rejected.
	dup := &models.Submission{RoundID: round.ID, ReviewerID: reviewerID, Responses: map[string]string{}}
	if err := r.Submissions.Create(ctx, dup); err == nil {
		t.Fatal("expected unique-violation error on duplicate submission")
	}

	n, err := r.Submissions.CountByRoundAndReviewer(ctx, round.ID, reviewerID)
	if err != nil || n != 1 {
		t.Fatalf("expected count 1, got %d (err %v)", n, err)
	}
}

func TestConsolidations_JSONBRoundtripAndSharedLookup(t *testing.T) {
	r := gateway(t)
	ctx := context.Background()

	subjectID := makeUser(t, r, "subject")
	round := &models.FeedbackRound{SubjectID: subjectID, CreatedByID: subjectID}
	if err := r.Rounds.Create(ctx, round); err != nil {
		t.Fatalf("create round: %v", err)
	}

	peerAvg := 4.5
	c := &models.Consolidation{
		RoundID:             round.ID,
		GeneratedByID:       subjectID,
		ExecutiveSummary:    "Strong quarter.",
		Strengths:           []string{"clarity", "follow-through"},
		AreasForImprovement: []string{"delegation"},
		QuestionSummaries:   map[string]string{"a": "continue mentoring"},
		SelfVsOthersDelta:   &models.SelfVsOthersDelta{SelfSubmitted: true, BlindSpots: []string{"impatience"}},
		CompetencyRatings:   []models.CompetencyRatingAggregate{{Key: "execution", Name: "Execution", PeerAverage: &peerAvg, OthersCount: 2}},
		ManagerOnlyChannel:  &models.ManagerOnlyChannel{NoteCount: 1, Synthesis: "watch burnout"},
	}
	if err := r.Consolidations.Create(ctx, c); err != nil {
		t.Fatalf("create consolidation: %v", err)
	}

	got, err := r.Consolidations.FindByRoundID(ctx, round.ID)
	if err != nil {
		t.Fatalf("find consolidation: %v", err)
	}
	if len(got.Strengths) != 2 || got.QuestionSummaries["a"] != "continue mentoring" {
		t.Fatalf("jsonb round-trip mismatch: %+v", got)
	}
	if got.SelfVsOthersDelta == nil || !got.SelfVsOthersDelta.SelfSubmitted {
		t.Fatalf("delta not round-tripped: %+v", got.SelfVsOthersDelta)
	}
	if got.CompetencyRatings[0].PeerAverage == nil || *got.CompetencyRatings[0].PeerAverage != 4.5 {
		t.Fatalf("competency pointer not round-tripped: %+v", got.CompetencyRatings)
	}
	if got.ManagerOnlyChannel == nil || got.ManagerOnlyChannel.Synthesis != "watch burnout" {
		t.Fatalf("manager channel not round-tripped: %+v", got.ManagerOnlyChannel)
	}

	// Not shared yet → not returned by the shared lookup.
	shared, _ := r.Consolidations.FindSharedBySubjectID(ctx, subjectID)
	if len(shared) != 0 {
		t.Fatalf("expected no shared consolidations, got %d", len(shared))
	}
}

func TestRounds_PaginationCoversEveryRowOnce(t *testing.T) {
	r := gateway(t)
	ctx := context.Background()
	subjectID := makeUser(t, r, "subject")

	// Insert 5 rounds that all share the same created_at, so ordering by
	// created_at alone would be ambiguous — the id tiebreaker must make paging
	// deterministic (no skips, no duplicates across page boundaries).
	for i := 0; i < 5; i++ {
		if _, err := testPool.Exec(ctx, `
			INSERT INTO feedback_rounds (subject_id, created_by_id, status, created_at, updated_at)
			VALUES ($1, $1, 'draft', '2026-01-01 00:00:00+00', '2026-01-01 00:00:00+00')`, subjectID); err != nil {
			t.Fatalf("insert round %d: %v", i, err)
		}
	}

	// Page through with a small page size and collect every id.
	seen := map[string]int{}
	total := 0
	for offset := 0; ; offset += 2 {
		page, err := r.Rounds.FindPaged(ctx, 2, offset)
		if err != nil {
			t.Fatalf("find paged (offset %d): %v", offset, err)
		}
		if len(page) == 0 {
			break
		}
		for _, rd := range page {
			seen[rd.ID]++
			total++
		}
	}

	if total != 5 {
		t.Fatalf("expected to page over exactly 5 rows, got %d", total)
	}
	if len(seen) != 5 {
		t.Fatalf("expected 5 distinct rows, got %d (a duplicate crossed a page boundary)", len(seen))
	}
	for id, n := range seen {
		if n != 1 {
			t.Fatalf("row %s appeared %d times across pages", id, n)
		}
	}
}

func TestTemplates_UpsertBySlug(t *testing.T) {
	r := gateway(t)
	ctx := context.Background()

	tmpl := &models.Template{
		Slug:         "default",
		Name:         "First",
		Questions:    []models.TemplateQuestion{{Key: "a", CardTitle: "Continue"}},
		Competencies: []models.TemplateCompetency{{Key: "execution", Name: "Execution"}},
	}
	if err := r.Templates.Upsert(ctx, tmpl); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	firstID := tmpl.ID

	// Upsert with the same slug updates in place (no new row).
	tmpl2 := &models.Template{Slug: "default", Name: "Renamed"}
	if err := r.Templates.Upsert(ctx, tmpl2); err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if tmpl2.ID != firstID {
		t.Fatalf("expected same id on slug conflict, got %s vs %s", tmpl2.ID, firstID)
	}

	got, err := r.Templates.FindBySlug(ctx, "default")
	if err != nil {
		t.Fatalf("find by slug: %v", err)
	}
	if got.Name != "Renamed" {
		t.Fatalf("expected updated name, got %q", got.Name)
	}
	all, _ := r.Templates.FindAll(ctx)
	if len(all) != 1 {
		t.Fatalf("expected exactly 1 template, got %d", len(all))
	}
}
