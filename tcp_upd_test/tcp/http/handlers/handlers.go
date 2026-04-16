package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UserHandler interface {
	User(http.ResponseWriter, *http.Request)
	UserList(http.ResponseWriter, *http.Request)
	UserUpdate(http.ResponseWriter, *http.Request)
	UserDelete(http.ResponseWriter, *http.Request)
	UserCreate(http.ResponseWriter, *http.Request)
	Routes() chi.Router
}
