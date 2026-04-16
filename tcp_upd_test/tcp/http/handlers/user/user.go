package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	c "tcp_upd_test/tcp/http/converters"
	"tcp_upd_test/tcp/http/handlers"
	"tcp_upd_test/tcp/http/models"
	"tcp_upd_test/tcp/http/models/dto"
	"tcp_upd_test/tcp/http/service"
	"tcp_upd_test/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type handler struct {
	serv service.UserService
}

const (
	validationErrCode = "VALIDATION_ERROR"
)

func NewHandler(serv service.UserService) handlers.UserHandler {
	return &handler{serv: serv}
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", h.UserList)
	r.Post("/", h.UserCreate)

	r.Route("/id/{id}", func(r chi.Router) {
		r.Get("/", h.User)
		r.Put("/", h.UserUpdate)
		r.Delete("/", h.UserDelete)
	})

	return r
}

func (h *handler) User(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	var err error
	var idInt64 int64
	defer func() {
		if err != nil {
			createServerErr(w, fmt.Sprintf("user get error: %v", err))
		}
	}()

	idInt64, err = strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	var user *models.User

	user, err = h.serv.Get(ctx, idInt64)
	if err != nil {
		return
	}

	userResp := c.FromUserToCreateUserResponse(*user)

	utils.WriteJSON(w, http.StatusOK, userResp)
}

func (h *handler) UserUpdate(w http.ResponseWriter, r *http.Request) {
	var user models.User
	ctx := r.Context()
	var err error
	defer func() {
		if err != nil {
			createServerErr(w, fmt.Sprintf("user update error: %v", err))
		}
	}()
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return
	}

	err = h.serv.Update(ctx, &user)
	if err != nil {
		return
	}

	userResp := c.FromUserToUserResponse(user)

	utils.WriteJSON(w, http.StatusOK, userResp)
}

func (h *handler) UserDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	var err error
	var idInt64 int64
	defer func() {
		if err != nil {
			createServerErr(w, fmt.Sprintf("user delete error: %v", err))
		}
	}()

	idInt64, err = strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	err = h.serv.Delete(ctx, idInt64)
	if err != nil {
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *handler) UserList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	defer func() {
		if err != nil {
			createServerErr(w, fmt.Sprintf("getting users error: %v", err))
		}
	}()
	var users []*models.User

	users, err = h.serv.GetAll(ctx)
	if err != nil {
		return
	}

	var usersListResponse dto.UserListResponse
	for _, user := range users {
		userResp := c.FromUserToUserResponse(*user)
		usersListResponse.Users = append(usersListResponse.Users, userResp)
	}
	usersListResponse.Total = len(usersListResponse.Users)

	utils.WriteJSON(w, http.StatusOK, usersListResponse)
}

func (h *handler) UserCreate(w http.ResponseWriter, r *http.Request) {
	var reqUser dto.CreateUserRequest
	var err error
	validateErr := dto.ValidationErrorResponse{}
	ctx := r.Context()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	defer r.Body.Close()
	defer func() {
		if err != nil {
			createServerErr(w, fmt.Sprintf("Create User Error: %v", err))
		}
	}()

	if err = decoder.Decode(&reqUser); err != nil {
		return
	}

	if reqUser.Name == nil {
		fieldErr := dto.FieldError{
			Field:   "name",
			Message: fmt.Sprintf("Field 'name' is required"),
		}
		validateErr.Error.Fields = append(validateErr.Error.Fields, fieldErr)
	}

	if reqUser.Email == nil {
		fieldErr := dto.FieldError{
			Field:   "email",
			Message: fmt.Sprintf("Field 'email' is required"),
		}
		validateErr.Error.Fields = append(validateErr.Error.Fields, fieldErr)
	}

	if reqUser.Password == nil {
		fieldErr := dto.FieldError{
			Field:   "password",
			Message: fmt.Sprintf("Field 'password' is required"),
		}
		validateErr.Error.Fields = append(validateErr.Error.Fields, fieldErr)
	}

	if validateErr.Error.Fields != nil && len(validateErr.Error.Fields) > 0 {
		validateErr.Error.Code = validationErrCode
		validateErr.Error.Message = fmt.Sprint("invalid request data")
		utils.WriteJSON(w, http.StatusBadRequest, validateErr)
		return
	}

	user := c.FromCreateUserRequestToUser(reqUser)

	user.ID, err = h.serv.Create(ctx, &user)
	if err != nil {
		return
	}

	respUser := dto.CreateUserResponse{Id: user.ID}

	utils.WriteJSON(w, http.StatusCreated, respUser)
}

func createServerErr(w http.ResponseWriter, errorMsg string) {
	serverErr := dto.ServerError{
		Code:    http.StatusBadRequest,
		Message: errorMsg,
	}
	utils.WriteJSON(w, http.StatusBadRequest, serverErr)
}
