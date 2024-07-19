package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"myapp/config"
	"myapp/internal/handler"
	"myapp/internal/usecase"
	"myapp/internal/usecase/repository"
)

func Run(cfg *config.Config) {
	pg, err := repository.PostgresConnect(cfg.Postgres)
	if err != nil {
		logrus.Panicf("app - Run - repository.PostgresConnect, error - %w", err)
	}
	defer pg.Close()

	postgresRepo := repository.NewPostgresRepo(pg)

	// Use case
	screenRecorderUseCases := usecase.NewScreenRecorderCases(postgresRepo)

	// HTTP Server
	router := gin.Default()

	handler.NewRouter(router, *screenRecorderUseCases, authUseCases)

	if err = router.Run(fmt.Sprintf(":%d", cfg.Http.Port)); err != nil {
		logrus.Panicf("app - Run - router.Run, error - %v", err)
	}
}
