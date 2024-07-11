package postgres

import (
	"database/sql"
	"fmt"
	"myapp/config"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func New(cfg config.Postgres) (*bun.DB, error) {

	//postgresql://screenrecorder:CTVYbM735t@172.17.32.30:31506/screenrecorder?serverVersion=14&charset=utf8

	//connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?serverVersion=14&charset=utf8", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Dbname)

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Dbname, cfg.Sslmode)

	connection, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - sql.Open: %w", err)
	}
	if err := connection.Ping(); err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - Ping: %w", err)
	}
	Bun := bun.NewDB(connection, pgdialect.New())

	return Bun, nil
}
