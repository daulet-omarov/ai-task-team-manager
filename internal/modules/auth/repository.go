package auth

import (
	"gorm.io/gorm"
	"time"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(email, password string) (int64, error) {
	user := &User{Email: email, Password: password}
	result := r.db.Create(user)
	return int64(user.ID), result.Error
}

func (r *Repository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) MarkEmailVerified(userID int64) error {
	return r.db.Model(&User{}).
		Where("id = ?", userID).
		Update("is_verified", true).Error
}

func (r *Repository) CreateVerificationToken(userID int64, token string, expiresAt time.Time) error {
	t := &EmailVerificationToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	return r.db.Create(t).Error
}

func (r *Repository) GetVerificationToken(token string) (*EmailVerificationToken, error) {
	var t EmailVerificationToken
	err := r.db.Where("token = ?", token).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) DeleteVerificationToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&EmailVerificationToken{}).Error
}

func (r *Repository) DeleteUser(id int64) error {
	return r.db.Delete(&User{}, id).Error
}

func (r *Repository) CreatePasswordResetToken(userID int64, token string, expiresAt time.Time) error {
	// Delete old token first
	r.db.Where("user_id = ?", userID).Delete(&PasswordResetToken{})

	t := &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	return r.db.Create(t).Error
}

func (r *Repository) GetPasswordResetToken(token string) (*PasswordResetToken, error) {
	var t PasswordResetToken
	err := r.db.Where("token = ?", token).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) UpdatePassword(userID int64, hashedPassword string) error {
	return r.db.Model(&User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

func (r *Repository) DeletePasswordResetToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&PasswordResetToken{}).Error
}
