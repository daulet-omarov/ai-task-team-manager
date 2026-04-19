package attachment

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/uploader"
)

type taskGetter interface {
	GetByID(id uint) (*models.Task, error)
}

type boardMemberChecker interface {
	IsMember(boardID uint, userID int64) (bool, error)
}

type Service struct {
	repo      *Repository
	taskRepo  taskGetter
	boardRepo boardMemberChecker
}

func NewService(repo *Repository, taskRepo taskGetter, boardRepo boardMemberChecker) *Service {
	return &Service{repo: repo, taskRepo: taskRepo, boardRepo: boardRepo}
}

func (s *Service) Upload(taskID uint, userID int64, fh *multipart.FileHeader, r *http.Request) (*AttachmentResponse, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(task.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	path, err := uploader.SaveFile(fh)
	if err != nil {
		return nil, err
	}

	a := &models.Attachment{
		TaskID:   taskID,
		FilePath: path,
		FileName: fh.Filename,
		FileSize: int(fh.Size),
	}

	if err := s.repo.Create(a); err != nil {
		return nil, err
	}

	return toResponse(a, uploader.FullURL(r, path)), nil
}

func (s *Service) GetByTaskID(taskID uint, userID int64, r *http.Request) ([]*AttachmentResponse, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(task.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	attachments, err := s.repo.GetByTaskID(taskID)
	if err != nil {
		return nil, err
	}

	result := make([]*AttachmentResponse, 0, len(attachments))
	for _, a := range attachments {
		result = append(result, toResponse(a, uploader.FullURL(r, a.FilePath)))
	}
	return result, nil
}

func (s *Service) Delete(attachmentID uint, userID int64) error {
	a, err := s.repo.GetByID(attachmentID)
	if err != nil {
		return errors.New("attachment not found")
	}

	task, err := s.taskRepo.GetByID(a.TaskID)
	if err != nil {
		return errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(task.BoardID, userID)
	if err != nil || !isMember {
		return errors.New("access denied")
	}

	return s.repo.Delete(attachmentID)
}

func toResponse(a *models.Attachment, url string) *AttachmentResponse {
	return &AttachmentResponse{
		ID:        a.ID,
		TaskID:    a.TaskID,
		FileName:  a.FileName,
		FileSize:  a.FileSize,
		URL:       url,
		CreatedAt: a.CreatedAt,
	}
}
