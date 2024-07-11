package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"myapp/config"
	"myapp/internal/handler"
	"myapp/internal/usecase"
	"myapp/internal/usecase/repository"
	"myapp/pkg/postgres"
)

func Run(cfg *config.Config) {
	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("app - Run - Postgres.New")
	}
	defer pg.Close()

	postgresRepo := repository.NewScreenRecorderPostgres(pg)

	// Use case
	screenRecorderUseCases := usecase.NewScreenRecorderCases(postgresRepo)

	// HTTP Server
	router := gin.Default()

	handler.NewRouter(router, *screenRecorderUseCases, authUseCases)

	router.Run(fmt.Sprintf(":%d", cfg.Http.Port))
}
