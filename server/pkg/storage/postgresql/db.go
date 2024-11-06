package postgresql

import (
	"context"
	"fmt"

	// "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(dbURL string) (*pgxpool.Pool, error) {
	// conn, err := pgx.Connect(context.Background(), dbURL)
	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return dbpool, nil
}
