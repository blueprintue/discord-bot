package exporter

import (
	"context"
	"database/sql"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type userStorage struct {
	ID            string
	Username      string
	Discriminator string
	GlobalName    string
	Avatar        string
}

func createUsersTable(ctx context.Context, db *sql.DB) bool {
	log.Info().
		Msg("discord_bot.exporter.creating_users_table")

	statement, err := db.PrepareContext(ctx, `
	CREATE TABLE IF NOT EXISTS "users" (
		id            VARCHAR (16) PRIMARY KEY,
		username      VARCHAR (255) NOT NULL,
		discriminator VARCHAR (255) NOT NULL,
		global_name   VARCHAR (255) NOT NULL,
		avatar        VARCHAR (255) NOT NULL
	);`)
	if err != nil {
		log.Error().Err(err).
			Str("step", "prepare_context").
			Msg("discord_bot.exporter.users_table_creating_failed")

		return false
	}

	//nolint:errcheck
	defer statement.Close()

	_, err = statement.ExecContext(ctx)
	if err != nil {
		log.Error().Err(err).
			Str("step", "exec_context").
			Msg("discord_bot.exporter.users_table_creating_failed")

		return false
	}

	log.Info().
		Msg("discord_bot.exporter.users_table_created")

	return true
}

func translateUser(user *discordgo.User) userStorage {
	return userStorage{
		ID:            user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		GlobalName:    user.GlobalName,
		Avatar:        user.Avatar,
	}
}

func (e *Manager) addOrUpdateUser(ctx context.Context, user userStorage) bool {
	log.Info().
		Str("id", user.ID).
		Msg("discord_bot.exporter.saving_user")

	statement, err := e.db.PrepareContext(ctx, `REPLACE INTO users (id, username, discriminator, global_name, avatar)
	VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		log.Error().Err(err).
			Str("id", user.ID).
			Str("step", "prepare_context").
			Msg("discord_bot.exporter.user_saving_failed")

		return false
	}

	//nolint:errcheck
	defer statement.Close()

	_, err = statement.ExecContext(ctx,
		user.ID,
		user.Username,
		user.Discriminator,
		user.GlobalName,
		user.Avatar,
	)
	if err != nil {
		log.Error().Err(err).
			Str("id", user.ID).
			Str("step", "exec_context").
			Msg("discord_bot.exporter.user_saving_failed")

		return false
	}

	log.Info().
		Str("id", user.ID).
		Msg("discord_bot.exporter.user_saved")

	return true
}
