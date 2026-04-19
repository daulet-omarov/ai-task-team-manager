package models

import "time"

type Sprint struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `json:"name"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Board struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	OwnerID     int64 `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Members     []BoardMember `gorm:"foreignKey:BoardID"`
	Tasks       []Task        `gorm:"foreignKey:BoardID"`
	Statuses    []Status      `gorm:"many2many:board_statuses;"`
}

// BoardMember represents a user's membership in a board.
// Role can be "owner" or "member".
type BoardMember struct {
	ID       uint      `gorm:"primaryKey"`
	BoardID  uint      `gorm:"not null;index"`
	UserID   int64     `gorm:"not null;index"`
	Role     string    `gorm:"not null;default:'member'"`
	JoinedAt time.Time `gorm:"autoCreateTime"`
}

type Status struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Code      string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
