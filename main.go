package main

import (
	"context"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/r-scheele/escout/api"
	db "github.com/r-scheele/escout/db/sqlc"
	"github.com/r-scheele/escout/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	// runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	runGinServer(config, store)

}

// func runDBMigration(migrationURL string, dbSource string) {
// 	migration, err := migrate.New(migrationURL, dbSource)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("cannot create new migrate instance")
// 	}

// 	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
// 		log.Fatal().Err(err).Msg("failed to run migrate up")
// 	}

// 	log.Info().Msg("db migrated successfully")
// }

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
