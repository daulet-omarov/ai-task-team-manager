package auth

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(email, password string) (int64, error) {
	query := `
        INSERT INTO users (email, password)
        VALUES ($1, $2)
        RETURNING id
    `
	var id int64
	err := r.db.QueryRow(query, email, password).Scan(&id)
	return id, err
}

func (r *Repository) GetByEmail(email string) (*User, error) {
	query := `SELECT id, email, password, is_verified FROM users WHERE email=$1`
	user := &User{}
	err := r.db.QueryRow(query, email).
		Scan(&user.ID, &user.Email, &user.Password, &user.IsVerified)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) MarkEmailVerified(userID int64) error {
	_, err := r.db.Exec(`UPDATE users SET is_verified=TRUE WHERE id=$1`, userID)
	return err
}

func (r *Repository) CreateVerificationToken(userID int64, token string, expiresAt time.Time) error {
	query := `
        INSERT INTO email_verification_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.Exec(query, userID, token, expiresAt)
	return err
}

func (r *Repository) GetVerificationToken(token string) (*EmailVerificationToken, error) {
	query := `
        SELECT id, user_id, token, expires_at
        FROM email_verification_tokens
        WHERE token=$1
    `
	t := &EmailVerificationToken{}
	err := r.db.QueryRow(query, token).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) DeleteVerificationToken(token string) error {
	_, err := r.db.Exec(`DELETE FROM email_verification_tokens WHERE token=$1`, token)
	return err
}

func (r *Repository) DeleteUser(id int64) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *Repository) CreatePasswordResetToken(userID int64, token string, expiresAt time.Time) error {
	// Сначала удаляем старый токен если был — чтобы не накапливались
	_, _ = r.db.Exec(`DELETE FROM password_reset_tokens WHERE user_id=$1`, userID)

	query := `
        INSERT INTO password_reset_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.Exec(query, userID, token, expiresAt)
	return err
}

func (r *Repository) GetPasswordResetToken(token string) (*PasswordResetToken, error) {
	query := `
        SELECT id, user_id, token, expires_at
        FROM password_reset_tokens
        WHERE token=$1
    `
	t := &PasswordResetToken{}
	err := r.db.QueryRow(query, token).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) UpdatePassword(userID int64, hashedPassword string) error {
	_, err := r.db.Exec(`UPDATE users SET password=$1 WHERE id=$2`, hashedPassword, userID)
	return err
}

func (r *Repository) DeletePasswordResetToken(token string) error {
	_, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE token=$1`, token)
	return err
}
