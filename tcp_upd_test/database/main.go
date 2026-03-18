package database

import (
	"context"
	"fmt"
	"log"
	"tcp_upd_test/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDBConnection do not forget to close connection
func NewDBConnection(ctx context.Context) (*pgxpool.Pool, error) {

	dbConfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		utils.GetEnvParam("DB_HOST"),
		utils.GetEnvParam("DB_PORT"),
		utils.GetEnvParam("DB_USER"),
		utils.GetEnvParam("DB_PASSWORD"),
		utils.GetEnvParam("DB_NAME"))

	pgx, pgxErr := pgxpool.New(ctx, dbConfig)
	if pgxErr != nil {
		log.Println(pgxErr)
	}

	pingErr := pgx.Ping(ctx)
	if pingErr != nil {
		return nil, pingErr
	}

	return pgx, nil
}
