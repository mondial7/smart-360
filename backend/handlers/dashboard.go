package handlers

import (
	"net/http"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardStats struct {
	TotalUsers      int64 `json:"totalUsers"`
	AdminCount      int64 `json:"adminCount"`
	MemberCount     int64 `json:"memberCount"`
	TotalRounds     int64 `json:"totalRounds"`
	ActiveRounds    int64 `json:"activeRounds"`
	PendingReviews  int64 `json:"pendingReviews"`
	MyFeedbackCount int64 `json:"myFeedbackCount"`
}

func GetDashboardStats(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	db := database.GetDB()
	var stats DashboardStats

	// User counts (visible to all)
	db.Model(&models.User{}).Count(&stats.TotalUsers)
	db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&stats.AdminCount)
	db.Model(&models.User{}).Where("role = ?", models.RoleMember).Count(&stats.MemberCount)

	// Round counts
	db.Model(&models.FeedbackRound{}).Count(&stats.TotalRounds)
	db.Model(&models.FeedbackRound{}).Where("status = ?", models.RoundActive).Count(&stats.ActiveRounds)

	// Pending reviews for current user
	var pendingRounds []models.FeedbackRound
	db.Joins("JOIN round_reviewers ON round_reviewers.round_id = feedback_rounds.id").
		Where("round_reviewers.reviewer_id = ?", currentUser.ID).
		Where("feedback_rounds.status = ?", models.RoundActive).
		Where("feedback_rounds.deadline > ?", time.Now()).
		Find(&pendingRounds)

	// Check which ones don't have submissions
	for _, round := range pendingRounds {
		var count int64
		db.Model(&models.Submission{}).Where("round_id = ? AND reviewer_id = ?", round.ID, currentUser.ID).Count(&count)
		if count == 0 {
			stats.PendingReviews++
		}
	}

	// My feedback count (rounds shared with me as subject)
	db.Model(&models.FeedbackRound{}).Where("subject_id = ? AND status = ?", currentUser.ID, models.RoundShared).Count(&stats.MyFeedbackCount)

	c.JSON(http.StatusOK, stats)
}

func GetActiveRounds(c *gin.Context) {
	db := database.GetDB()

	var rounds []models.FeedbackRound
	db.Preload("Subject").Preload("Reviewers.Reviewer").
		Where("status = ?", models.RoundActive).
		Order("deadline ASC").
		Find(&rounds)

	c.JSON(http.StatusOK, rounds)
}

func GetPendingReviews(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	db := database.GetDB()

	// Find rounds where user is reviewer
	var rounds []models.FeedbackRound
	db.Joins("JOIN round_reviewers ON round_reviewers.round_id = feedback_rounds.id").
		Where("round_reviewers.reviewer_id = ?", currentUser.ID).
		Where("feedback_rounds.status = ?", models.RoundActive).
		Where("feedback_rounds.deadline > ?", time.Now()).
		Preload("Subject").
		Preload("CreatedBy").
		Find(&rounds)

	// Filter out submitted ones
	var pending []models.FeedbackRound
	for _, round := range rounds {
		var count int64
		db.Model(&models.Submission{}).Where("round_id = ? AND reviewer_id = ?", round.ID, currentUser.ID).Count(&count)
		if count == 0 {
			pending = append(pending, round)
		}
	}

	c.JSON(http.StatusOK, pending)
}
