package models

import (
	"encoding/json"
	"time"
)

type UserGamification struct {
	UserID         int64      `gorm:"primaryKey"`
	TotalPoints    int        `gorm:"not null;default:0"`
	CurrentLevel   int        `gorm:"not null;default:1"`
	CurrentStreak  int        `gorm:"not null;default:0"`
	LongestStreak  int        `gorm:"not null;default:0"`
	LastActiveDate *time.Time `gorm:"type:date"`
	UpdatedAt      time.Time
}

func (UserGamification) TableName() string { return "user_gamification" }

type PointTransaction struct {
	ID       uint            `gorm:"primaryKey"`
	UserID   int64           `gorm:"not null"`
	TaskID   *uint           `gorm:"index"`
	BoardID  *uint           `gorm:"index"`
	Points   int             `gorm:"not null"`
	Reason   string          `gorm:"not null;size:50"`
	EarnedAt time.Time       `gorm:"not null"`
	Metadata json.RawMessage `gorm:"type:jsonb"`
}

type Kudos struct {
	ID         uint      `gorm:"primaryKey"`
	FromUserID int64     `gorm:"not null"`
	ToUserID   int64     `gorm:"not null"`
	TaskID     *uint
	Message    string `gorm:"type:text"`
	CreatedAt  time.Time
}
