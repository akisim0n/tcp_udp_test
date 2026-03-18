package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"tcp_upd_test/database"
	"tcp_upd_test/tcp/http/handlers"
	userHand "tcp_upd_test/tcp/http/handlers/user"
	userRepo "tcp_upd_test/tcp/http/repository/user"
	userServ "tcp_upd_test/tcp/http/service/user"
	"tcp_upd_test/utils"

	"github.com/joho/godotenv"
)

func StartHTTPServer(ctx context.Context, addr string, port int) error {

	envErr := godotenv.Load("http.env")
	if envErr != nil {
		log.Fatal("Error loading http.env file")
	}

	db, err := database.NewDBConnection(ctx)
	if err != nil {
		return fmt.Errorf("db connection error: %w", err)
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
		return fmt.Errorf("start http server err:%v", lisErr)
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
