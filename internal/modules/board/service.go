package board

import (
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"gorm.io/gorm"
)

// employeeChecker abstracts the employee repository to avoid a direct package dependency.
type employeeChecker interface {
	GetByUserID(userID uint) (*models.Employee, error)
}

type Service struct {
	repo    *Repository
	empRepo employeeChecker
}

func NewService(repo *Repository, empRepo employeeChecker) *Service {
	return &Service{repo: repo, empRepo: empRepo}
}

// GetDashboard returns the user's boards and whether this is their first login
// (first login = no employee profile yet).
func (s *Service) GetDashboard(userID int64) (*DashboardResponse, error) {
	boards, err := s.repo.GetBoardsByUserID(userID)
	if err != nil {
		return nil, err
	}

	_, empErr := s.empRepo.GetByUserID(uint(userID))
	isFirstLogin := errors.Is(empErr, gorm.ErrRecordNotFound)

	boardResponses := make([]BoardResponse, 0, len(boards))
	for _, b := range boards {
		count, _ := s.repo.CountMembers(b.ID)
		boardResponses = append(boardResponses, BoardResponse{
			ID:          b.ID,
			Name:        b.Name,
			Description: b.Description,
			MemberCount: int(count),
			IsOwner:     b.OwnerID == userID,
			OwnerID:     b.OwnerID,
		})
	}

	return &DashboardResponse{
		Boards:       boardResponses,
		IsFirstLogin: isFirstLogin,
	}, nil
}

func (s *Service) Create(ownerID int64, req CreateBoardRequest) (*BoardResponse, error) {
	b := &models.Board{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	}

	if err := s.repo.Create(b); err != nil {
		return nil, err
	}

	// Creator is automatically added as owner member.
	if err := s.repo.AddMember(b.ID, ownerID, "owner"); err != nil {
		return nil, err
	}

	// Default statuses: TO DO, IN PROGRESS, DONE.
	if err := s.repo.AddDefaultStatuses(b.ID); err != nil {
		return nil, err
	}

	return &BoardResponse{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		MemberCount: 1,
		IsOwner:     true,
		OwnerID:     ownerID,
	}, nil
}

func (s *Service) GetMembers(boardID uint, userID int64) ([]*MemberResponse, error) {
	isMember, err := s.repo.IsMember(boardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}
	return s.repo.GetMembersWithDetails(boardID)
}

func (s *Service) Delete(boardID uint, requesterID int64) error {
	isOwner, err := s.repo.IsOwner(boardID, requesterID)
	if err != nil {
		return errors.New("board not found")
	}
	if !isOwner {
		return errors.New("access denied")
	}
	return s.repo.Delete(boardID)
}

func (s *Service) DeleteMember(boardMemberID uint, requesterID int64) (uint, error) {
	log.Printf("board_member_id = %d, requester_id = %d", boardMemberID, requesterID)
	member, err := s.repo.GetMemberByID(boardMemberID)
	if err != nil {
		return 0, errors.New("member not found")
	}

	isOwner, err := s.repo.IsOwner(member.BoardID, requesterID)
	if err != nil || !isOwner {
		return 0, errors.New("access denied")
	}

	if member.Role == "owner" {
		return 0, errors.New("cannot remove the board owner")
	}

	return member.BoardID, s.repo.DeleteMember(boardMemberID)
}

func (s *Service) GetStatuses(boardID uint, userID int64) ([]*StatusResponse, error) {
	isMember, err := s.repo.IsMember(boardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}
	return s.repo.GetBoardStatuses(boardID)
}

func (s *Service) CreateStatus(userID int64, req CreateStatusRequest) (*StatusResponse, error) {
	isMember, err := s.repo.IsMember(req.BoardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}
	code := titleToCode(req.Title)
	return s.repo.UpsertStatus(req.BoardID, req.Title, code, req.Colour)
}

func (s *Service) UpdateStatus(boardStatusID uint, userID int64, req UpdateStatusRequest) (uint, *StatusResponse, error) {
	boardID, err := s.repo.GetBoardIDByBoardStatusID(boardStatusID)
	if err != nil || boardID == 0 {
		return 0, nil, errors.New("status not found")
	}
	isMember, err := s.repo.IsMember(boardID, userID)
	if err != nil || !isMember {
		return 0, nil, errors.New("access denied")
	}
	status, err := s.repo.UpdateBoardStatus(boardStatusID, req.Title, req.Colour)
	return boardID, status, err
}

func (s *Service) ReorderStatuses(userID int64, req ReorderStatusesRequest) error {
	return s.repo.ReorderStatuses(req.Statuses)
}

func (s *Service) SetDefaultStatus(boardStatusID uint, userID int64) (uint, error) {
	boardID, err := s.repo.GetBoardIDByBoardStatusID(boardStatusID)
	if err != nil || boardID == 0 {
		return 0, errors.New("status not found")
	}
	isMember, err := s.repo.IsMember(boardID, userID)
	if err != nil || !isMember {
		return 0, errors.New("access denied")
	}
	return boardID, s.repo.SetDefaultBoardStatus(boardStatusID)
}

func (s *Service) DeleteStatus(boardStatusID uint, userID int64) (uint, error) {
	boardID, err := s.repo.GetBoardIDByBoardStatusID(boardStatusID)
	if err != nil || boardID == 0 {
		return 0, errors.New("status not found")
	}
	isMember, err := s.repo.IsMember(boardID, userID)
	if err != nil || !isMember {
		return 0, errors.New("access denied")
	}
	return boardID, s.repo.DeleteBoardStatus(boardStatusID)
}

func (s *Service) IsMember(boardID uint, userID int64) (bool, error) {
	return s.repo.IsMember(boardID, userID)
}

func (s *Service) BoardIDByBoardStatusID(boardStatusID uint) (uint, error) {
	return s.repo.GetBoardIDByBoardStatusID(boardStatusID)
}

// titleToCode converts "Code Review" → "code_review".
var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

func titleToCode(title string) string {
	lower := strings.ToLower(strings.TrimSpace(title))
	return nonAlphanumRe.ReplaceAllString(lower, "_")
}

func (s *Service) GetByID(boardID uint, userID int64) (*BoardResponse, error) {
	b, err := s.repo.GetByID(boardID)
	if err != nil {
		return nil, errors.New("board not found")
	}

	isMember, err := s.repo.IsMember(boardID, userID)
	if err != nil || !isMember {
		return nil, errors.New("access denied")
	}

	count, _ := s.repo.CountMembers(boardID)

	return &BoardResponse{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		MemberCount: int(count),
		IsOwner:     b.OwnerID == userID,
		OwnerID:     b.OwnerID,
	}, nil
}
