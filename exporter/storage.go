package exporter

import (
	"context"
	"database/sql"
	"path"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

func initializeDatabase(outputPath string, databaseFilename string) *sql.DB {
	ctx := context.Background()

	db := openFile(ctx, outputPath, databaseFilename)
	if db == nil {
		return nil
	}

	created := createGuildsTable(ctx, db)
	if !created {
		//nolint:errcheck
		defer db.Close()

		return nil
	}

	created = createUsersTable(ctx, db)
	if !created {
		//nolint:errcheck
		defer db.Close()

		return nil
	}

	created = createChannelsTable(ctx, db)
	if !created {
		//nolint:errcheck
		defer db.Close()

		return nil
	}

	created = createMessagesTable(ctx, db)
	if !created {
		//nolint:errcheck
		defer db.Close()

		return nil
	}

	return db
}

func openFile(ctx context.Context, outputPath string, databaseFilename string) *sql.DB {
	var (
		dbName  string
		version string
	)

	if databaseFilename == ":memory:" {
		dbName = databaseFilename
	} else {
		dbName = path.Join(outputPath, databaseFilename)
	}

	log.Info().
		Str("database", dbName).
		Msg("discord_bot.exporter.creating_database")

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Error().Err(err).
			Str("database", dbName).
			Msg("discord_bot.exporter.database_creating_failed")

		return nil
	}

	log.Info().
		Str("database", dbName).
		Msg("discord_bot.exporter.database_created")

	log.Info().
		Str("database", dbName).
		Msg("discord_bot.exporter.checking_database_version")

	err = db.QueryRowContext(ctx, "SELECT SQLITE_VERSION()").Scan(&version)
	if err != nil {
		//nolint:errcheck
		defer db.Close()

		log.Error().Err(err).
			Str("database", dbName).
			Msg("discord_bot.exporter.database_version_checking_failed")

		return nil
	}

	log.Info().
		Str("database", dbName).
		Str("version", version).
		Msg("discord_bot.exporter.database_version_checked")

	return db
}
