package employee

import "github.com/daulet-omarov/ai-task-team-manager/internal/models"

type CreateEmployeeRequest struct {
	FullName    string `json:"full_name" validate:"required"`
	Photo       string `json:"photo"`
	Email       string `json:"email" validate:"required,email"`
	Birthday    string `json:"birthday" validate:"required"` // format: 2006-01-02
	PhoneNumber string `json:"phone_number"`
	GenderID    uint   `json:"gender_id" validate:"required"`
}

type UpdateEmployeeRequest struct {
	FullName    string `json:"full_name"`
	Photo       string `json:"photo"`
	Email       string `json:"email" validate:"omitempty,email"`
	Birthday    string `json:"birthday"` // format: 2006-01-02
	PhoneNumber string `json:"phone_number"`
	GenderID    uint   `json:"gender_id"`
}

type EmployeeResponse struct {
	ID          uint          `json:"id"`
	UserID      uint          `json:"user_id"`
	FullName    string        `json:"full_name"`
	Photo       string        `json:"photo"`
	Email       string        `json:"email"`
	PhoneNumber string        `json:"phone_number"`
	Birthday    string        `json:"birthday"`
	Gender      models.Gender `json:"gender"`
}
