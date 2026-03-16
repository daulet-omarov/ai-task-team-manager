package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/daulet-omarov/ai-task-team-manager/pkg/jwt"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/mailer"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo       *Repository
	mailer     *mailer.Mailer
	appBaseURL string
}

func NewService(repo *Repository, mailer *mailer.Mailer, appBaseURL string) *Service {
	return &Service{repo: repo, mailer: mailer, appBaseURL: appBaseURL}
}

func (s *Service) Register(email, password string) error {
	existing, _ := s.repo.GetByEmail(email)
	if existing != nil {
		return errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userID, err := s.repo.CreateUser(email, string(hash))
	if err != nil {
		return err
	}

	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(24 * time.Hour)

	if err := s.repo.CreateVerificationToken(userID, token, expiresAt); err != nil {
		return err
	}

	// Send verification email (non-blocking — don't fail registration if mail fails)
	verificationURL := s.appBaseURL + "/auth/verify-email?token=" + token
	go s.mailer.SendVerificationEmail(email, verificationURL)

	return nil
}

func (s *Service) VerifyEmail(token string) error {
	vt, err := s.repo.GetVerificationToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if time.Now().After(vt.ExpiresAt) {
		_ = s.repo.DeleteVerificationToken(token)
		return errors.New("token has expired")
	}

	if err := s.repo.MarkEmailVerified(vt.UserID); err != nil {
		return err
	}

	return s.repo.DeleteVerificationToken(token)
}

func (s *Service) Login(email, password string) (string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !user.IsVerified {
		return "", errors.New("email not verified, please check your inbox")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return jwt.GenerateToken(user.ID)
}

func (s *Service) ForgotPassword(email string) error {
	_, err := s.repo.GetByEmail(email)
	return err
}

func (s *Service) DeleteAccount(userID int64) error {
	return s.repo.DeleteUser(userID)
}
