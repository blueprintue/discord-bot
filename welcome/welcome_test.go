//nolint:paralleltest
package welcome_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/blueprintue/discord-bot/welcome"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

const (
	guildName           = "guild-name"
	internalServerError = "500 Internal Server Error"
)

func TestNewWelcomeManager(t *testing.T) {
	var bufferLogs bytes.Buffer

	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	session.State.Guilds = append(session.State.Guilds, &discordgo.Guild{
		ID:       "guild-123",
		Name:     guildName,
		Channels: []*discordgo.Channel{{ID: "channel-123", Name: "my-channel"}},
		Emojis:   []*discordgo.Emoji{{ID: "emoji-123", Name: "my-emoji-1"}},
		Roles:    []*discordgo.Role{{ID: "role-123", Name: "my role 1"}},
	})

	welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
		Channel:   "my-channel",
		Messages: []welcome.Message{
			{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
		},
	}, guildName, session)
	require.NotNil(t, welcomeManager)

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
	require.JSONEq(t, `{"level":"info","guild_id":"guild-123","guild":"guild-name","message":"discord_bot.welcome.set_guild_id"}`, parts[1])
	require.JSONEq(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"discord_bot.welcome.set_channel_id"}`, parts[2])
	require.JSONEq(t, `{"level":"info","message index":0,"role_id":"role-123","role":"my role 1","message":"discord_bot.welcome.set_role_id"}`, parts[3])
	require.JSONEq(t, `{"level":"info","message index":0,"emoji_id":"emoji-123","emoji":"my-emoji-1","message":"discord_bot.welcome.set_emoji_id"}`, parts[4])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.configuration_validated"}`, parts[5])
	require.Empty(t, parts[6])
}

//nolint:funlen
func TestNewWelcomeManager_ErrorHasValidConfigurationInFile(t *testing.T) {
	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	t.Run("should return nil because channel is empty", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","message":"discord_bot.welcome.configuration_empty_channel"}`, parts[1])
		require.JSONEq(t, `{"level":"error","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
		require.Empty(t, parts[3])
	})

	t.Run("should return nil because messages is empty array", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel: "my-channel",
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","message":"discord_bot.welcome.configuration_empty_messages"}`, parts[1])
		require.JSONEq(t, `{"level":"error","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
		require.Empty(t, parts[3])
	})

	t.Run("should return nil because message.title and message.description are empty", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel:  "my-channel",
			Messages: []welcome.Message{{}},
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","message index":0,"title":"","description":"","message":"discord_bot.welcome.configuration_empty_title_description_message"}`, parts[1])
		require.JSONEq(t, `{"level":"error","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
		require.Empty(t, parts[3])
	})

	t.Run("should return nil because message.emoji is empty", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel:  "my-channel",
			Messages: []welcome.Message{{Title: "my title"}},
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","message index":0,"message":"discord_bot.welcome.configuration_empty_emoji_message"}`, parts[1])
		require.JSONEq(t, `{"level":"error","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
		require.Empty(t, parts[3])
	})

	t.Run("should return nil because message.role is empty", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel:  "my-channel",
			Messages: []welcome.Message{{Title: "my title", Emoji: "my-emoji"}},
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","message index":0,"message":"discord_bot.welcome.configuration_empty_role_message"}`, parts[1])
		require.JSONEq(t, `{"level":"error","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
		require.Empty(t, parts[3])
	})

	t.Run("should return nil because message.role is empty on the second message", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel: "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", Role: "my role 1"},
				{Title: "my title 2", Emoji: "my-emoji-2"},
			},
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","message index":1,"message":"discord_bot.welcome.configuration_empty_role_message"}`, parts[1])
		require.JSONEq(t, `{"level":"error","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
		require.Empty(t, parts[3])
	})
}

//nolint:funlen
func TestNewWelcomeManager_ErrorHasValidConfigurationAgainstDiscordServer(t *testing.T) {
	t.Run("should return nil because guild not found in Discord server", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		session, err := discordgo.New("fake-token")
		require.NoError(t, err)

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel: "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", Role: "my role 1"},
			},
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","guild":"guild-name","message":"discord_bot.welcome.configuration_guild_missed"}`, parts[1])
		require.JSONEq(t, `{"level":"error","step":2,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
		require.Empty(t, parts[3])
	})

	t.Run("should return nil because channel not found in Discord server", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		session, err := discordgo.New("fake-token")
		require.NoError(t, err)

		session.State.Guilds = append(session.State.Guilds, &discordgo.Guild{ID: "guild-123", Name: guildName})

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel: "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", Role: "my role 1"},
			},
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"info","guild_id":"guild-123","guild":"guild-name","message":"discord_bot.welcome.set_guild_id"}`, parts[1])
		require.JSONEq(t, `{"level":"error","channel":"my-channel","message":"discord_bot.welcome.configuration_channel_missed"}`, parts[2])
		require.JSONEq(t, `{"level":"error","step":2,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[3])
		require.Empty(t, parts[4])
	})

	t.Run("should return nil because emoji not found in Discord server", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		session, err := discordgo.New("fake-token")
		require.NoError(t, err)

		session.State.Guilds = append(session.State.Guilds, &discordgo.Guild{
			ID:       "guild-123",
			Name:     guildName,
			Channels: []*discordgo.Channel{{ID: "channel-123", Name: "my-channel"}},
		})

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel: "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", Role: "my role 1"},
			},
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"info","guild_id":"guild-123","guild":"guild-name","message":"discord_bot.welcome.set_guild_id"}`, parts[1])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"discord_bot.welcome.set_channel_id"}`, parts[2])
		require.JSONEq(t, `{"level":"error","message index":0,"emoji":"my-emoji-1","message":"discord_bot.welcome.configuration_emoji_missed"}`, parts[3])
		require.JSONEq(t, `{"level":"error","step":2,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[4])
		require.Empty(t, parts[5])
	})

	t.Run("should return nil because role not found in Discord server", func(t *testing.T) {
		var bufferLogs bytes.Buffer

		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		session, err := discordgo.New("fake-token")
		require.NoError(t, err)

		session.State.Guilds = append(session.State.Guilds, &discordgo.Guild{
			ID:       "guild-123",
			Name:     guildName,
			Channels: []*discordgo.Channel{{ID: "channel-123", Name: "my-channel"}},
			Emojis:   []*discordgo.Emoji{{ID: "emoji-123", Name: "my-emoji-1"}},
		})

		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel:   "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
			},
		}, guildName, session)
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"info","guild_id":"guild-123","guild":"guild-name","message":"discord_bot.welcome.set_guild_id"}`, parts[1])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"discord_bot.welcome.set_channel_id"}`, parts[2])
		require.JSONEq(t, `{"level":"info","message index":0,"emoji_id":"emoji-123","emoji":"my-emoji-1","message":"discord_bot.welcome.set_emoji_id"}`, parts[3])
		require.JSONEq(t, `{"level":"error","message index":0,"role":"my role 1","message":"discord_bot.welcome.configuration_role_missed"}`, parts[4])
		require.JSONEq(t, `{"level":"error","step":2,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[5])
		require.Empty(t, parts[6])
	})
}
