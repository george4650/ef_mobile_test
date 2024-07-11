package models

import (
	"time"

	"github.com/uptrace/bun"
)

type (
	ListVideosRequest struct {
		SearchValue []SearchValue `json:"filter"`
		OrderValue  []OrderValue  `json:"order"`
		Offset      Offset        `json:"offset"`
	}

	SearchValue struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}
	OrderValue struct {
		Field string `json:"field"`
		Value bool   `json:"value"`
	}
	Offset int
)

type Video struct {
	bun.BaseModel `bun:"table:videos"`

	Name        string    `bun:"name"`
	Login       string    `bun:"login"`
	SessionId   string    `bun:"session_id"`
	CreatedAt   time.Time `bun:"created_at"`
	Fullpath    string    `bun:"fullpath"`
	MacAddr     string    `bun:"mac_addr"`
	IpAddr      string    `bun:"ip_addr"`
	GetArgs     string    `bun:"get_args"`
	FileName    string    `bun:"filename"`
	ProjectUUID string    `bun:"project_id"`
}

type ListVideos struct {
	bun.BaseModel `bun:"table:videos"`

	Login       string    `bun:"login" json:"login"`
	SessionId   string    `bun:"session_id" json:"session_id"`
	CreatedAt   time.Time `bun:"created_at" json:"created_at"`
	IpAddr      string    `bun:"ip_addr" json:"ip_addr"`
	FileName    string    `bun:"filename" json:"filename"`
	ProjectUUID string    `bun:"project_uuid" json:"project_uuid"`
	ProjectName string    `bun:"name" json:"project_name"`
}
