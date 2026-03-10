package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(email, password string) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.CreateUser(email, string(hash))
}

func (s *Service) Login(email, password string) error {

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		return errors.New("invalid credentials")
	}

	return nil
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
