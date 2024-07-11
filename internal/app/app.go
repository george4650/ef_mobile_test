package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"myapp/config"
	"myapp/internal/handler"
	"myapp/internal/usecase"
	"myapp/internal/usecase/repository"
	"myapp/pkg/ldap"
	"myapp/pkg/oracle"
	"myapp/pkg/postgres"
	"myapp/pkg/samba"
)

func Run(cfg *config.Config) {

	// Ldap
	ldapConn, err := ldap.ConnectToServerLDAP(cfg.Ldap)
	if err != nil {
		log.Error().Err(err).Msg("app - Run - ldap.ConnectToServerLDAP")
	}
	defer ldapConn.Close()

	// Samba
	smbSession, err := samba.New(cfg.Samba)
	if err != nil {
		log.Error().Err(err).Msg("app - Run - samba.New")
	}
	defer smbSession.Logoff()

	// Repository
	ora, err := oracle.New(cfg.Oracle)
	if err != nil {
		log.Error().Err(err).Msg("app - Run - oracle.New")
	}
	defer ora.Close()

	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("app - Run - Postgres.New")
	}
	defer pg.Close()

	callDetailsPostgresRepo := repository.NewScreenRecorderPostgres(pg)

	// Use case
	authUseCases := usecase.NewAuthUseCases(ldapConn, cfg.Auth.Key)
	screenRecorderUseCases := usecase.NewScreenRecorderCases(cfg.Links.Audio, smbSession, callDetailsPostgresRepo, callDetailsOracleRepo)

	// HTTP Server
	router := gin.Default()

	handler.NewRouter(router, *screenRecorderUseCases, authUseCases)

	router.Run(fmt.Sprintf("localhost:%d", cfg.Http.Port))

}
