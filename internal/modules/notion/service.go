package notion

import (
	"errors"
	"fmt"

	"github.com/daulet-omarov/ai-task-team-manager/internal/logger"
	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/task"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/uploader"
	"go.uber.org/zap"
)

type attachmentCreator interface {
	Create(a *models.Attachment) error
}

type Service struct {
	boardRepo      *board.Repository
	taskRepo       *task.Repository
	attachmentRepo attachmentCreator
	employeeRepo   *employee.Repository
}

func NewService(boardRepo *board.Repository, taskRepo *task.Repository, attachmentRepo attachmentCreator, employeeRepo *employee.Repository) *Service {
	return &Service{boardRepo: boardRepo, taskRepo: taskRepo, attachmentRepo: attachmentRepo, employeeRepo: employeeRepo}
}

func (s *Service) Import(userID int64, req ImportRequest) (*ImportResult, error) {
	if req.Token == "" {
		return nil, errors.New("notion token is required")
	}
	if req.DatabaseID == "" {
		return nil, errors.New("database_id is required")
	}

	client := newClient(req.Token)

	dbTitle, schemaStatuses, err := client.getDatabase(req.DatabaseID)
	if err != nil {
		return nil, fmt.Errorf("notion API error: %w", err)
	}
	if dbTitle == "" {
		dbTitle = "Imported from Notion"
	}

	boardID := req.BoardID
	if boardID == 0 {
		b := &models.Board{
			Name:    dbTitle,
			OwnerID: userID,
		}
		if err := s.boardRepo.Create(b); err != nil {
			return nil, err
		}
		if err := s.boardRepo.AddMember(b.ID, userID, "owner"); err != nil {
			return nil, err
		}
		boardID = b.ID
	} else {
		isMember, err := s.boardRepo.IsMember(boardID, userID)
		if err != nil || !isMember {
			return nil, errors.New("access denied: not a board member")
		}
	}

	// statusID resolves a Notion status label to a DB status ID.
	// Results are cached so each unique label hits the DB only once.
	statusCache := map[string]uint{}
	statusID := func(label string) (uint, error) {
		if label == "" {
			label = "Not started"
		}
		if id, ok := statusCache[label]; ok {
			return id, nil
		}
		code := codeFromName(label)
		sr, err := s.boardRepo.UpsertStatus(boardID, label, code, "")
		if err != nil {
			return 0, err
		}
		statusCache[label] = sr.StatusID
		return sr.StatusID, nil
	}

	// Pre-create all statuses found in the Notion database schema so that
	// statuses with no tasks are still present on the board.
	for _, label := range schemaStatuses {
		if _, err := statusID(label); err != nil {
			logger.Log.Warn("notion import: failed to pre-create schema status",
				zap.String("label", label),
				zap.Error(err),
			)
		}
	}

	// developerID resolves an email to an employee ID and ensures the employee
	// is a member of the board (adds them as "member" if they are not).
	// Results are cached; unknown emails return 0 (no assignee).
	developerCache := map[string]uint{}
	developerID := func(email string) uint {
		if email == "" {
			return 0
		}
		if id, ok := developerCache[email]; ok {
			return id
		}
		emp, err := s.employeeRepo.GetByEmail(email)
		if err != nil {
			// Employee not found or DB error — skip assignment silently.
			developerCache[email] = 0
			return 0
		}
		// Ensure the employee's user account is a board member.
		isMember, err := s.boardRepo.IsMember(boardID, int64(emp.UserID))
		if err == nil && !isMember {
			if addErr := s.boardRepo.AddMember(boardID, int64(emp.UserID), "member"); addErr != nil {
				logger.Log.Warn("notion import: failed to add assignee as board member",
					zap.String("email", email),
					zap.Error(addErr),
				)
			}
		}
		developerCache[email] = emp.ID
		return emp.ID
	}

	pages, err := client.queryDatabase(req.DatabaseID)
	if err != nil {
		return nil, fmt.Errorf("notion query error: %w", err)
	}

	result := &ImportResult{
		BoardID:   boardID,
		BoardName: dbTitle,
		Errors:    []string{},
	}

	for _, p := range pages {
		title := extractTitle(p)
		if title == "" {
			result.Skipped++
			continue
		}

		rawStatus := extractRawStatus(p)
		logger.Log.Info("notion import: page status",
			zap.String("page_id", p.ID),
			zap.String("title", title),
			zap.String("raw_status", rawStatus),
			zap.String("prop_types", propTypes(p)),
		)

		sid, err := statusID(rawStatus)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("page %s status: %v", p.ID, err))
			continue
		}

		desc, err := client.getPageText(p.ID)
		if err != nil {
			logger.Log.Warn("notion import: failed to fetch page content",
				zap.String("page_id", p.ID),
				zap.Error(err),
			)
		}

		t := &models.Task{
			BoardID:     boardID,
			Title:       title,
			Description: desc,
			StatusID:    sid,
			ReporterID:  uint(userID),
			DeveloperID: developerID(extractAssigneeEmail(p)),
		}

		if err := s.taskRepo.Create(t); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("page %s: %v", p.ID, err))
			continue
		}
		result.TasksCreated++

		for _, f := range extractFiles(p) {
			path, size, err := uploader.SaveFromURL(f.URL, f.Name)
			if err != nil {
				errMsg := fmt.Sprintf("page %s file %q: download failed: %v", p.ID, f.Name, err)
				logger.Log.Error("notion import: file download failed",
					zap.String("page_id", p.ID),
					zap.String("file_name", f.Name),
					zap.String("url", f.URL),
					zap.Error(err),
				)
				result.Errors = append(result.Errors, errMsg)
				continue
			}
			if err := s.attachmentRepo.Create(&models.Attachment{
				TaskID:   t.ID,
				FilePath: path,
				FileName: f.Name,
				FileSize: int(size),
			}); err != nil {
				errMsg := fmt.Sprintf("page %s file %q: save to db failed: %v", p.ID, f.Name, err)
				logger.Log.Error("notion import: attachment create failed",
					zap.String("page_id", p.ID),
					zap.String("file_name", f.Name),
					zap.String("path", path),
					zap.Error(err),
				)
				result.Errors = append(result.Errors, errMsg)
			}
		}
	}

	return result, nil
}
