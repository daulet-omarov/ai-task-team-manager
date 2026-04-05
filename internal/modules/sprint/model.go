package sprint

import "time"

type Sprint struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `json:"name"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
