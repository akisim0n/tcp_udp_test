package user

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"tcp_upd_test/tcp/http/models"
)

type serviceRepoStub struct {
	createFn func(context.Context, *models.User) (int64, error)
	updateFn func(context.Context, *models.User) error
	getFn    func(context.Context, int64) (*models.User, error)
	getAllFn func(context.Context) ([]*models.User, error)
	deleteFn func(context.Context, int64) error
}

func (s *serviceRepoStub) Create(ctx context.Context, data *models.User) (int64, error) {
	return s.createFn(ctx, data)
}

func (s *serviceRepoStub) Update(ctx context.Context, data *models.User) error {
	return s.updateFn(ctx, data)
}

func (s *serviceRepoStub) Delete(ctx context.Context, id int64) error {
	return s.deleteFn(ctx, id)
}

func (s *serviceRepoStub) Get(ctx context.Context, id int64) (*models.User, error) {
	return s.getFn(ctx, id)
}

func (s *serviceRepoStub) GetAll(ctx context.Context) ([]*models.User, error) {
	return s.getAllFn(ctx)
}

func TestCreateHashesPasswordAndSetsTimestamps(t *testing.T) {
	t.Setenv("HEX_SALT", "4")

	var createdUser *models.User
	repo := &serviceRepoStub{
		createFn: func(_ context.Context, user *models.User) (int64, error) {
			copyUser := *user
			createdUser = &copyUser
			return 42, nil
		},
		updateFn: func(context.Context, *models.User) error { return nil },
		getFn:    func(context.Context, int64) (*models.User, error) { return nil, nil },
		getAllFn: func(context.Context) ([]*models.User, error) { return nil, nil },
		deleteFn: func(context.Context, int64) error { return nil },
	}

	svc := NewService(repo)
	user := &models.User{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret",
	}

	id, err := svc.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if id != 42 {
		t.Fatalf("Create() id = %d, want 42", id)
	}

	if createdUser == nil {
		t.Fatal("repository Create was not called")
	}

	if createdUser.Password == "secret" {
		t.Fatal("password was not hashed")
	}

	if createdUser.CreatedAt.IsZero() {
		t.Fatal("CreatedAt was not set")
	}

	if createdUser.UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt was not set")
	}

	if !createdUser.CreatedAt.Equal(createdUser.UpdatedAt) {
		t.Fatalf("timestamps differ: created=%v updated=%v", createdUser.CreatedAt, createdUser.UpdatedAt)
	}
}

func TestCreateValidatesRequiredFields(t *testing.T) {
	t.Setenv("HEX_SALT", "4")

	testCases := []struct {
		name    string
		user    *models.User
		wantErr string
	}{
		{
			name:    "missing name",
			user:    &models.User{Email: "alice@example.com", Password: "secret"},
			wantErr: "name is required",
		},
		{
			name:    "missing email",
			user:    &models.User{Name: "Alice", Password: "secret"},
			wantErr: "email is required",
		},
		{
			name:    "missing password",
			user:    &models.User{Name: "Alice", Email: "alice@example.com"},
			wantErr: "password is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &serviceRepoStub{
				createFn: func(context.Context, *models.User) (int64, error) {
					t.Fatal("repository Create must not be called")
					return 0, nil
				},
				updateFn: func(context.Context, *models.User) error { return nil },
				getFn:    func(context.Context, int64) (*models.User, error) { return nil, nil },
				getAllFn: func(context.Context) ([]*models.User, error) { return nil, nil },
				deleteFn: func(context.Context, int64) error { return nil },
			}

			svc := NewService(repo)
			_, err := svc.Create(context.Background(), tc.user)
			if err == nil {
				t.Fatal("Create() error = nil, want validation error")
			}

			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Create() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name        string
		user        *models.User
		wantErr     string
		checkUpdate bool
	}{
		{
			name:        "sets updated at for valid user",
			user:        &models.User{ID: 7, Name: "Bob"},
			checkUpdate: true,
		},
		{
			name:    "rejects nil user",
			user:    nil,
			wantErr: "user is required",
		},
		{
			name:    "rejects invalid id",
			user:    &models.User{ID: 0, Name: "Bob"},
			wantErr: "invalid user id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &serviceRepoStub{
				createFn: func(context.Context, *models.User) (int64, error) { return 0, nil },
				updateFn: func(_ context.Context, user *models.User) error {
					if tc.checkUpdate && user.UpdatedAt.IsZero() {
						t.Fatal("UpdatedAt was not set before repository call")
					}
					return nil
				},
				getFn:    func(context.Context, int64) (*models.User, error) { return nil, nil },
				getAllFn: func(context.Context) ([]*models.User, error) { return nil, nil },
				deleteFn: func(context.Context, int64) error { return nil },
			}

			svc := NewService(repo)
			before := time.Now()
			err := svc.Update(context.Background(), tc.user)

			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("Update() error = %v, want %q", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}

			if tc.user.UpdatedAt.Before(before) {
				t.Fatalf("Update() UpdatedAt = %v, want >= %v", tc.user.UpdatedAt, before)
			}
		})
	}
}

func TestDeleteRejectsInvalidID(t *testing.T) {
	testCases := []struct {
		name    string
		id      int64
		wantErr string
	}{
		{name: "zero id", id: 0, wantErr: "invalid user id"},
		{name: "negative id", id: -1, wantErr: "invalid user id"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &serviceRepoStub{
				createFn: func(context.Context, *models.User) (int64, error) { return 0, nil },
				updateFn: func(context.Context, *models.User) error { return nil },
				getFn:    func(context.Context, int64) (*models.User, error) { return nil, nil },
				getAllFn: func(context.Context) ([]*models.User, error) { return nil, nil },
				deleteFn: func(context.Context, int64) error {
					return errors.New("repository should not be called")
				},
			}

			svc := NewService(repo)
			err := svc.Delete(context.Background(), tc.id)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Delete() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
