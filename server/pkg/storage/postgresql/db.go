package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewDB(dbURL string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return conn, nil
}
