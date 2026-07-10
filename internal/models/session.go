package models

import "time"

// Session is a server-side login session. The opaque ID is stored in an
// HttpOnly cookie; the row is deleted on logout or expiry.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}
