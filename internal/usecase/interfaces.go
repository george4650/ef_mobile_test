package usecase

import (
	"context"
	"myapp/internal/models"
)

type ScreenRecorderPostgres interface {
	GetProject(ctx context.Context, uuid string) (*models.Project, error)
	AddProject(ctx context.Context, project models.Project) error
	AddVideo(ctx context.Context, video models.Video) error
	GetVideo(ctx context.Context, sessionId string) (*models.Video, error)

	ListVideos(ctx context.Context, searchBy []models.SearchValue, orderBy []models.OrderValue, offset models.Offset) ([]models.ListVideos, error)
	CountVideos(ctx context.Context, searchBy []models.SearchValue, orderBy []models.OrderValue, offset models.Offset) (int, error)
	AddProjectsFromOracle(ctx context.Context, project models.Project) error
}
