package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"myapp/internal/models"
	"myapp/internal/usecase"
)

type PostgresRepo struct {
	connection *sql.DB
}

func NewPostgresRepo(connection *sql.DB) usecase.ScreenRecorderPostgres {
	return &PostgresRepo{
		connection: connection,
	}
}

func (db *PostgresRepo) GetVideo(ctx context.Context, videoId string) (*models.Video, error) {

	video := models.Video{}

	err := db.Bun.NewSelect().Model(&video).Where("id = ?", videoId).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Video does not exist")
		}
		log.Error().Err(err)
		return nil, fmt.Errorf("ScreenRecorderPostgres - GetVideo - db.Bun.NewSelect: %w", err)
	}

	return &video, nil
}

func (db *PostgresRepo) AddProjectsFromOracle(ctx context.Context, project models.Project) error {
	_, err := db.Bun.NewInsert().
		Model(&project).
		On("CONFLICT (project_uuid) DO UPDATE").
		Set("name = EXCLUDED.name").
		Exec(ctx)

	if err != nil {
		log.Error().Err(err)
		return fmt.Errorf("ScreenRecorderPostgres - AddProjectsFromOracle - db.Bun.NewInsert: %w", err)
	}
	return nil
}

func (db *PostgresRepo) AddVideo(ctx context.Context, video models.Video) error {

	_, err := db.Bun.NewInsert().Model(&video).Exec(ctx)
	if err != nil {
		log.Error().Err(err)
		return fmt.Errorf("ScreenRecorderPostgres - AddVideo - db.Bun.NewInsert: %w", err)
	}

	return nil
}

func (db *PostgresRepo) GetProject(ctx context.Context, uuid string) (*models.Project, error) {
	project := models.Project{}
	err := db.Bun.NewSelect().Model(&project).Where("project_uuid = ?", uuid).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Project does not exist")
		}
		log.Error().Err(err)
		return nil, fmt.Errorf("ScreenRecorderPostgres - GetProject - db.Bun.NewSelect: %w", err)
	}
	return &project, nil
}

func (db *PostgresRepo) AddProject(ctx context.Context, project models.Project) error {
	_, err := db.Bun.NewInsert().Model(&project).Exec(ctx)
	if err != nil {
		log.Error().Err(err)
		return fmt.Errorf("ScreenRecorderPostgres - AddProject - db.Bun.NewInsert: %w", err)
	}
	return nil
}

func (db *PostgresRepo) Query(ctx context.Context, searchBy []models.SearchValue, orderBy []models.OrderValue, offset models.Offset) ([]models.ListVideos, error) {

	videos := []models.ListVideos{}

	whereString := ""

	for _, search := range searchBy {
		if search.Value != "" {
			if whereString == "" {
				switch search.Field {
				case "login", "session_id", "ip_addr", "projects.name":
					whereString += fmt.Sprintf("%s = '%s'", search.Field, search.Value)
				case "created_at_start":
					whereString += fmt.Sprintf("created_at > '%s'", search.Value)
				case "created_at_end":
					whereString += fmt.Sprintf("created_at <= '%s'", search.Value)
				}
			} else {
				switch search.Field {
				case "login", "session_id", "ip_addr", "projects.name":
					whereString += fmt.Sprintf(" AND %s = '%s'", search.Field, search.Value)
				case "created_at_start":
					whereString += fmt.Sprintf(" AND created_at > '%s'", search.Value)
				case "created_at_end":
					whereString += fmt.Sprintf(" AND created_at <= '%s'", search.Value)
				}
			}
		}
	}

	log.Info().Str("whereString", whereString).Msg("")

	orderString := []string{}
	for _, order := range orderBy {
		if order.Value == true {
			orderString = append(orderString, fmt.Sprintf("%s ASC", order.Field))
		} else {
			orderString = append(orderString, fmt.Sprintf("%s DESC", order.Field))
		}
	}

	log.Info().Msgf("orderBy %v", orderString)
	log.Info().Msgf("offset %v", offset)

	var err error

	if whereString != "" {
		err = db.Bun.NewSelect().Table("videos").
			Column("login", "session_id", "created_at", "ip_addr", "filename", "projects.project_uuid", "projects.name").
			Join("left join projects").
			JoinOn("projects.project_uuid = videos.project_id").
			Where(whereString).
			Order(orderString...).
			Offset(int(offset)).
			Scan(ctx, &videos)
	} else {
		err = db.Bun.NewSelect().Table("videos").
			Column("login", "session_id", "created_at", "ip_addr", "filename", "projects.project_uuid", "projects.name").
			Join("left join projects").
			JoinOn("projects.project_uuid = videos.project_id").
			Order(orderString...).
			Offset(int(offset)).
			Scan(ctx, &videos)

	}
	return videos, err
}

func (db *PostgresRepo) ListVideos(ctx context.Context, searchBy []models.SearchValue, orderBy []models.OrderValue, offset models.Offset) ([]models.ListVideos, error) {

	videos, err := db.Query(ctx, searchBy, orderBy, offset)

	if err != nil {
		log.Error().Err(err)
		return nil, fmt.Errorf("ScreenRecorderPostgres - ListVideos - db.Query: %w", err)
	}

	return videos, nil
}

func (db *PostgresRepo) CountVideos(ctx context.Context, searchBy []models.SearchValue, orderBy []models.OrderValue, offset models.Offset) (int, error) {

	videos, err := db.Query(ctx, searchBy, orderBy, offset)
	if err != nil {
		log.Error().Err(err)
		return 0, fmt.Errorf("ScreenRecorderPostgres - CountVideos - db.Query: %w", err)
	}

	return len(videos), nil
}
