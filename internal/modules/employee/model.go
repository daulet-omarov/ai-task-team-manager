package employee

import (
	"time"
)

type Employee struct {
	ID          uint `gorm:"primaryKey"`
	UserID      uint
	FullName    string `gorm:"not null"`
	Photo       string
	Email       string `gorm:"uniqueIndex;not null"`
	TeamID      uint
	Birthday    time.Time `gorm:"not null"`
	PhoneNumber string
	GenderID    uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Team        Team
	Gender      Gender
}

type Role struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string `gorm:"unique"`
}

type Gender struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string `gorm:"unique"`
}

type Team struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string `gorm:"unique"`
}
