package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDBConnection do not forget to close connection
func NewDBConnection(ctx context.Context) (*pgxpool.Pool, error) {

	dbConfig := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST"),
		getEnv("DB_PORT"),
		getEnv("DB_USER"),
		getEnv("DB_PASSWORD"),
		getEnv("DB_NAME"))

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

func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return ""
}
