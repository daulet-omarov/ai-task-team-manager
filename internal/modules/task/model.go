package task

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"time"
)

type Task struct {
	ID           uint   `gorm:"primaryKey"`
	Title        string `gorm:"not null"`
	StatusID     uint
	PriorityID   uint
	DifficultyID uint
	BoardID      uint
	DeveloperID  uint
	TesterID     uint
	ReporterID   uint
	Description  string `gorm:"type:text"`
	TimeSpent    uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Status       board.Status
	Priority     Priority
	Difficulty   Difficulty
	Board        board.Board
	Developer    employee.Employee
	Tester       employee.Employee
	Reporter     employee.Employee
	Types        []Type `gorm:"many2many:task_types;"`
}

type Priority struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string `gorm:"unique"`
}

type Difficulty struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string `gorm:"unique"`
}

type Type struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Code string `gorm:"unique"`
}

type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	Content   string `gorm:"type:text"`
	TaskID    uint
	AuthorID  uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Attachment struct {
	ID        uint `gorm:"primaryKey"`
	TaskID    uint
	FilePath  string
	FileName  string
	FileSize  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
