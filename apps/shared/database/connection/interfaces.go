package connection

import (
	"context"
	"database/sql"
	"shared/database/connection/migrations"
)

type DatabaseConnection interface {
	Test() error
	Close() error
	QueryRow(ctx context.Context, query string, scan func(*sql.Rows) error, args ...interface{}) error
	QueryRows(ctx context.Context, query string, scan func(*sql.Rows) error, args ...interface{}) error
	RunMigrations(migrations []migrations.Migration) error
}
