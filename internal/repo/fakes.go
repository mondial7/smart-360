package repo

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mondial7/smart-360/internal/models"
)

// This file provides in-memory implementations of every repository interface,
// used by handler tests. They are intentionally simple: map-backed, mutex-
// guarded, with monotonic string IDs. They are not a performance model of
// Postgres, only a behavioural stand-in.

// ids hands out deterministic, unique string IDs for fakes.
type ids struct {
	mu sync.Mutex
	n  int
}

func (i *ids) next(prefix string) string {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.n++
	return fmt.Sprintf("%s-%d", prefix, i.n)
}

var fakeIDs = &ids{}

func now() time.Time { return time.Now().UTC() }

// NewFakes returns a Repositories bundle backed entirely by in-memory fakes.
func NewFakes() Repositories {
	return Repositories{
		Users:          NewFakeUsers(),
		Teams:          NewFakeTeams(),
		Rounds:         NewFakeRounds(),
		Submissions:    NewFakeSubmissions(),
		Templates:      NewFakeTemplates(),
		Consolidations: NewFakeConsolidations(),
		Audit:          NewFakeAudit(),
		Moderation:     NewFakeModeration(),
		Sessions:       NewFakeSessions(),
	}
}

// ---- Users ----

type FakeUsers struct {
	mu   sync.Mutex
	data map[string]models.User
}

func NewFakeUsers() *FakeUsers { return &FakeUsers{data: map[string]models.User{}} }

func (f *FakeUsers) FindByID(_ context.Context, id string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, ok := f.data[id]; ok {
		return &u, nil
	}
	return nil, ErrNotFound
}

func (f *FakeUsers) FindByEmail(_ context.Context, email string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.data {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, ErrNotFound
}

func (f *FakeUsers) Create(_ context.Context, u *models.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u.ID == "" {
		u.ID = fakeIDs.next("user")
	}
	if u.Role == "" {
		u.Role = models.RoleMember
	}
	u.CreatedAt, u.UpdatedAt = now(), now()
	f.data[u.ID] = *u
	return nil
}

func (f *FakeUsers) UpdateRole(_ context.Context, id string, role models.UserRole) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.data[id]
	if !ok {
		return ErrNotFound
	}
	u.Role = role
	u.UpdatedAt = now()
	f.data[id] = u
	return nil
}

func (f *FakeUsers) UpdateLastLogin(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.data[id]
	if !ok {
		return ErrNotFound
	}
	t := now()
	u.LastLogin = &t
	f.data[id] = u
	return nil
}

func (f *FakeUsers) SetTeam(_ context.Context, userID string, teamID *string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.data[userID]
	if !ok {
		return ErrNotFound
	}
	u.TeamID = teamID
	f.data[userID] = u
	return nil
}

func (f *FakeUsers) FindAll(_ context.Context) ([]models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]models.User, 0, len(f.data))
	for _, u := range f.data {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// ---- Teams ----

type FakeTeams struct {
	mu      sync.Mutex
	data    map[string]models.Team
	members map[string][]string // teamID -> userIDs
}

func NewFakeTeams() *FakeTeams {
	return &FakeTeams{data: map[string]models.Team{}, members: map[string][]string{}}
}

func (f *FakeTeams) FindByID(_ context.Context, id string) (*models.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	t.MemberIDs = append([]string(nil), f.members[id]...)
	return &t, nil
}

func (f *FakeTeams) FindAll(_ context.Context) ([]models.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]models.Team, 0, len(f.data))
	for id, t := range f.data {
		t.MemberIDs = append([]string(nil), f.members[id]...)
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *FakeTeams) Create(_ context.Context, t *models.Team) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t.ID == "" {
		t.ID = fakeIDs.next("team")
	}
	t.CreatedAt, t.UpdatedAt = now(), now()
	f.data[t.ID] = *t
	return nil
}

func (f *FakeTeams) Update(_ context.Context, t *models.Team) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.data[t.ID]; !ok {
		return ErrNotFound
	}
	cur := f.data[t.ID]
	cur.Name = t.Name
	cur.TeamAdminID = t.TeamAdminID
	cur.UpdatedAt = now()
	f.data[t.ID] = cur
	return nil
}

