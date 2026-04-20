package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"tcp_upd_test/tcp/http/models"
	"tcp_upd_test/tcp/http/repository"
	"tcp_upd_test/tcp/http/service"
	"tcp_upd_test/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type serv struct {
	repo repository.UserRepository
	salt int
}

func NewService(userRepo repository.UserRepository) service.UserService {
	salt, convErr := strconv.Atoi(utils.GetEnvParam("HEX_SALT"))
	if convErr != nil {
		salt = bcrypt.DefaultCost
		log.Print("Salt not set correctly")
	}

	return &serv{
		repo: userRepo,
		salt: salt,
	}
}

func (s *serv) Get(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user id")
	}

	return s.repo.Get(ctx, id)
}

func (s *serv) GetAll(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *serv) Update(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user is required")
	}

	if user.ID <= 0 {
		return errors.New("invalid user id")
	}

	user.UpdatedAt = time.Now()
	return s.repo.Update(ctx, user)
}

func (s *serv) Create(ctx context.Context, user *models.User) (int64, error) {
	if user == nil {
		return 0, errors.New("user is required")
	}

	if user.Name == "" {
		return 0, errors.New("name is required")
	}

	if user.Email == "" {
		return 0, errors.New("email is required")
	}

	if user.Password == "" {
		return 0, errors.New("password is required")
	}

	hashedPassword, err := s.hashPassword(user.Password)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	user.Password = hashedPassword
	user.CreatedAt = now
	user.UpdatedAt = now

	return s.repo.Create(ctx, user)
}

func (s *serv) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid user id")
	}

	return s.repo.Delete(ctx, id)
}

func (s *serv) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.salt)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hashedPassword), nil
}
