package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Egalam47"
	dbname   = "test_http"
)

// NewDBConnection do not forget to close connection
func NewDBConnection(ctx context.Context) (*pgxpool.Pool, error) {
	dbConfig := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

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
