package service

import (
	"context"
	"tcp_upd_test/tcp/http/models"
)

type UserService interface {
	Create(ctx context.Context, data *models.User) (int64, error)
	Update(ctx context.Context, data *models.User) error
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*models.User, error)
	GetAll(ctx context.Context) ([]*models.User, error)
}