func (f *FakeTeams) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, id)
	delete(f.members, id)
	return nil
}

func (f *FakeTeams) AddMember(_ context.Context, teamID, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, m := range f.members[teamID] {
		if m == userID {
			return nil
		}
	}
	f.members[teamID] = append(f.members[teamID], userID)
	return nil
}

func (f *FakeTeams) RemoveMember(_ context.Context, teamID, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cur := f.members[teamID]
	out := cur[:0]
	for _, m := range cur {
		if m != userID {
			out = append(out, m)
		}
	}
	f.members[teamID] = out
	return nil
}

func (f *FakeTeams) GetMemberIDs(_ context.Context, teamID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.members[teamID]...), nil
}

// ---- Rounds ----

type FakeRounds struct {
	mu        sync.Mutex
	data      map[string]models.FeedbackRound
	reviewers map[string][]models.RoundReviewer // roundID -> reviewers
}

func NewFakeRounds() *FakeRounds {
	return &FakeRounds{data: map[string]models.FeedbackRound{}, reviewers: map[string][]models.RoundReviewer{}}
}

func (f *FakeRounds) hydrate(r models.FeedbackRound) models.FeedbackRound {
	r.Reviewers = append([]models.RoundReviewer(nil), f.reviewers[r.ID]...)
	return r
}

func (f *FakeRounds) FindByID(_ context.Context, id string) (*models.FeedbackRound, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	h := f.hydrate(r)
	return &h, nil
}

