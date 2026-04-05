package employee

type CreateEmployeeRequest struct {
	UserID      uint   `json:"user_id" validate:"required"`
	FullName    string `json:"full_name" validate:"required"`
	Photo       string `json:"photo"`
	Email       string `json:"email" validate:"required,email"`
	RoleID      uint   `json:"role_id" validate:"required"`
	Birthday    string `json:"birthday" validate:"required"` // format: 2006-01-02
	PhoneNumber string `json:"phone_number"`
	GenderID    uint   `json:"gender_id" validate:"required"`
}

type UpdateEmployeeRequest struct {
	FullName    string `json:"full_name"`
	Photo       string `json:"photo"`
	Email       string `json:"email" validate:"omitempty,email"`
	RoleID      uint   `json:"role_id"`
	Birthday    string `json:"birthday"` // format: 2006-01-02
	PhoneNumber string `json:"phone_number"`
	GenderID    uint   `json:"gender_id"`
}

type EmployeeResponse struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	FullName    string `json:"full_name"`
	Photo       string `json:"photo"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Birthday    string `json:"birthday"`
	Role        Role   `json:"role"`
	Gender      Gender `json:"gender"`
}
