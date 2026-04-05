package board

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/sprint"
	"time"
)

type Board struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	SprintID  uint
	CreatedAt time.Time
	UpdatedAt time.Time
	Statuses  []Status `gorm:"many2many:board_statuses;"`
	Sprint    sprint.Sprint
}

type Status struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Code      string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
