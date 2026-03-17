package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"tcp_upd_test/tcp/http/handlers"
	"tcp_upd_test/tcp/http/models"
	"tcp_upd_test/tcp/http/models/dto"
	"tcp_upd_test/tcp/http/service"
)

type handler struct {
	serv service.UserService
}

const (
	successMsg = "success"
	failMsg    = "fail"
)

func NewHandler(serv service.UserService) handlers.UserHandler {
	return &handler{serv: serv}
}

func (h *handler) User(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.URL.Query().Get("id")

	convId, convErr := strconv.ParseInt(id, 10, 64)
	if convErr != nil {
		log.Println(convErr)
	}

	user, getErr := h.serv.Get(ctx, convId)
	if getErr != nil {
		log.Println(getErr)
	}

	encodeErr := json.NewEncoder(w).Encode(user)
	if encodeErr != nil {
		log.Println(encodeErr)
	}
}

func (h *handler) UserUpdate(w http.ResponseWriter, r *http.Request) {
	var user models.User
	ctx := r.Context()
	decodeErr := json.NewDecoder(r.Body).Decode(&user)
	if decodeErr != nil {
		log.Println(decodeErr)
	}

	updateErr := h.serv.Update(ctx, &user)
	if updateErr != nil {
		log.Println(updateErr)
		encodeErr := json.NewEncoder(w).Encode(failMsg)
		if encodeErr != nil {
			log.Println(encodeErr)
		}
	}
	encodeErr := json.NewEncoder(w).Encode(successMsg)
	if encodeErr != nil {
		log.Println(encodeErr)
		encodeErr = json.NewEncoder(w).Encode(failMsg)
		if encodeErr != nil {
			log.Println(encodeErr)
		}
	}
}

func (h *handler) UserDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.URL.Query().Get("id")

	idInt64, convErr := strconv.ParseInt(idStr, 10, 64)
	if convErr != nil {
		log.Println(convErr)
		encodeErr := json.NewEncoder(w).Encode(failMsg)
		if encodeErr != nil {
			log.Println(encodeErr)
		}
	}

	deleteErr := h.serv.Delete(ctx, idInt64)
	if deleteErr != nil {
		log.Println(deleteErr)
		encodeErr := json.NewEncoder(w).Encode(failMsg)
		if encodeErr != nil {
			log.Println(encodeErr)
		}
	}
}

func (h *handler) UserList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, getAllErr := h.serv.GetAll(ctx)
	if getAllErr != nil {
		log.Println(getAllErr)
		encodeErr := json.NewEncoder(w).Encode(failMsg)
		if encodeErr != nil {
			log.Println(encodeErr)
		}
	}

	encodeErr := json.NewEncoder(w).Encode(users)
	if encodeErr != nil {
		log.Println(encodeErr)
	}
}

func (h *handler) UserCreate(w http.ResponseWriter, r *http.Request) {
	var reqUser dto.CreateUserRequest
	ctx := r.Context()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	defer r.Body.Close()

	if err := decoder.Decode(&reqUser); err != nil {
		log.Println(err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if reqUser.Name == nil || reqUser.Email == nil || reqUser.Password == nil {
		http.Error(w, "name and email are required", http.StatusBadRequest)
		return
	}

	userData := models.UserData{
		Name:     *reqUser.Name,
		Email:    *reqUser.Email,
		Password: *reqUser.Password,
	}

	if reqUser.Surname != nil {
		userData.Surname = reqUser.Surname
	}
	if reqUser.Age != nil {
		userData.Age = reqUser.Age
	}

	userId, createErr := h.serv.Create(ctx, &models.User{Data: userData})
	if createErr != nil {
		log.Println(createErr)
		http.Error(w, fmt.Sprintf("error during creation: %s", createErr.Error()), http.StatusBadRequest)
	}

	respUser := dto.CreateUserResponse{Id: userId}

	w.WriteHeader(http.StatusCreated)
	encodeErr := json.NewEncoder(w).Encode(respUser)
	if encodeErr != nil {
		log.Println(encodeErr)
		http.Error(w, fmt.Sprintf("error during encoding: %s", encodeErr.Error()), http.StatusBadRequest)
	}
}
