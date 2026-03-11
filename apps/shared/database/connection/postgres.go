package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	migrations "shared/database/connection/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConnection struct {
	client *sql.DB
}

func CreatePostgresConnection(host, user, password, database, sslmode string, port int) (DatabaseConnection, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		database,
		sslmode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	return &PostgresConnection{
		client: db,
	}, nil
}

func (p PostgresConnection) Test() error {
	err := p.client.Ping()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

func (p PostgresConnection) Close() error {
	err := p.client.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

func (p PostgresConnection) QueryRow(ctx context.Context, query string, scan func(*sql.Rows) error, args ...interface{}) error {
	transaction, err := p.client.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	rows, err := transaction.QueryContext(ctx, query, args...)
	if err != nil {
		transaction.Rollback()
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := scan(rows); err != nil {
			transaction.Rollback()
			return fmt.Errorf("failed to scan row: %w", err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p PostgresConnection) RunMigrations(migrations []migrations.Migration) error {
	_, err := p.client.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	for _, m := range migrations {
		var exists bool
		err := p.client.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", m.Version,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %d: %w", m.Version, err)
		}

		if exists {
			log.Printf("migration %d (%s) already applied, skipping", m.Version, m.Name)
			continue
		}

		tx, err := p.client.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(m.Up); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %d (%s): %w", m.Version, m.Name, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)", m.Version, m.Name,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.Version, err)
		}

		log.Printf("migration %d (%s) applied successfully", m.Version, m.Name)
	}

	return nil
}
