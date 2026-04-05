package auth

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/modules/employee"
	"time"
)

type User struct {
	ID         int64
	Email      string
	Password   string
	IsVerified bool
	CreatedAt  time.Time
	Employee   *employee.Employee
}

type EmailVerificationToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type PasswordResetToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
