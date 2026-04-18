package task

import (
	"errors"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
)

// boardMemberChecker abstracts board.Repository to check membership without a hard import cycle.
type boardMemberChecker interface {
	IsMember(boardID uint, userID int64) (bool, error)
}

type Service struct {
	repo      *Repository
	boardRepo boardMemberChecker
}

func NewService(repo *Repository, boardRepo boardMemberChecker) *Service {
	return &Service{repo: repo, boardRepo: boardRepo}
}

func (s *Service) Create(boardID uint, reporterUserID int64, req CreateTaskRequest) (*TaskResponse, error) {
	isMember, err := s.boardRepo.IsMember(boardID, reporterUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("access denied: not a board member")
	}

	t := &models.Task{
		BoardID:     boardID,
		Title:       req.Title,
		Description: req.Description,
		PriorityID:  req.PriorityID,
		DeveloperID: req.AssigneeID,
		ReporterID:  uint(reporterUserID),
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}

	return toResponse(t), nil
}

func (s *Service) GetByID(id uint, userID int64) (*TaskResponse, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(t.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	return toResponse(t), nil
}

func (s *Service) Update(id uint, userID int64, req UpdateTaskRequest) (*TaskResponse, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(t.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	if req.Title != "" {
		t.Title = req.Title
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if req.StatusID != 0 {
		t.StatusID = req.StatusID
	}
	if req.PriorityID != 0 {
		t.PriorityID = req.PriorityID
	}
	if req.AssigneeID != 0 {
		t.DeveloperID = req.AssigneeID
	}
	if req.TimeSpent != 0 {
		t.TimeSpent = req.TimeSpent
	}

	if err := s.repo.Update(t); err != nil {
		return nil, err
	}

	return toResponse(t), nil
}

func toResponse(t *models.Task) *TaskResponse {
	return &TaskResponse{
		ID:          t.ID,
		BoardID:     t.BoardID,
		Title:       t.Title,
		Description: t.Description,
		StatusID:    t.StatusID,
		PriorityID:  t.PriorityID,
		AssigneeID:  t.DeveloperID,
		ReporterID:  t.ReporterID,
		TimeSpent:   t.TimeSpent,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
