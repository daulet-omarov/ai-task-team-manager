package employee

import (
	"errors"
	"time"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(userID uint, req CreateEmployeeRequest) error {
	birthday, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		return errors.New("invalid birthday format, use YYYY-MM-DD")
	}

	e := &models.Employee{
		ID:          userID, // employee.id == user.id always (1-to-1)
		UserID:      userID,
		FullName:    req.FullName,
		Photo:       req.Photo,
		Email:       req.Email,
		Birthday:    birthday,
		PhoneNumber: req.PhoneNumber,
		GenderID:    req.GenderID,
	}

	return s.repo.Create(e)
}

func (s *Service) GetByID(id uint) (*EmployeeResponse, error) {
	e, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("employee not found")
	}
	return toResponse(e), nil
}

func (s *Service) GetByUserID(userID uint) (*EmployeeResponse, error) {
	e, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("employee not found")
	}
	return toResponse(e), nil
}

func (s *Service) GetAll() ([]*EmployeeResponse, error) {
	employees, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var result []*EmployeeResponse
	for _, e := range employees {
		result = append(result, toResponse(e))
	}
	return result, nil
}

func (s *Service) Update(userID uint, req UpdateEmployeeRequest) error {
	e, err := s.repo.GetByUserID(userID)
	if err != nil {
		return errors.New("employee not found")
	}

	if req.FullName != "" {
		e.FullName = req.FullName
	}
	if req.Photo != "" {
		e.Photo = req.Photo
	}
	if req.Email != "" {
		e.Email = req.Email
	}
	if req.GenderID != 0 {
		e.GenderID = req.GenderID
	}
	if req.PhoneNumber != "" {
		e.PhoneNumber = req.PhoneNumber
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			return errors.New("invalid birthday format, use YYYY-MM-DD")
		}
		e.Birthday = birthday
	}

	return s.repo.Update(e)
}

func (s *Service) Delete(userID uint) error {
	return s.repo.Delete(userID)
}

func (s *Service) Exists(userID uint) (bool, error) {
	_, err := s.repo.GetByUserID(userID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// --- helper ---

func toResponse(e *models.Employee) *EmployeeResponse {
	return &EmployeeResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		FullName:    e.FullName,
		Photo:       e.Photo,
		Email:       e.Email,
		PhoneNumber: e.PhoneNumber,
		Birthday:    e.Birthday.Format("2006-01-02"),
		Gender:      e.Gender,
	}
}
