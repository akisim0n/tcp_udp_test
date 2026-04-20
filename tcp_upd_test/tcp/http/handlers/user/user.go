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
	idInt64, err := parseIDParam(r)
	if err != nil {
		createServerErr(w, fmt.Sprintf("user get error: %v", err))
		return
	}

	user, err := h.serv.Get(ctx, idInt64)
	if err != nil {
		createServerErr(w, fmt.Sprintf("user get error: %v", err))
		return
	}

	userResp := c.FromUserToCreateUserResponse(*user)
	utils.WriteJSON(w, http.StatusOK, userResp)
}

func (h *handler) UserUpdate(w http.ResponseWriter, r *http.Request) {
	var user models.User
	ctx := r.Context()
	if err := decodeJSON(r, &user); err != nil {
		createServerErr(w, fmt.Sprintf("user update error: %v", err))
		return
	}

	if err := h.serv.Update(ctx, &user); err != nil {
		createServerErr(w, fmt.Sprintf("user update error: %v", err))
		return
	}

	userResp := c.FromUserToUserResponse(user)
	utils.WriteJSON(w, http.StatusOK, userResp)
}

func (h *handler) UserDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idInt64, err := parseIDParam(r)
	if err != nil {
		createServerErr(w, fmt.Sprintf("user delete error: %v", err))
		return
	}

	if err := h.serv.Delete(ctx, idInt64); err != nil {
		createServerErr(w, fmt.Sprintf("user delete error: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *handler) UserList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := h.serv.GetAll(ctx)
	if err != nil {
		createServerErr(w, fmt.Sprintf("getting users error: %v", err))
		return
	}

	usersListResponse := dto.UserListResponse{
		Users: make([]dto.UserResponse, 0, len(users)),
	}

	for _, user := range users {
		userResp := c.FromUserToUserResponse(*user)
		usersListResponse.Users = append(usersListResponse.Users, userResp)
	}
	usersListResponse.Total = len(usersListResponse.Users)

	utils.WriteJSON(w, http.StatusOK, usersListResponse)
}

func (h *handler) UserCreate(w http.ResponseWriter, r *http.Request) {
	var reqUser dto.CreateUserRequest
	ctx := r.Context()

	if err := decodeJSON(r, &reqUser); err != nil {
		createServerErr(w, fmt.Sprintf("Create User Error: %v", err))
		return
	}

	validateErr := validateCreateUserRequest(reqUser)
	if len(validateErr.Error.Fields) > 0 {
		validateErr.Error.Code = validationErrCode
		validateErr.Error.Message = "invalid request data"
		utils.WriteJSON(w, http.StatusBadRequest, validateErr)
		return
	}

	user := c.FromCreateUserRequestToUser(reqUser)

	id, err := h.serv.Create(ctx, &user)
	if err != nil {
		createServerErr(w, fmt.Sprintf("Create User Error: %v", err))
		return
	}
	user.ID = id

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

func parseIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func validateCreateUserRequest(req dto.CreateUserRequest) dto.ValidationErrorResponse {
	var validateErr dto.ValidationErrorResponse

	if req.Name == nil {
		validateErr.Error.Fields = append(validateErr.Error.Fields, requiredFieldError("name"))
	}

	if req.Email == nil {
		validateErr.Error.Fields = append(validateErr.Error.Fields, requiredFieldError("email"))
	}

	if req.Password == nil {
		validateErr.Error.Fields = append(validateErr.Error.Fields, requiredFieldError("password"))
	}

	return validateErr
}

func requiredFieldError(field string) dto.FieldError {
	return dto.FieldError{
		Field:   field,
		Message: fmt.Sprintf("Field '%s' is required", field),
	}
}