func (f *FakeRounds) filter(pred func(models.FeedbackRound) bool) []models.FeedbackRound {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.FeedbackRound
	for _, r := range f.data {
		if pred(r) {
			out = append(out, f.hydrate(r))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *FakeRounds) FindBySubjectID(_ context.Context, subjectID string) ([]models.FeedbackRound, error) {
	return f.filter(func(r models.FeedbackRound) bool { return r.SubjectID == subjectID }), nil
}

func (f *FakeRounds) FindByCreatedByID(_ context.Context, creatorID string) ([]models.FeedbackRound, error) {
	return f.filter(func(r models.FeedbackRound) bool { return r.CreatedByID == creatorID }), nil
}

func (f *FakeRounds) FindByReviewerID(_ context.Context, reviewerID string) ([]models.FeedbackRound, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.FeedbackRound
	for id, revs := range f.reviewers {
		for _, rev := range revs {
			if rev.ReviewerID == reviewerID {
				if r, ok := f.data[id]; ok {
					out = append(out, f.hydrate(r))
				}
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *FakeRounds) FindAll(_ context.Context) ([]models.FeedbackRound, error) {
	return f.filter(func(models.FeedbackRound) bool { return true }), nil
}

func (f *FakeRounds) Create(_ context.Context, r *models.FeedbackRound) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.ID == "" {
		r.ID = fakeIDs.next("round")
	}
	if r.Status == "" {
		r.Status = models.RoundDraft
	}
	r.CreatedAt, r.UpdatedAt = now(), now()
	f.data[r.ID] = *r
	return nil
}

func (f *FakeRounds) Update(_ context.Context, r *models.FeedbackRound) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cur, ok := f.data[r.ID]
	if !ok {
		return ErrNotFound
	}
	cur.SubjectID = r.SubjectID
	cur.TemplateID = r.TemplateID
	cur.Deadline = r.Deadline
	cur.Status = r.Status
	cur.UpdatedAt = now()
	f.data[r.ID] = cur
	return nil
}

func (f *FakeRounds) UpdateStatus(_ context.Context, id string, status models.RoundStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cur, ok := f.data[id]
	if !ok {
		return ErrNotFound
	}
	cur.Status = status
	cur.UpdatedAt = now()
	f.data[id] = cur
	return nil
}

func (f *FakeRounds) AddReviewer(_ context.Context, roundID string, reviewer models.RoundReviewer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.reviewers[roundID] {
		if r.ReviewerID == reviewer.ReviewerID {
			return nil
		}
	}
	if reviewer.ID == "" {
		reviewer.ID = fakeIDs.next("reviewer")
	}
	reviewer.RoundID = roundID
	reviewer.CreatedAt = now()
	f.reviewers[roundID] = append(f.reviewers[roundID], reviewer)
	return nil
}

func (f *FakeRounds) RemoveReviewer(_ context.Context, roundID, reviewerID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cur := f.reviewers[roundID]
	out := cur[:0]
	for _, r := range cur {
		if r.ReviewerID != reviewerID {
			out = append(out, r)
		}
	}
	f.reviewers[roundID] = out
	return nil
}

func (f *FakeRounds) GetReviewers(_ context.Context, roundID string) ([]models.RoundReviewer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]models.RoundReviewer(nil), f.reviewers[roundID]...), nil
}

// ---- Submissions ----

type FakeSubmissions struct {
	mu   sync.Mutex
	data map[string]models.Submission
}

func NewFakeSubmissions() *FakeSubmissions {
	return &FakeSubmissions{data: map[string]models.Submission{}}
}

func (f *FakeSubmissions) FindByRoundID(_ context.Context, roundID string) ([]models.Submission, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Submission
	for _, s := range f.data {
		if s.RoundID == roundID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *FakeSubmissions) FindByReviewerID(_ context.Context, reviewerID string) ([]models.Submission, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Submission
	for _, s := range f.data {
		if s.ReviewerID == reviewerID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *FakeSubmissions) FindByRoundAndReviewer(_ context.Context, roundID, reviewerID string) (*models.Submission, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.data {
		if s.RoundID == roundID && s.ReviewerID == reviewerID {
			return &s, nil
		}
	}
	return nil, ErrNotFound
}

func (f *FakeSubmissions) CountByRoundAndReviewer(_ context.Context, roundID, reviewerID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var n int64
	for _, s := range f.data {
		if s.RoundID == roundID && s.ReviewerID == reviewerID {
			n++
		}
	}
	return n, nil
}

func (f *FakeSubmissions) FindByID(_ context.Context, id string) (*models.Submission, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s, ok := f.data[id]; ok {
		return &s, nil
	}
	return nil, ErrNotFound
}

func (f *FakeSubmissions) Create(_ context.Context, s *models.Submission) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.ID == "" {
		s.ID = fakeIDs.next("submission")
	}
	s.SubmittedAt, s.UpdatedAt = now(), now()
	f.data[s.ID] = *s
	return nil
}

func (f *FakeSubmissions) Update(_ context.Context, s *models.Submission) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.data[s.ID]; !ok {
		return ErrNotFound
	}
	s.UpdatedAt = now()
	f.data[s.ID] = *s
	return nil
}

// ---- Templates ----

type FakeTemplates struct {
	mu   sync.Mutex
	data map[string]models.Template
}

func NewFakeTemplates() *FakeTemplates { return &FakeTemplates{data: map[string]models.Template{}} }

func (f *FakeTemplates) FindByID(_ context.Context, id string) (*models.Template, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.data[id]; ok {
		return &t, nil
	}
	return nil, ErrNotFound
}

func (f *FakeTemplates) FindBySlug(_ context.Context, slug string) (*models.Template, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, t := range f.data {
		if t.Slug == slug {
			return &t, nil
		}
	}
	return nil, ErrNotFound
}

func (f *FakeTemplates) FindAll(_ context.Context) ([]models.Template, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]models.Template, 0, len(f.data))
	for _, t := range f.data {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out, nil
}

func (f *FakeTemplates) Upsert(_ context.Context, t *models.Template) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, existing := range f.data {
		if existing.Slug == t.Slug {
			t.ID = id
			t.UpdatedAt = now()
			f.data[id] = *t
			return nil
		}
	}
	if t.ID == "" {
		t.ID = fakeIDs.next("template")
	}
	t.CreatedAt, t.UpdatedAt = now(), now()
	f.data[t.ID] = *t
	return nil
}

// ---- Consolidations ----

type FakeConsolidations struct {
	mu   sync.Mutex
	data map[string]models.Consolidation
}

func NewFakeConsolidations() *FakeConsolidations {
	return &FakeConsolidations{data: map[string]models.Consolidation{}}
}

