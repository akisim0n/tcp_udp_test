package http

import (
	"context"
	"log"
	"net/http"
	"tcp_upd_test/database"
	"tcp_upd_test/tcp/http/handlers"
	userHand "tcp_upd_test/tcp/http/handlers/user"
	userRepo "tcp_upd_test/tcp/http/repository/user"
	userServ "tcp_upd_test/tcp/http/service/user"
	"tcp_upd_test/utils"
)

func StartHTTPServer(ctx context.Context, addr string, port int) error {

	db, err := database.NewDBConnection(ctx)
	if err != nil {
		return err
	}

	userRepository := userRepo.NewRepository(db)

	userService := userServ.NewService(userRepository)

	userHandler := userHand.NewHandler(userService)

	router := NewRouter(userHandler)

	go func() {
		<-ctx.Done()
		defer db.Close()
	}()

	log.Printf("start http server at %s:%d", addr, port)
	lisErr := http.ListenAndServe(utils.CreateServerAddress(addr, port), router)
	if lisErr != nil {
		return lisErr
	}

	return nil
}

func NewRouter(handler handlers.UserHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/user/", handler.User)
	mux.HandleFunc("/users", handler.UserList)
	mux.HandleFunc("/user/delete/", handler.UserDelete)
	mux.HandleFunc("/user/update", handler.UserUpdate)
	mux.HandleFunc("/user/create", handler.UserCreate)

	return mux
}
