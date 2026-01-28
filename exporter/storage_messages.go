package exporter

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type messageStorage struct {
	ID        string
	GuildID   string
	ChannelID string
	AuthorID  string
	Content   string
	SentAt    string
	IsEmbed   bool
}

func createMessagesTable(ctx context.Context, db *sql.DB) bool {
	log.Info().
		Msg("discord_bot.exporter.creating_messages_table")

	statement, err := db.PrepareContext(ctx, `
	CREATE TABLE IF NOT EXISTS "messages" (
		id         VARCHAR (31) PRIMARY KEY,
		guild_id   VARCHAR (255) NOT NULL,
		channel_id VARCHAR (255) NOT NULL,
		author_id  VARCHAR (255) NOT NULL,
		content    TEXT NOT NULL,
		sent_at    VARCHAR (255) NOT NULL,
		is_embed   INTEGER
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
		Msg("discord_bot.exporter.messages_table_created")

	return true
}

func translateMessage(message *discordgo.Message, guildID string) messageStorage {
	gID := guildID

	if message.GuildID != "" {
		gID = message.GuildID
	}

	content := message.Content
	IsEmbed := false

	if len(message.Embeds) > 0 {
		IsEmbed = true

		tmp := make([]string, len(message.Embeds))

		for idx := range message.Embeds {
			tmp[idx] = "--- EMBED #" + strconv.Itoa(idx) + "---\n" + message.Embeds[idx].Title + "\n\n" + message.Embeds[idx].Description + "\n---"
		}

		if content != "" {
			content = content + "\n" + strings.Join(tmp, "\n")
		} else {
			content = strings.Join(tmp, "\n")
		}
	}

	if len(message.Attachments) > 0 {
		tmp := make([]string, len(message.Attachments))

		for idx := range message.Attachments {
			tmp[idx] = "ATTACHMENT #" + strconv.Itoa(idx) + ": " + message.Attachments[idx].ID + "_" + message.Attachments[idx].Filename
		}

		if content != "" {
			content = content + "\n" + strings.Join(tmp, "\n")
		} else {
			content = strings.Join(tmp, "\n")
		}
	}

	return messageStorage{
		ID:        message.ID,
		GuildID:   gID,
		ChannelID: message.ChannelID,
		AuthorID:  message.Author.ID,
		Content:   content,
		SentAt:    message.Timestamp.UTC().Format(time.DateTime),
		IsEmbed:   IsEmbed,
	}
}

func (e *Manager) addOrUpdateMessage(ctx context.Context, message messageStorage) bool {
	log.Info().
		Str("id", message.ID).
		Msg("discord_bot.exporter.saving_message")

	statement, err := e.db.PrepareContext(ctx, `REPLACE INTO messages (id, guild_id, channel_id, author_id, content, sent_at, is_embed)
	VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		log.Error().Err(err).
			Str("id", message.ID).
			Str("step", "prepare_context").
			Msg("discord_bot.exporter.message_saving_failed")

		return false
	}

	//nolint:errcheck
	defer statement.Close()

	_, err = statement.ExecContext(ctx,
		message.ID,
		message.GuildID,
		message.ChannelID,
		message.AuthorID,
		message.Content,
		message.SentAt,
		message.IsEmbed,
	)
	if err != nil {
		log.Error().Err(err).
			Str("id", message.ID).
			Str("step", "exec_context").
			Msg("discord_bot.exporter.message_saving_failed")

		return false
	}

	log.Info().
		Str("id", message.ID).
		Msg("discord_bot.exporter.message_saved")

	return true
}
