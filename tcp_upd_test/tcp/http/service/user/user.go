package user

import (
	"context"
	"tcp_upd_test/tcp/http/models"
	"tcp_upd_test/tcp/http/repository"
	"tcp_upd_test/tcp/http/service"
)

type serv struct {
	repo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) service.UserService {
	return &serv{repo: userRepo}
}

func (s *serv) Get(ctx context.Context, id int64) (*models.User, error) {
	return s.repo.Get(ctx, id)
}
func (s *serv) GetAll(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAll(ctx)
}
func (s *serv) Update(ctx context.Context, user *models.User) error {
	return s.repo.Update(ctx, user)
}
func (s *serv) Create(ctx context.Context, user *models.User) (int64, error) {
	return s.repo.Create(ctx, user)
}
func (s *serv) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
