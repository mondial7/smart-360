package models

import (
	"time"
)

type Consolidation struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	RoundID           uint      `gorm:"not null;uniqueIndex" json:"roundId"`
	Round             FeedbackRound `gorm:"foreignKey:RoundID" json:"round,omitempty"`
	GeneratedByID     uint      `gorm:"not null" json:"generatedById"`
	GeneratedBy       User      `gorm:"foreignKey:GeneratedByID" json:"generatedBy,omitempty"`
	ExecutiveSummary  string    `gorm:"type:text" json:"executiveSummary"`
	Strengths         string    `gorm:"type:text" json:"strengths"`          // JSON array
	AreasForImprovement string  `gorm:"type:text" json:"areasForImprovement"` // JSON array
	ActionableInsights string   `gorm:"type:text" json:"actionableInsights"`  // JSON array
	QuestionSummaries string   `gorm:"type:text" json:"questionSummaries"`    // JSON object
	AdminNotes        string    `gorm:"type:text" json:"adminNotes"`
	SharedAt          *time.Time `json:"sharedAt"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (Consolidation) TableName() string {
	return "consolidations"
}
