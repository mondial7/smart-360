package models

import (
	"time"
)

type RoundStatus string

const (
	RoundDraft  RoundStatus = "draft"
	RoundActive RoundStatus = "active"
	RoundClosed RoundStatus = "closed"
	RoundShared RoundStatus = "shared"
)

type FeedbackRound struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	SubjectID   uint            `gorm:"not null" json:"subjectId"`
	Subject     User            `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	CreatedByID uint            `gorm:"not null" json:"createdById"`
	CreatedBy   User            `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
	Deadline    *time.Time      `json:"deadline"`
	Status      RoundStatus     `gorm:"default:draft" json:"status"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	Reviewers   []RoundReviewer `gorm:"foreignKey:RoundID" json:"reviewers,omitempty"`
}

func (FeedbackRound) TableName() string {
	return "feedback_rounds"
}

type RoundReviewer struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RoundID    uint      `gorm:"not null" json:"roundId"`
	ReviewerID uint      `gorm:"not null" json:"reviewerId"`
	Reviewer   User      `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (RoundReviewer) TableName() string {
	return "round_reviewers"
}

type Submission struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	RoundID     uint          `gorm:"not null" json:"roundId"`
	Round       FeedbackRound `gorm:"foreignKey:RoundID" json:"round,omitempty"`
	ReviewerID  uint          `gorm:"not null" json:"reviewerId"`
	Responses   string        `gorm:"type:text" json:"responses"` // JSON string
	SubmittedAt time.Time     `json:"submittedAt"`
}

func (Submission) TableName() string {
	return "submissions"
}
