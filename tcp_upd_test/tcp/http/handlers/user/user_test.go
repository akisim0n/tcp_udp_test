package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tcp_upd_test/tcp/http/models"
)

type handlerServiceStub struct {
	createFn func(context.Context, *models.User) (int64, error)
	updateFn func(context.Context, *models.User) error
	getFn    func(context.Context, int64) (*models.User, error)
	getAllFn func(context.Context) ([]*models.User, error)
	deleteFn func(context.Context, int64) error
}

func (s *handlerServiceStub) Create(ctx context.Context, data *models.User) (int64, error) {
	return s.createFn(ctx, data)
}

func (s *handlerServiceStub) Update(ctx context.Context, data *models.User) error {
	return s.updateFn(ctx, data)
}

func (s *handlerServiceStub) Delete(ctx context.Context, id int64) error {
	return s.deleteFn(ctx, id)
}

func (s *handlerServiceStub) Get(ctx context.Context, id int64) (*models.User, error) {
	return s.getFn(ctx, id)
}

func (s *handlerServiceStub) GetAll(ctx context.Context) ([]*models.User, error) {
	return s.getAllFn(ctx)
}

func TestUserCreateSuccess(t *testing.T) {
	service := &handlerServiceStub{
		createFn: func(_ context.Context, user *models.User) (int64, error) {
			if user.Name != "Alice" || user.Email != "alice@example.com" || user.Password != "secret" {
				t.Fatalf("unexpected user passed to service: %+v", user)
			}
			return 15, nil
		},
		updateFn: func(context.Context, *models.User) error { return nil },
		getFn:    func(context.Context, int64) (*models.User, error) { return nil, nil },
		getAllFn: func(context.Context) ([]*models.User, error) { return nil, nil },
		deleteFn: func(context.Context, int64) error { return nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"secret"}`))
	rec := httptest.NewRecorder()

	NewHandler(service).Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp map[string]int64
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response json error = %v", err)
	}

	if resp["id"] != 15 {
		t.Fatalf("response id = %d, want 15", resp["id"])
	}
}

func TestUserCreate(t *testing.T) {
	testCases := []struct {
		name       string
		body       string
		createFn   func(context.Context, *models.User) (int64, error)
		wantStatus int
		wantBody   []string
	}{
		{
			name: "validation error",
			body: `{"email":"alice@example.com"}`,
			createFn: func(context.Context, *models.User) (int64, error) {
				t.Fatal("service Create must not be called")
				return 0, nil
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   []string{`"VALIDATION_ERROR"`, `"name"`, `"password"`},
		},
		{
			name: "service error",
			body: `{"name":"Alice","email":"alice@example.com","password":"secret"}`,
			createFn: func(context.Context, *models.User) (int64, error) {
				return 0, errors.New("db down")
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   []string{"Create User Error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &handlerServiceStub{
				createFn: tc.createFn,
				updateFn: func(context.Context, *models.User) error { return nil },
				getFn:    func(context.Context, int64) (*models.User, error) { return nil, nil },
				getAllFn: func(context.Context) ([]*models.User, error) { return nil, nil },
				deleteFn: func(context.Context, int64) error { return nil },
			}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			NewHandler(service).Routes().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			body := rec.Body.String()
			for _, part := range tc.wantBody {
				if !strings.Contains(body, part) {
					t.Fatalf("response body %q does not contain %q", body, part)
				}
			}
		})
	}
}

func TestUserGetErrors(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		getFn      func(context.Context, int64) (*models.User, error)
		wantStatus int
		wantBody   string
	}{
		{
			name: "invalid id",
			path: "/id/not-a-number/",
			getFn: func(context.Context, int64) (*models.User, error) {
				t.Fatal("service Get must not be called for invalid id")
				return nil, nil
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "user get error",
		},
		{
			name: "service error",
			path: "/id/5/",
			getFn: func(context.Context, int64) (*models.User, error) {
				return nil, errors.New("not found")
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "user get error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &handlerServiceStub{
				createFn: func(context.Context, *models.User) (int64, error) { return 0, nil },
				updateFn: func(context.Context, *models.User) error { return nil },
				getFn:    tc.getFn,
				getAllFn: func(context.Context) ([]*models.User, error) { return nil, nil },
				deleteFn: func(context.Context, int64) error { return nil },
			}

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			NewHandler(service).Routes().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			if !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("unexpected body: %s", rec.Body.String())
			}
		})
	}
}

func TestUserListSuccess(t *testing.T) {
	surname := "Doe"
	service := &handlerServiceStub{
		createFn: func(context.Context, *models.User) (int64, error) { return 0, nil },
		updateFn: func(context.Context, *models.User) error { return nil },
		getFn:    func(context.Context, int64) (*models.User, error) { return nil, nil },
		getAllFn: func(context.Context) ([]*models.User, error) {
			return []*models.User{
				{ID: 1, Name: "Alice", Email: "alice@example.com", Surname: &surname},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			}, nil
		},
		deleteFn: func(context.Context, int64) error { return nil },
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	NewHandler(service).Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"total":2`) || !strings.Contains(body, `"alice@example.com"`) || !strings.Contains(body, `"bob@example.com"`) {
		t.Fatalf("unexpected list response: %s", body)
	}
}
