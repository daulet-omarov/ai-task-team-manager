package task

import (
	"errors"
	"net/http"
	"time"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/uploader"
)

// boardMemberChecker abstracts board.Repository to check membership without a hard import cycle.
type boardMemberChecker interface {
	IsMember(boardID uint, userID int64) (bool, error)
}

type attachmentSaver interface {
	Create(a *models.Attachment) error
}

type attachmentFetcher interface {
	GetByTaskID(taskID uint) ([]*models.Attachment, error)
}

type CommentWithAuthor struct {
	ID             uint
	AuthorID       uint
	AuthorFullName string
	AuthorPhoto    string
	Content        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type commentFetcher interface {
	GetByTaskID(taskID uint) ([]*CommentWithAuthor, error)
}

type Service struct {
	repo            *Repository
	boardRepo       boardMemberChecker
	attachmentRepo  attachmentSaver
	attachmentFetch attachmentFetcher
	commentFetch    commentFetcher
}

func NewService(repo *Repository, boardRepo boardMemberChecker, attachmentRepo attachmentSaver, attachmentFetch attachmentFetcher, commentFetch commentFetcher) *Service {
	return &Service{repo: repo, boardRepo: boardRepo, attachmentRepo: attachmentRepo, attachmentFetch: attachmentFetch, commentFetch: commentFetch}
}

func (s *Service) Create(boardID uint, reporterUserID int64, req CreateTaskRequest, r *http.Request) (*TaskResponse, error) {
	isMember, err := s.boardRepo.IsMember(boardID, reporterUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("access denied: not a board member")
	}

	todoStatusID, err := s.repo.GetStatusIDByCode("to_do")
	if err != nil {
		return nil, err
	}

	t := &models.Task{
		BoardID:      boardID,
		Title:        req.Title,
		Description:  req.Description,
		StatusID:     todoStatusID,
		PriorityID:   req.PriorityID,
		DifficultyID: req.DifficultyID,
		DeveloperID:  req.AssigneeID,
		TesterID:     req.TesterID,
		ReporterID:   uint(reporterUserID),
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}

	resp := toResponse(t)

	for _, fh := range req.Files {
		path, err := uploader.SaveFile(fh)
		if err != nil {
			continue // skip invalid files, don't fail the whole request
		}
		a := &models.Attachment{
			TaskID:   t.ID,
			FilePath: path,
			FileName: fh.Filename,
			FileSize: int(fh.Size),
		}
		if err := s.attachmentRepo.Create(a); err != nil {
			continue
		}
		resp.Attachments = append(resp.Attachments, AttachmentInfo{
			ID:       a.ID,
			FileName: a.FileName,
			FileSize: a.FileSize,
			URL:      uploader.FullURL(r, path),
		})
	}

	// newly created task has no comments yet
	return resp, nil
}

func (s *Service) GetByID(id uint, userID int64, r *http.Request) (*TaskResponse, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(t.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	resp := toResponse(t)
	s.loadAttachments(resp, t.ID, r)
	s.loadComments(resp, t.ID)
	return resp, nil
}

func (s *Service) Delete(id uint, userID int64) error {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(t.BoardID, userID)
	if err != nil || !isMember {
		return errors.New("access denied")
	}

	return s.repo.Delete(id)
}

func (s *Service) Update(id uint, userID int64, req UpdateTaskRequest, r *http.Request) (*TaskResponse, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("task not found")
	}

	isMember, err := s.boardRepo.IsMember(t.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.StatusID != 0 {
		updates["status_id"] = req.StatusID
	}
	if req.PriorityID != 0 {
		updates["priority_id"] = req.PriorityID
	}
	if req.AssigneeID != 0 {
		updates["developer_id"] = req.AssigneeID
	}
	if req.TesterID != 0 {
		updates["tester_id"] = req.TesterID
	}
	if req.DifficultyID != nil {
		updates["difficulty_id"] = *req.DifficultyID
	}
	if req.TimeSpent != 0 {
		updates["time_spent"] = req.TimeSpent
	}

	if err := s.repo.Update(t.ID, updates); err != nil {
		return nil, err
	}

	// Reload to get fresh associations after the update.
	t, err = s.repo.GetByID(t.ID)
	if err != nil {
		return nil, err
	}

	resp := toResponse(t)
	s.loadAttachments(resp, t.ID, r)
	s.loadComments(resp, t.ID)
	return resp, nil
}

func (s *Service) GetByBoardID(boardID uint, userID int64, r *http.Request) ([]*TaskResponse, error) {
	isMember, err := s.boardRepo.IsMember(boardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	tasks, err := s.repo.GetByBoardID(boardID)
	if err != nil {
		return nil, err
	}

	result := make([]*TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp := toResponse(t)
		s.loadAttachments(resp, t.ID, r)
		s.loadComments(resp, t.ID)
		result = append(result, resp)
	}
	return result, nil
}

// loadAttachments fetches attachments for taskID and appends them to resp.
// Errors are silently ignored so that a missing attachment row never breaks the task response.
func (s *Service) loadAttachments(resp *TaskResponse, taskID uint, r *http.Request) {
	attachments, err := s.attachmentFetch.GetByTaskID(taskID)
	if err != nil {
		return
	}
	for _, a := range attachments {
		resp.Attachments = append(resp.Attachments, AttachmentInfo{
			ID:       a.ID,
			FileName: a.FileName,
			FileSize: a.FileSize,
			URL:      uploader.FullURL(r, a.FilePath),
		})
	}
}

// loadComments fetches comments for taskID and appends them to resp.
func (s *Service) loadComments(resp *TaskResponse, taskID uint) {
	comments, err := s.commentFetch.GetByTaskID(taskID)
	if err != nil {
		return
	}
	for _, c := range comments {
		resp.Comments = append(resp.Comments, CommentInfo{
			ID:             c.ID,
			AuthorID:       c.AuthorID,
			AuthorFullName: c.AuthorFullName,
			AuthorPhoto:    c.AuthorPhoto,
			Content:        c.Content,
			CreatedAt:      c.CreatedAt,
			UpdatedAt:      c.UpdatedAt,
		})
	}
}

func employeeInfo(e models.Employee) *EmployeeInfo {
	if e.ID == 0 {
		return nil
	}
	return &EmployeeInfo{
		ID:       e.ID,
		FullName: e.FullName,
		Photo:    e.Photo,
	}
}

func toResponse(t *models.Task) *TaskResponse {
	return &TaskResponse{
		ID:           t.ID,
		BoardID:      t.BoardID,
		Title:        t.Title,
		Description:  t.Description,
		StatusID:     t.StatusID,
		PriorityID:   t.PriorityID,
		DifficultyID: t.DifficultyID,
		AssigneeID:   t.DeveloperID,
		Assignee:     employeeInfo(t.Developer),
		TesterID:     t.TesterID,
		Tester:       employeeInfo(t.Tester),
		ReporterID:   t.ReporterID,
		Reporter:     employeeInfo(t.Reporter),
		TimeSpent:    t.TimeSpent,
		Attachments:  []AttachmentInfo{},
		Comments:     []CommentInfo{},
		CreatedAt:    t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
