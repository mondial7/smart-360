package models

import "time"

// Team is a group of users with a designated team admin. Membership is stored
// in the team_members join table (not an embedded array), so MemberIDs is
// populated by the repository on read, not persisted on the row.
type Team struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	TeamAdminID string    `json:"teamAdminId"`
	MemberIDs   []string  `json:"memberIds"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
