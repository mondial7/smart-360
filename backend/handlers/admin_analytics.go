package handlers

import (
	"context"
	"net/http"
	"smart360/database"
	"smart360/models"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminTotals struct {
	Users               int64   `json:"users"`
	Teams               int64   `json:"teams"`
	Rounds              int64   `json:"rounds"`
	Submissions         int64   `json:"submissions"`
	ConsolidationsShared int64  `json:"consolidationsShared"`
	CompletionRate      float64 `json:"completionRate"`
	AvgResponseSeconds  float64 `json:"avgResponseSeconds"`
}

type RoundsByStatus struct {
	Draft  int64 `json:"draft"`
	Active int64 `json:"active"`
	Closed int64 `json:"closed"`
	Shared int64 `json:"shared"`
}

type CompletionPoint struct {
	RoundID        string    `json:"roundId"`
	SubjectName    string    `json:"subjectName"`
	CreatedAt      time.Time `json:"createdAt"`
	Expected       int       `json:"expected"`
	Received       int       `json:"received"`
	CompletionRate float64   `json:"completionRate"`
	Status         string    `json:"status"`
}

type TeamActivity struct {
	TeamID             string  `json:"teamId"`
	TeamName           string  `json:"teamName"`
	MemberCount        int     `json:"memberCount"`
	ActiveRounds       int     `json:"activeRounds"`
	TotalSubmissions   int     `json:"totalSubmissions"`
	AvgResponseSeconds float64 `json:"avgResponseSeconds"`
}

type ThemePhrase struct {
	Phrase string `json:"phrase"`
	Count  int    `json:"count"`
}

type AdminThemes struct {
	Strengths    []ThemePhrase `json:"strengths"`
	Improvements []ThemePhrase `json:"improvements"`
}

type AdminAnalyticsResponse struct {
	Totals          AdminTotals       `json:"totals"`
	RoundsByStatus  RoundsByStatus    `json:"roundsByStatus"`
	CompletionTrend []CompletionPoint `json:"completionTrend"`
	TeamActivity    []TeamActivity    `json:"teamActivity"`
	TopThemes       AdminThemes       `json:"topThemes"`
}

func GetAdminAnalytics(c *gin.Context) {
	db := database.GetDB()
	ctx := context.Background()

	totalUsers, _ := db.Collection("users").CountDocuments(ctx, bson.M{})
	totalTeams, _ := db.Collection("teams").CountDocuments(ctx, bson.M{})
	totalRounds, _ := db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{})
	totalSubmissions, _ := db.Collection("submissions").CountDocuments(ctx, bson.M{})

	consolidationsShared, _ := db.Collection("consolidations").CountDocuments(ctx, bson.M{
		"shared_at": bson.M{"$ne": nil},
	})

	statusCounts := RoundsByStatus{}
	statusCounts.Draft, _ = db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{"status": models.RoundDraft})
	statusCounts.Active, _ = db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{"status": models.RoundActive})
	statusCounts.Closed, _ = db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{"status": models.RoundClosed})
	statusCounts.Shared, _ = db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{"status": models.RoundShared})

	rounds, err := loadRoundsForAnalytics(ctx, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load rounds"})
		return
	}

	subjectNames := loadUserNames(ctx, db, collectSubjectIDs(rounds))

	trend, expectedTotal, receivedTotal, totalResponseSeconds, responseSamples := buildCompletionTrend(ctx, db, rounds, subjectNames)

	completionRate := 0.0
	if expectedTotal > 0 {
		completionRate = float64(receivedTotal) / float64(expectedTotal)
	}

	avgResponseSeconds := 0.0
	if responseSamples > 0 {
		avgResponseSeconds = totalResponseSeconds / float64(responseSamples)
	}

	teamActivity, err := buildTeamActivity(ctx, db, rounds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build team activity"})
		return
	}

	themes := buildTopThemes(ctx, db)

	c.JSON(http.StatusOK, AdminAnalyticsResponse{
		Totals: AdminTotals{
			Users:                totalUsers,
			Teams:                totalTeams,
			Rounds:               totalRounds,
			Submissions:          totalSubmissions,
			ConsolidationsShared: consolidationsShared,
			CompletionRate:       completionRate,
			AvgResponseSeconds:   avgResponseSeconds,
		},
		RoundsByStatus:  statusCounts,
		CompletionTrend: trend,
		TeamActivity:    teamActivity,
		TopThemes:       themes,
	})
}

