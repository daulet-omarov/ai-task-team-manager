package board

import (
	"errors"

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

	return &BoardResponse{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		MemberCount: 1,
		IsOwner:     true,
		OwnerID:     ownerID,
	}, nil
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
