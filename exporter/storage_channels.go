package exporter

import (
	"context"
	"database/sql"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type channelStorage struct {
	ID       string
	GuildID  string
	Name     string
	Topic    string
	Type     string
	ParentID string
	OwnerID  string
	Position int
}

func createChannelsTable(ctx context.Context, db *sql.DB) bool {
	log.Info().
		Msg("discord_bot.exporter.creating_channels_table")

	statement, err := db.PrepareContext(ctx, `
	CREATE TABLE IF NOT EXISTS "channels" (
		id        VARCHAR (31) PRIMARY KEY,
		guild_id  VARCHAR (255) NOT NULL,
		name      VARCHAR (255) NOT NULL,
		topic     TEXT NULL,
		type      VARCHAR (255) NOT NULL,
		position  INTEGER NOT NULL,
		parent_id VARCHAR (255) NULL,
		owner_id  VARCHAR (255) NOT NULL
	);`)
	if err != nil {
		log.Error().Err(err).
			Str("step", "prepare_context").
			Msg("discord_bot.exporter.channels_table_creating_failed")

		return false
	}

	//nolint:errcheck
	defer statement.Close()

	_, err = statement.ExecContext(ctx)
	if err != nil {
		log.Error().Err(err).
			Str("step", "exec_context").
			Msg("discord_bot.exporter.channels_table_creating_failed")

		return false
	}

	log.Info().
		Msg("discord_bot.exporter.channels_table_created")

	return true
}

func translateChannel(channel *discordgo.Channel) channelStorage {
	return channelStorage{
		ID:       channel.ID,
		GuildID:  channel.GuildID,
		Name:     channel.Name,
		Topic:    channel.Topic,
		Type:     translateChannelType(channel.Type),
		Position: channel.Position,
		ParentID: channel.ParentID,
		OwnerID:  channel.OwnerID,
	}
}

//nolint:cyclop
func translateChannelType(channelType discordgo.ChannelType) string {
	switch channelType {
	case discordgo.ChannelTypeGuildText:
		return "guild_text"
	case discordgo.ChannelTypeDM:
		return "dm"
	case discordgo.ChannelTypeGuildVoice:
		return "guild_voice"
	case discordgo.ChannelTypeGroupDM:
		return "group_dm"
	case discordgo.ChannelTypeGuildCategory:
		return "guild_category"
	case discordgo.ChannelTypeGuildNews:
		return "guild_news"
	case discordgo.ChannelTypeGuildStore:
		return "guild_store"
	case discordgo.ChannelTypeGuildNewsThread:
		return "guild_news_thread"
	case discordgo.ChannelTypeGuildPublicThread:
		return "guild_public_thread"
	case discordgo.ChannelTypeGuildPrivateThread:
		return "guild_private_thread"
	case discordgo.ChannelTypeGuildStageVoice:
		return "guild_stage_voice"
	case discordgo.ChannelTypeGuildDirectory:
		return "guild_directory"
	case discordgo.ChannelTypeGuildForum:
		return "guild_forum"
	case discordgo.ChannelTypeGuildMedia:
		return "guild_media"
	default:
		return ""
	}
}

func (e *Manager) addOrUpdateChannel(ctx context.Context, channel channelStorage) bool {
	log.Info().
		Str("id", channel.ID).
		Msg("discord_bot.exporter.saving_channel")

	statement, err := e.db.PrepareContext(ctx, `REPLACE INTO channels (id, guild_id, name, topic, type, position, parent_id, owner_id)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		log.Error().Err(err).
			Str("id", channel.ID).
			Str("step", "prepare_context").
			Msg("discord_bot.exporter.channel_saving_failed")

		return false
	}

	//nolint:errcheck
	defer statement.Close()

	_, err = statement.ExecContext(ctx,
		channel.ID,
		channel.GuildID,
		channel.Name,
		channel.Topic,
		channel.Type,
		channel.Position,
		channel.ParentID,
		channel.OwnerID,
	)
	if err != nil {
		log.Error().Err(err).
			Str("id", channel.ID).
			Str("step", "exec_context").
			Msg("discord_bot.exporter.channel_saving_failed")

		return false
	}

	log.Info().
		Str("id", channel.ID).
		Msg("discord_bot.exporter.channel_saved")

	return true
}
