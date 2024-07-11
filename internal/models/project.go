package models

import (
	"github.com/uptrace/bun"
)

type Project struct {
	bun.BaseModel `bun:"table:projects"`

	UUID string `bun:"project_uuid"`
	Name string `bun:"name"`
}

type ProjectIdName struct {
	Id   string `db:"id"`
	Name string `db:"title"`
}
