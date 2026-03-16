package auth

import "time"

type User struct {
	ID         int64
	Email      string
	Password   string
	IsVerified bool
	CreatedAt  time.Time
}

type EmailVerificationToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
