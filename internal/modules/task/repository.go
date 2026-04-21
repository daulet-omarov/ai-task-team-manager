package task

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetStatusIDByCode(code string) (uint, error) {
	var status models.Status
	err := r.db.Where("code = ?", code).First(&status).Error
	return status.ID, err
}

func (r *Repository) Create(t *models.Task) error {
	cols := []string{"BoardID", "Title", "Description", "StatusID", "ReporterID", "TimeSpent"}
	if t.PriorityID != 0 {
		cols = append(cols, "PriorityID")
	}
	if t.DeveloperID != 0 {
		cols = append(cols, "DeveloperID")
	}
	if t.TesterID != 0 {
		cols = append(cols, "TesterID")
	}
	if t.DifficultyID != nil {
		cols = append(cols, "DifficultyID")
	}
	return r.db.Select(cols).Create(t).Error
}

func (r *Repository) GetByID(id uint) (*models.Task, error) {
	var t models.Task
	err := r.db.Preload("Developer").Preload("Tester").Preload("Reporter").First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&models.Task{ID: id}).Updates(fields).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&models.Task{}, id).Error
}

func (r *Repository) GetByBoardID(boardID uint) ([]*models.Task, error) {
	var tasks []*models.Task
	err := r.db.Preload("Developer").Preload("Tester").Preload("Reporter").
		Where("board_id = ?", boardID).Order("created_at").Find(&tasks).Error
	return tasks, err
}
