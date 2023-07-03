package main

import (
	"be-wedding/internal/config"
	"be-wedding/internal/rest"
	"be-wedding/pkg/logger"
	"be-wedding/pkg/pgsql"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// -----------------------------------------------------------------------------------------------------------------
	// LOAD APPLICATION CONFIG FROM ENVIRONMENT VARIABLES
	// -----------------------------------------------------------------------------------------------------------------
	cfgPath := flag.String("c", "config.toml", "path to config file")
	cfg, err := config.LoadEnvFromFile(*cfgPath)
	if err != nil {
		log.Fatalln(err)
	}

	// -----------------------------------------------------------------------------------------------------------------
	// STRUCTURED LOGGER
	// -----------------------------------------------------------------------------------------------------------------
	zlogger := logger.New(cfg.Logger).With().
		Logger()

	// -----------------------------------------------------------------------------------------------------------------
	// INFRASTRUCTURE OBJECTS
	// -----------------------------------------------------------------------------------------------------------------
	// PGSQL
	sqlDB, sqlDBErr := pgsql.NewDB(cfg.PostgreSQL, zlogger)
	if sqlDBErr != nil {
		zlogger.Error().Err(sqlDBErr).Msgf("rest: main failed to construct pgsql: %s", sqlDBErr)
		return
	}

	migrate := flag.Bool("migrate", cfg.PostgreSQL.Migration, "do migration")
	if *migrate {
		if migrateErr := pgsql.Migrate(sqlDB, cfg.PostgreSQL.Database); err != nil {
			zlogger.Error().Err(migrateErr).Msgf("rest: migration failed to migrate: %s", migrateErr)
			return
		}
	}

	// -----------------------------------------------------------------------------------------------------------------
	// SERVER SETUP AND EXECUTE
	// -----------------------------------------------------------------------------------------------------------------
	restServerHandler := rest.New(cfg, zlogger, sqlDB)

	zlogger.Info().Msgf("REST Server started on port %d", cfg.API.RESTPort)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.API.RESTPort), restServerHandler)
}
