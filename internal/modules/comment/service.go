package comment

import (
	"errors"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
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

func (s *Service) Create(taskID uint, userID int64, req CreateCommentRequest) (*CommentResponse, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(task.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	c := &models.Comment{
		TaskID:   taskID,
		AuthorID: uint(userID),
		Content:  req.Content,
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}

	return toResponse(c), nil
}

func (s *Service) GetByTaskID(taskID uint, userID int64) ([]*CommentResponse, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(task.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	comments, err := s.repo.GetByTaskID(taskID)
	if err != nil {
		return nil, err
	}

	result := make([]*CommentResponse, 0, len(comments))
	for _, c := range comments {
		result = append(result, toResponse(c))
	}
	return result, nil
}

func (s *Service) Delete(commentID uint, userID int64) error {
	c, err := s.repo.GetByID(commentID)
	if err != nil {
		return errors.New("comment not found")
	}

	if c.AuthorID != uint(userID) {
		return errors.New("access denied")
	}

	return s.repo.Delete(commentID)
}

func toResponse(c *models.Comment) *CommentResponse {
	return &CommentResponse{
		ID:        c.ID,
		TaskID:    c.TaskID,
		AuthorID:  c.AuthorID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
