package auth

import (
	"errors"

	"github.com/daulet-omarov/ai-task-team-manager/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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

	return s.repo.CreateUser(email, string(hash))
}

func (s *Service) Login(email, password string) (string, error) {

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		return "", errors.New("invalid credentials")
	}

	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) ForgotPassword(email string) error {

	_, err := s.repo.GetByEmail(email)
	if err != nil {
		return err
	}

	// здесь позже можно отправить email
	return nil
}

func (s *Service) DeleteAccount(userID int64) error {
	return s.repo.DeleteUser(userID)
}
