package exporter

import (
	"context"
	"database/sql"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type guildStorage struct {
	ID      string
	Name    string
	Icon    string
	OwnerID string
}

func createGuildsTable(ctx context.Context, db *sql.DB) bool {
	log.Info().
		Msg("discord_bot.exporter.creating_guilds_table")

	statement, err := db.PrepareContext(ctx, `
	CREATE TABLE IF NOT EXISTS "guilds" (
		id               VARCHAR (31) PRIMARY KEY,
		name             VARCHAR (255) NOT NULL,
		icon             VARCHAR (255) NULL,
		owner_id         VARCHAR (255) NULL
	);`)
	if err != nil {
		log.Error().Err(err).
			Str("step", "prepare_context").
			Msg("discord_bot.exporter.guilds_table_creating_failed")

		return false
	}

	//nolint:errcheck
	defer statement.Close()

	_, err = statement.ExecContext(ctx)
	if err != nil {
		log.Error().Err(err).
			Str("step", "exec_context").
			Msg("discord_bot.exporter.guilds_table_creating_failed")

		return false
	}

	log.Info().
		Msg("discord_bot.exporter.guilds_table_created")

	return true
}

func translateGuild(guild *discordgo.Guild) guildStorage {
	return guildStorage{
		ID:      guild.ID,
		Name:    guild.Name,
		Icon:    guild.Icon,
		OwnerID: guild.OwnerID,
	}
}

func (e *Manager) addOrUpdateGuild(ctx context.Context, guild guildStorage) bool {
	log.Info().
		Str("id", guild.ID).
		Msg("discord_bot.exporter.saving_guild")

	statement, err := e.db.PrepareContext(ctx, `REPLACE INTO guilds (id, name, icon, owner_id)
	VALUES (?, ?, ?, ?)`)
	if err != nil {
		log.Error().Err(err).
			Str("id", guild.ID).
			Str("step", "prepare_context").
			Msg("discord_bot.exporter.guild_saving_failed")

		return false
	}

	//nolint:errcheck
	defer statement.Close()

	_, err = statement.ExecContext(ctx,
		guild.ID,
		guild.Name,
		guild.Icon,
		guild.OwnerID,
	)
	if err != nil {
		log.Error().Err(err).
			Str("id", guild.ID).
			Str("step", "exec_context").
			Msg("discord_bot.exporter.guild_saving_failed")

		return false
	}

	log.Info().
		Str("id", guild.ID).
		Msg("discord_bot.exporter.guild_saved")

	return true
}
