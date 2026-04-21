package comment

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

func (r *Repository) Create(c *models.Comment) error {
	return r.db.Create(c).Error
}

func (r *Repository) GetByID(id uint) (*models.Comment, error) {
	var c models.Comment
	err := r.db.First(&c, id).Error
	return &c, err
}

func (r *Repository) GetByTaskID(taskID uint) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := r.db.Where("task_id = ?", taskID).Order("created_at").Find(&comments).Error
	return comments, err
}

func (r *Repository) GetByTaskIDWithAuthor(taskID uint) ([]*CommentWithAuthor, error) {
	var rows []*CommentWithAuthor
	err := r.db.Raw(`
		SELECT c.id,
		       c.task_id,
		       c.author_id,
		       c.content,
		       c.created_at,
		       c.updated_at,
		       COALESCE(e.full_name, '') AS author_full_name,
		       COALESCE(e.photo, '')     AS author_photo
		FROM comments c
		LEFT JOIN employees e ON e.id = c.author_id
		WHERE c.task_id = ?
		ORDER BY c.created_at
	`, taskID).Scan(&rows).Error
	return rows, err
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&models.Comment{}, id).Error
}
