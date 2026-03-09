package models

import (
	"time"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleMember UserRole = "member"
)

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Email      string    `gorm:"uniqueIndex;not null" json:"email"`
	Name       string    `json:"name"`
	PhotoURL   string    `json:"photoUrl"`
	Role       UserRole  `gorm:"default:member" json:"role"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	LastLogin  *time.Time `json:"lastLogin"`
}

func (User) TableName() string {
	return "users"
}