func loadRoundsForAnalytics(ctx context.Context, db *mongo.Database) ([]models.FeedbackRound, error) {
	cursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var rounds []models.FeedbackRound
	if err := cursor.All(ctx, &rounds); err != nil {
		return nil, err
	}
	return rounds, nil
}

func collectSubjectIDs(rounds []models.FeedbackRound) []primitive.ObjectID {
	seen := map[primitive.ObjectID]struct{}{}
	ids := []primitive.ObjectID{}
	for _, r := range rounds {
		if _, ok := seen[r.SubjectID]; ok {
			continue
		}
		seen[r.SubjectID] = struct{}{}
		ids = append(ids, r.SubjectID)
	}
	return ids
}

func loadUserNames(ctx context.Context, db *mongo.Database, ids []primitive.ObjectID) map[primitive.ObjectID]string {
	names := map[primitive.ObjectID]string{}
	if len(ids) == 0 {
		return names
	}
	cursor, err := db.Collection("users").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return names
	}
	defer cursor.Close(ctx)
	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return names
	}
	for _, u := range users {
		names[u.ID] = u.Name
	}
	return names
}

const completionTrendLimit = 12

func buildCompletionTrend(
	ctx context.Context,
	db *mongo.Database,
	rounds []models.FeedbackRound,
	subjectNames map[primitive.ObjectID]string,
) (points []CompletionPoint, expectedTotal, receivedTotal int, totalResponseSeconds float64, responseSamples int) {
	sort.Slice(rounds, func(i, j int) bool {
		return rounds[i].CreatedAt.After(rounds[j].CreatedAt)
	})

	considered := rounds
	if len(considered) > completionTrendLimit {
		considered = considered[:completionTrendLimit]
	}

	for _, r := range considered {
		if r.Status == models.RoundDraft {
			continue
		}
		expected := len(r.Reviewers)
		received := 0
		var subs []models.Submission
		if cursor, err := db.Collection("submissions").Find(ctx, bson.M{"round_id": r.ID}); err == nil {
			cursor.All(ctx, &subs)
			cursor.Close(ctx)
			received = len(subs)
		}
		rate := 0.0
		if expected > 0 {
			rate = float64(received) / float64(expected)
		}

		for _, s := range subs {
			delta := s.SubmittedAt.Sub(r.CreatedAt).Seconds()
			if delta > 0 {
				totalResponseSeconds += delta
				responseSamples++
			}
		}

		expectedTotal += expected
		receivedTotal += received

		points = append(points, CompletionPoint{
			RoundID:        r.ID.Hex(),
			SubjectName:    subjectNames[r.SubjectID],
			CreatedAt:      r.CreatedAt,
			Expected:       expected,
			Received:       received,
			CompletionRate: rate,
			Status:         string(r.Status),
		})
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].CreatedAt.Before(points[j].CreatedAt)
	})
	return
}

