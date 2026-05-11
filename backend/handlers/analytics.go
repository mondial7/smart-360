package handlers

import (
	"context"
	"net/http"
	"smart360/database"
	"smart360/models"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoundAnalytics struct {
	RoundID             string    `json:"roundId"`
	CreatedAt           time.Time `json:"createdAt"`
	SharedAt            time.Time `json:"sharedAt"`
	StrengthsCount      int       `json:"strengthsCount"`
	ImprovementsCount   int       `json:"improvementsCount"`
	InsightsCount       int       `json:"insightsCount"`
	BehaviorsHasSummary bool      `json:"behaviorsHasSummary"`
	GrowthHasSummary    bool      `json:"growthHasSummary"`
}

type RadarAxes struct {
	Strengths    int `json:"strengths"`
	Improvements int `json:"improvements"`
	Behaviors    int `json:"behaviors"`
	Growth       int `json:"growth"`
}

type MyAnalyticsResponse struct {
	FeedbackReceivedCount int              `json:"feedbackReceivedCount"`
	FeedbackGivenCount    int              `json:"feedbackGivenCount"`
	PendingReviewsCount   int              `json:"pendingReviewsCount"`
	Rounds                []RoundAnalytics `json:"rounds"`
	LatestRadar           RadarAxes        `json:"latestRadar"`
}

func GetMyAnalytics(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	db := database.GetDB()
	ctx := context.Background()

	feedbackGiven, _ := db.Collection("submissions").CountDocuments(ctx, bson.M{"reviewer_id": currentUser.ID})

	assignedTotal, _ := db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{
		"reviewers.reviewer_id": currentUser.ID,
		"status":                bson.M{"$in": []models.RoundStatus{models.RoundActive, models.RoundClosed, models.RoundShared}},
	})
	pending := assignedTotal - feedbackGiven
	if pending < 0 {
		pending = 0
	}

	roundsCursor, err := db.Collection("feedback_rounds").Find(ctx, bson.M{
		"subject_id": currentUser.ID,
		"status":     models.RoundShared,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
		return
	}
	defer roundsCursor.Close(ctx)

	var sharedRounds []models.FeedbackRound
	if err := roundsCursor.All(ctx, &sharedRounds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode rounds"})
		return
	}

	var roundIDs []primitive.ObjectID
	roundByID := make(map[primitive.ObjectID]models.FeedbackRound, len(sharedRounds))
	for _, r := range sharedRounds {
		roundIDs = append(roundIDs, r.ID)
		roundByID[r.ID] = r
	}

	rounds := []RoundAnalytics{}
	if len(roundIDs) > 0 {
		consCursor, err := db.Collection("consolidations").Find(ctx, bson.M{
			"round_id":  bson.M{"$in": roundIDs},
			"shared_at": bson.M{"$ne": nil},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consolidations"})
			return
		}
		defer consCursor.Close(ctx)

		var consolidations []models.Consolidation
		if err := consCursor.All(ctx, &consolidations); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode consolidations"})
			return
		}

		for _, cons := range consolidations {
			round := roundByID[cons.RoundID]
			summaries := parseJSONStringMap(cons.QuestionSummaries)
			ra := RoundAnalytics{
				RoundID:             cons.RoundID.Hex(),
				CreatedAt:           round.CreatedAt,
				StrengthsCount:      len(parseJSONList(cons.Strengths)),
				ImprovementsCount:   len(parseJSONList(cons.AreasForImprovement)),
				InsightsCount:       len(parseJSONList(cons.ActionableInsights)),
				BehaviorsHasSummary: hasNonEmpty(summaries, "c"),
				GrowthHasSummary:    hasNonEmpty(summaries, "d"),
			}
			if cons.SharedAt != nil {
				ra.SharedAt = *cons.SharedAt
			}
			rounds = append(rounds, ra)
		}
	}

	sort.Slice(rounds, func(i, j int) bool {
		return rounds[i].SharedAt.Before(rounds[j].SharedAt)
	})

	radar := RadarAxes{}
	if n := len(rounds); n > 0 {
		latest := rounds[n-1]
		radar = RadarAxes{
			Strengths:    latest.StrengthsCount,
			Improvements: latest.ImprovementsCount,
			Behaviors:    boolToCount(latest.BehaviorsHasSummary),
			Growth:       boolToCount(latest.GrowthHasSummary),
		}
	}

	c.JSON(http.StatusOK, MyAnalyticsResponse{
		FeedbackReceivedCount: len(rounds),
		FeedbackGivenCount:    int(feedbackGiven),
		PendingReviewsCount:   int(pending),
		Rounds:                rounds,
		LatestRadar:           radar,
	})
}

func hasNonEmpty(m map[string]string, key string) bool {
	if m == nil {
		return false
	}
	v, ok := m[key]
	return ok && len(v) > 0
}

func boolToCount(b bool) int {
	if b {
		return 1
	}
	return 0
}
