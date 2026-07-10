// Package models holds the domain types persisted in Postgres. IDs are UUID
// strings; nullable columns use pointers or empty values as noted per field.
package models

import "time"

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleTeamAdmin UserRole = "team_admin"
	RoleMember    UserRole = "member"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleTeamAdmin, RoleMember:
		return true
	}
	return false
}

type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	PhotoURL  string     `json:"photoUrl"`
	Role      UserRole   `json:"role"`
	TeamID    *string    `json:"teamId,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`
}

func (u User) IsAdmin() bool     { return u.Role == RoleAdmin }
func (u User) IsTeamAdmin() bool { return u.Role == RoleTeamAdmin }