func buildTeamActivity(ctx context.Context, db *mongo.Database, rounds []models.FeedbackRound) ([]TeamActivity, error) {
	cursor, err := db.Collection("teams").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var teams []models.Team
	if err := cursor.All(ctx, &teams); err != nil {
		return nil, err
	}

	activity := make([]TeamActivity, 0, len(teams))
	for _, team := range teams {
		memberSet := map[primitive.ObjectID]struct{}{}
		for _, id := range team.MemberIDs {
			memberSet[id] = struct{}{}
		}
		if !team.TeamAdminID.IsZero() {
			memberSet[team.TeamAdminID] = struct{}{}
		}

		activeRounds := 0
		totalSubmissions := 0
		var totalResponseSeconds float64
		responseSamples := 0

		for _, r := range rounds {
			if _, ok := memberSet[r.SubjectID]; !ok {
				continue
			}
			if r.Status == models.RoundActive {
				activeRounds++
			}
			cursor, err := db.Collection("submissions").Find(ctx, bson.M{"round_id": r.ID})
			if err != nil {
				continue
			}
			var subs []models.Submission
			cursor.All(ctx, &subs)
			cursor.Close(ctx)
			totalSubmissions += len(subs)
			for _, s := range subs {
				delta := s.SubmittedAt.Sub(r.CreatedAt).Seconds()
				if delta > 0 {
					totalResponseSeconds += delta
					responseSamples++
				}
			}
		}

		avgResponse := 0.0
		if responseSamples > 0 {
			avgResponse = totalResponseSeconds / float64(responseSamples)
		}

		activity = append(activity, TeamActivity{
			TeamID:             team.ID.Hex(),
			TeamName:           team.Name,
			MemberCount:        len(memberSet),
			ActiveRounds:       activeRounds,
			TotalSubmissions:   totalSubmissions,
			AvgResponseSeconds: avgResponse,
		})
	}

	sort.Slice(activity, func(i, j int) bool {
		return activity[i].TotalSubmissions > activity[j].TotalSubmissions
	})
	return activity, nil
}

func buildTopThemes(ctx context.Context, db *mongo.Database) AdminThemes {
	cursor, err := db.Collection("consolidations").Find(ctx, bson.M{
		"shared_at": bson.M{"$ne": nil},
	})
	if err != nil {
		return AdminThemes{}
	}
	defer cursor.Close(ctx)
	var consolidations []models.Consolidation
	if err := cursor.All(ctx, &consolidations); err != nil {
		return AdminThemes{}
	}

	strengths := map[string]int{}
	improvements := map[string]int{}
	for _, cons := range consolidations {
		for _, item := range parseJSONList(cons.Strengths) {
			tallyTokens(strengths, item)
		}
		for _, item := range parseJSONList(cons.AreasForImprovement) {
			tallyTokens(improvements, item)
		}
	}

	return AdminThemes{
		Strengths:    topPhrases(strengths, 10),
		Improvements: topPhrases(improvements, 10),
	}
}

var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "for": {}, "with": {}, "that": {}, "this": {},
	"are": {}, "but": {}, "you": {}, "your": {}, "their": {}, "they": {},
	"have": {}, "has": {}, "from": {}, "into": {}, "more": {}, "than": {},
	"can": {}, "all": {}, "any": {}, "out": {}, "not": {}, "was": {}, "were": {},
	"about": {}, "very": {}, "such": {}, "only": {}, "when": {}, "what": {},
	"who": {}, "how": {}, "his": {}, "her": {}, "him": {}, "she": {},
	"strong": {}, "good": {}, "great": {}, "people": {}, "person": {}, "team": {},
	"work": {}, "shows": {}, "show": {}, "well": {}, "able": {}, "tends": {},
}

func tallyTokens(counts map[string]int, raw string) {
	for _, token := range tokenize(raw) {
		if _, skip := stopwords[token]; skip {
			continue
		}
		counts[token]++
	}
}

func tokenize(raw string) []string {
	cleaned := strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' {
			return r
		}
		return ' '
	}, raw)
	tokens := strings.Fields(cleaned)
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		t = strings.Trim(t, "-")
		if len(t) <= 3 {
			continue
		}
		out = append(out, t)
	}
	return out
}

func topPhrases(counts map[string]int, limit int) []ThemePhrase {
	phrases := make([]ThemePhrase, 0, len(counts))
	for phrase, count := range counts {
		phrases = append(phrases, ThemePhrase{Phrase: phrase, Count: count})
	}
	sort.Slice(phrases, func(i, j int) bool {
		if phrases[i].Count != phrases[j].Count {
			return phrases[i].Count > phrases[j].Count
		}
		return phrases[i].Phrase < phrases[j].Phrase
	})
	if len(phrases) > limit {
		phrases = phrases[:limit]
	}
	return phrases
}
