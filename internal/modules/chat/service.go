package chat

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/board"
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/uploader"
)

type Service struct {
	repo         *Repository
	boardRepo    *board.Repository
	employeeRepo *employee.Repository
}

func NewService(repo *Repository, boardRepo *board.Repository, employeeRepo *employee.Repository) *Service {
	return &Service{repo: repo, boardRepo: boardRepo, employeeRepo: employeeRepo}
}

// employeeIDFromUser resolves an auth userID to an employee ID.
func (s *Service) employeeIDFromUser(userID int64) (uint, error) {
	emp, err := s.employeeRepo.GetByUserID(uint(userID))
	if err != nil {
		return 0, errors.New("employee profile not found")
	}
	return emp.ID, nil
}

// checkMember verifies the user is a board member.
func (s *Service) checkMember(boardID uint, userID int64) error {
	ok, err := s.boardRepo.IsMember(boardID, userID)
	if err != nil || !ok {
		return errors.New("access denied: not a board member")
	}
	return nil
}

// GetMessages returns paginated messages for a board (newest first).
func (s *Service) GetMessages(boardID uint, userID int64, limit, offset int) ([]models.BoardChatMessage, error) {
	if err := s.checkMember(boardID, userID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.GetMessages(boardID, limit, offset)
}

// SendMessage creates a text / file / voice message.
// files is the parsed multipart form files field (may be nil).
func (s *Service) SendMessage(boardID uint, userID int64, r *http.Request, req SendMessageRequest) (*models.BoardChatMessage, error) {
	if err := s.checkMember(boardID, userID); err != nil {
		return nil, err
	}
	empID, err := s.employeeIDFromUser(userID)
	if err != nil {
		return nil, err
	}

	if req.Text == "" && len(r.MultipartForm.File["files"]) == 0 {
		return nil, errors.New("message must have text or at least one file")
	}

	msg := &models.BoardChatMessage{
		BoardID:   boardID,
		AuthorID:  empID,
		ReplyToID: req.ReplyToID,
		Text:      req.Text,
	}
	if err := s.repo.CreateMessage(msg); err != nil {
		return nil, err
	}

	// Save attached files
	for _, fh := range r.MultipartForm.File["files"] {
		path, mime, err := uploader.SaveAny(fh)
		if err != nil {
			continue // skip bad files, don't fail the whole message
		}
		_ = s.repo.CreateAttachment(&models.BoardChatAttachment{
			MessageID: msg.ID,
			FilePath:  path,
			FileName:  fh.Filename,
			FileSize:  int(fh.Size),
			MimeType:  mime,
		})
	}

	return s.repo.GetMessageByID(msg.ID)
}

// CreatePoll creates a poll message.
func (s *Service) CreatePoll(boardID uint, userID int64, req CreatePollRequest) (*models.BoardChatMessage, error) {
	if err := s.checkMember(boardID, userID); err != nil {
		return nil, err
	}
	empID, err := s.employeeIDFromUser(userID)
	if err != nil {
		return nil, err
	}

	msg := &models.BoardChatMessage{
		BoardID:  boardID,
		AuthorID: empID,
	}
	if err := s.repo.CreateMessage(msg); err != nil {
		return nil, err
	}

	options := make([]models.BoardPollOption, len(req.Options))
	for i, o := range req.Options {
		options[i] = models.BoardPollOption{Text: o}
	}
	poll := &models.BoardPoll{
		MessageID: msg.ID,
		Question:  req.Question,
		Options:   options,
	}
	if err := s.repo.CreatePoll(poll); err != nil {
		return nil, err
	}

	return s.repo.GetMessageByID(msg.ID)
}

// Vote adds or removes a vote for a poll option.
func (s *Service) Vote(boardID uint, userID int64, req VoteRequest) (*models.BoardPoll, error) {
	if err := s.checkMember(boardID, userID); err != nil {
		return nil, err
	}
	empID, err := s.employeeIDFromUser(userID)
	if err != nil {
		return nil, err
	}

	opt, err := s.repo.GetOptionByID(req.OptionID)
	if err != nil {
		return nil, errors.New("poll option not found")
	}

	// Verify the poll belongs to this board
	poll, err := s.repo.GetPollByID(opt.PollID)
	if err != nil {
		return nil, errors.New("poll not found")
	}
	msg, err := s.repo.GetMessageByID(poll.MessageID)
	if err != nil || msg.BoardID != boardID {
		return nil, errors.New("poll does not belong to this board")
	}

	// Single-choice: if the user already voted in this poll, remove the old vote first.
	existing, err := s.repo.GetVoteInPoll(poll.ID, empID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.OptionID == req.OptionID {
			return nil, errors.New("already voted for this option")
		}
		if err := s.repo.RemoveVote(existing.OptionID, empID); err != nil {
			return nil, err
		}
	}
	if err := s.repo.AddVote(&models.BoardPollVote{
		OptionID:   req.OptionID,
		EmployeeID: empID,
	}); err != nil {
		return nil, err
	}

	return s.repo.GetPollByID(poll.ID)
}

// Unvote removes the user's vote from a poll option.
func (s *Service) Unvote(boardID uint, userID int64, req VoteRequest) (*models.BoardPoll, error) {
	if err := s.checkMember(boardID, userID); err != nil {
		return nil, err
	}
	empID, err := s.employeeIDFromUser(userID)
	if err != nil {
		return nil, err
	}

	opt, err := s.repo.GetOptionByID(req.OptionID)
	if err != nil {
		return nil, errors.New("poll option not found")
	}

	poll, err := s.repo.GetPollByID(opt.PollID)
	if err != nil {
		return nil, errors.New("poll not found")
	}

	msg, err := s.repo.GetMessageByID(poll.MessageID)
	if err != nil || msg.BoardID != boardID {
		return nil, errors.New("poll does not belong to this board")
	}

	voted, err := s.repo.HasVoted(req.OptionID, empID)
	if err != nil {
		return nil, err
	}
	if !voted {
		return nil, errors.New("you have not voted for this option")
	}

	if err := s.repo.RemoveVote(req.OptionID, empID); err != nil {
		return nil, err
	}

	return s.repo.GetPollByID(poll.ID)
}

// DeleteMessage deletes a message owned by the user.
func (s *Service) DeleteMessage(boardID uint, userID int64, msgID uint) error {
	if err := s.checkMember(boardID, userID); err != nil {
		return err
	}
	empID, err := s.employeeIDFromUser(userID)
	if err != nil {
		return err
	}
	msg, err := s.repo.GetMessageByID(msgID)
	if err != nil {
		return errors.New("message not found")
	}
	if msg.BoardID != boardID {
		return errors.New("message does not belong to this board")
	}
	if msg.AuthorID != empID {
		return errors.New("access denied: not the message author")
	}
	return s.repo.DeleteMessage(msgID)
}

// toAuthorInfo converts an Employee to AuthorInfo.
func toAuthorInfo(e models.Employee, r *http.Request) AuthorInfo {
	return AuthorInfo{
		ID:       e.ID,
		FullName: e.FullName,
		Photo:    uploader.FullURL(r, e.Photo),
	}
}

// BuildMessageResponse converts a model to the API response shape.
func BuildMessageResponse(m models.BoardChatMessage, r *http.Request) MessageResponse {
	attachments := make([]AttachmentResponse, len(m.Attachments))
	for i, a := range m.Attachments {
		attachments[i] = AttachmentResponse{
			ID:       a.ID,
			FileName: a.FileName,
			FileSize: a.FileSize,
			MimeType: a.MimeType,
			URL:      uploader.FullURL(r, a.FilePath),
		}
	}

	var replyTo *ReplyToResponse
	if m.ReplyTo != nil {
		author := toAuthorInfo(m.ReplyTo.Author, r)
		replyTo = &ReplyToResponse{
			ID:     m.ReplyTo.ID,
			Author: &author,
			Text:   m.ReplyTo.Text,
		}
	}

	var poll *PollResponse
	if m.Poll != nil {
		opts := make([]PollOptionResponse, len(m.Poll.Options))
		for i, o := range m.Poll.Options {
			voters := make([]AuthorInfo, len(o.Votes))
			for j, v := range o.Votes {
				voters[j] = toAuthorInfo(v.Employee, r)
			}
			opts[i] = PollOptionResponse{
				ID:        o.ID,
				Text:      o.Text,
				VoteCount: len(o.Votes),
				Voters:    voters,
			}
		}
		poll = &PollResponse{
			ID:       m.Poll.ID,
			Question: m.Poll.Question,
			Options:  opts,
		}
	}

	return MessageResponse{
		ID:          m.ID,
		BoardID:     m.BoardID,
		Author:      toAuthorInfo(m.Author, r),
		Text:        m.Text,
		ReplyTo:     replyTo,
		Attachments: attachments,
		Poll:        poll,
		CreatedAt:   m.CreatedAt,
	}
}

// BuildPollResponse converts a poll model to the API response shape.
func BuildPollResponse(p models.BoardPoll) PollResponse {
	opts := make([]PollOptionResponse, len(p.Options))
	for i, o := range p.Options {
		voters := make([]AuthorInfo, len(o.Votes))
		for j, v := range o.Votes {
			voters[j] = AuthorInfo{
				ID:       v.Employee.ID,
				FullName: v.Employee.FullName,
				Photo:    v.Employee.Photo,
			}
		}
		opts[i] = PollOptionResponse{
			ID:        o.ID,
			Text:      o.Text,
			VoteCount: len(o.Votes),
			Voters:    voters,
		}
	}
	return PollResponse{
		ID:       p.ID,
		Question: p.Question,
		Options:  opts,
	}
}

// ParseMultipartOrEmpty ensures r.MultipartForm is never nil.
func ParseMultipartOrEmpty(r *http.Request) {
	if r.MultipartForm == nil {
		_ = r.ParseMultipartForm(200 << 20)
		if r.MultipartForm == nil {
			r.MultipartForm = &multipart.Form{
				Value: map[string][]string{},
				File:  map[string][]*multipart.FileHeader{},
			}
		}
	}
}