func (f *FakeConsolidations) FindByRoundID(_ context.Context, roundID string) (*models.Consolidation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.data {
		if c.RoundID == roundID {
			return &c, nil
		}
	}
	return nil, ErrNotFound
}

func (f *FakeConsolidations) FindByID(_ context.Context, id string) (*models.Consolidation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if c, ok := f.data[id]; ok {
		return &c, nil
	}
	return nil, ErrNotFound
}

func (f *FakeConsolidations) FindSharedBySubjectID(_ context.Context, subjectID string) ([]models.Consolidation, error) {
	// Fakes don't join to rounds; tests that need this can set SharedAt and
	// filter by round ownership themselves. Returns shared consolidations only.
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.Consolidation
	for _, c := range f.data {
		if c.SharedAt != nil {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *FakeConsolidations) Create(_ context.Context, c *models.Consolidation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if c.ID == "" {
		c.ID = fakeIDs.next("consolidation")
	}
	c.CreatedAt, c.UpdatedAt = now(), now()
	f.data[c.ID] = *c
	return nil
}

func (f *FakeConsolidations) Update(_ context.Context, c *models.Consolidation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.data[c.ID]; !ok {
		return ErrNotFound
	}
	c.UpdatedAt = now()
	f.data[c.ID] = *c
	return nil
}

func (f *FakeConsolidations) UpdateNotes(_ context.Context, id, notes string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.data[id]
	if !ok {
		return ErrNotFound
	}
	c.AdminNotes = notes
	c.UpdatedAt = now()
	f.data[id] = c
	return nil
}

// ---- Audit ----

type FakeAudit struct {
	mu   sync.Mutex
	logs []models.AuditLog
}

func NewFakeAudit() *FakeAudit { return &FakeAudit{} }

func (f *FakeAudit) Create(_ context.Context, a *models.AuditLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if a.ID == "" {
		a.ID = fakeIDs.next("audit")
	}
	a.CreatedAt = now()
	f.logs = append(f.logs, *a)
	return nil
}

func (f *FakeAudit) FindAll(_ context.Context, limit int) ([]models.AuditLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := reversedAudit(f.logs)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (f *FakeAudit) FindByRoundID(_ context.Context, roundID string) ([]models.AuditLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var filtered []models.AuditLog
	for _, a := range f.logs {
		if a.RoundID != nil && *a.RoundID == roundID {
			filtered = append(filtered, a)
		}
	}
	return reversedAudit(filtered), nil
}

func reversedAudit(in []models.AuditLog) []models.AuditLog {
	out := make([]models.AuditLog, len(in))
	for i, a := range in {
		out[len(in)-1-i] = a
	}
	return out
}

// ---- Moderation ----

type FakeModeration struct {
	mu   sync.Mutex
	logs []models.ModerationLog
}

func NewFakeModeration() *FakeModeration { return &FakeModeration{} }

func (f *FakeModeration) Create(_ context.Context, m *models.ModerationLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m.ID == "" {
		m.ID = fakeIDs.next("moderation")
	}
	m.ModeratedAt = now()
	f.logs = append(f.logs, *m)
	return nil
}

func (f *FakeModeration) FindByRoundID(_ context.Context, roundID string) ([]models.ModerationLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []models.ModerationLog
	for _, m := range f.logs {
		if m.RoundID == roundID {
			out = append(out, m)
		}
	}
	return out, nil
}

// ---- Sessions ----

type FakeSessions struct {
	mu   sync.Mutex
	data map[string]models.Session
}

func NewFakeSessions() *FakeSessions { return &FakeSessions{data: map[string]models.Session{}} }

func (f *FakeSessions) Create(_ context.Context, s *models.Session) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.ID == "" {
		s.ID = fakeIDs.next("session")
	}
	s.CreatedAt = now()
	f.data[s.ID] = *s
	return nil
}

func (f *FakeSessions) FindByID(_ context.Context, id string) (*models.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s, ok := f.data[id]; ok {
		return &s, nil
	}
	return nil, ErrNotFound
}

func (f *FakeSessions) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, id)
	return nil
}

func (f *FakeSessions) DeleteExpired(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, s := range f.data {
		if s.ExpiresAt.Before(now()) {
			delete(f.data, id)
		}
	}
	return nil
}
