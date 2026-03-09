package connection

import (
	"context"
	"database/sql"
)

type DatabaseConnection interface {
	Test() error
	Close() error
	QueryRow(ctx context.Context, query string, scan func(*sql.Rows) error, args ...interface{}) error
}
