package models

import "time"

const (
	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusRejected = "rejected"
)

// Invite represents a board invitation sent from one user to another.
type Invite struct {
	ID        uint   `gorm:"primaryKey"`
	BoardID   uint   `gorm:"not null;index"`
	InviterID int64  `gorm:"not null"`
	InviteeID int64  `gorm:"not null;index"`
	Status    string `gorm:"not null;default:'pending'"` // pending | accepted | rejected
	CreatedAt time.Time
	UpdatedAt time.Time
	Board     Board
}
