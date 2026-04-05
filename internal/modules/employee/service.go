package employee

import (
	"errors"
	"time"
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

	e := &Employee{
		UserID:      userID,
		FullName:    req.FullName,
		Photo:       req.Photo,
		Email:       req.Email,
		RoleID:      req.RoleID,
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

func (s *Service) Update(id uint, req UpdateEmployeeRequest) error {
	e, err := s.repo.GetByID(id)
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
	if req.RoleID != 0 {
		e.RoleID = req.RoleID
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

func (s *Service) Delete(id uint) error {
	return s.repo.Delete(id)
}

// --- helper ---

func toResponse(e *Employee) *EmployeeResponse {
	return &EmployeeResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		FullName:    e.FullName,
		Photo:       e.Photo,
		Email:       e.Email,
		PhoneNumber: e.PhoneNumber,
		Birthday:    e.Birthday.Format("2006-01-02"),
		Role:        e.Role,
		Gender:      e.Gender,
	}
}
