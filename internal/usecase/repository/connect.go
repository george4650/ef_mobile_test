package repository

import (
	"database/sql"
	"fmt"
	"myapp/config"
)

func PostgresConnect(cfg config.Postgres) (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Dbname, cfg.Sslmode)

	connection, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - sql.Open: %w", err)
	}
	if err = connection.Ping(); err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - Ping: %w", err)
	}

	return connection, nil
}
