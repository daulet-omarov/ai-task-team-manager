package task

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/hub"
	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"gorm.io/gorm"
)

// localAttachmentRepo is a minimal local implementation that avoids an import
// cycle with the attachment package while satisfying both attachmentSaver and
// attachmentFetcher.
type localAttachmentRepo struct{ db *gorm.DB }

func (r *localAttachmentRepo) Create(a *models.Attachment) error {
	return r.db.Create(a).Error
}

func (r *localAttachmentRepo) GetByTaskID(taskID uint) ([]*models.Attachment, error) {
	var attachments []*models.Attachment
	err := r.db.Where("task_id = ?", taskID).Order("created_at").Find(&attachments).Error
	return attachments, err
}

// localCommentRepo satisfies commentFetcher without importing the comment package.
type localCommentRepo struct{ db *gorm.DB }

func (r *localCommentRepo) GetByTaskID(taskID uint) ([]*CommentWithAuthor, error) {
	var rows []*CommentWithAuthor
	err := r.db.Raw(`
		SELECT c.id,
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

func NewModule(db *gorm.DB, h *hub.Hub) *Handler {
	repo := NewRepository(db)
	boardRepo := board.NewRepository(db)
	attachRepo := &localAttachmentRepo{db}
	commentRepo := &localCommentRepo{db}
	service := NewService(repo, boardRepo, attachRepo, attachRepo, commentRepo)
	return NewHandler(service, h)
}
