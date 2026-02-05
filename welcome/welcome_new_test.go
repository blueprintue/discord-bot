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
	require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
	require.JSONEq(t, `{"level":"info","package":"welcome","message":"Completing configuration with session.State"}`, parts[1])
	require.JSONEq(t, `{"level":"info","package":"welcome","guild_id":"guild-123","guild":"guild-name","message":"Set GuildID"}`, parts[2])
	require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Set ChannelID"}`, parts[3])
	require.JSONEq(t, `{"level":"info","package":"welcome","role_id":"role-123","role":"my role 1","message":"Set RoleID"}`, parts[4])
	require.JSONEq(t, `{"level":"info","package":"welcome","emoji_id":"emoji-123","emoji":"my-emoji-1","message":"Set EmojiID"}`, parts[5])
	require.JSONEq(t, `{"level":"info","package":"welcome","message":"Checking configuration 2/2"}`, parts[6])
	require.Empty(t, parts[7])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","package":"welcome","message":"discord_bot.welcome.empty_channel"}`, parts[1])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","package":"welcome","message":"discord_bot.welcome.empty_messages"}`, parts[1])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","package":"welcome","message index":0,"title":"","description":"","message":"discord_bot.welcome.empty_title_description_message"}`, parts[1])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","package":"welcome","message index":0,"message":"discord_bot.welcome.empty_emoji_message"}`, parts[1])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","package":"welcome","message index":0,"message":"discord_bot.welcome.empty_role_message"}`, parts[1])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"error","package":"welcome","message index":1,"message":"discord_bot.welcome.empty_role_message"}`, parts[1])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":1,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[2])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Completing configuration with session.State"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Checking configuration 2/2"}`, parts[2])
		require.JSONEq(t, `{"level":"error","package":"welcome","guild in config":"guild-name","message":"Guild not found in Discord server"}`, parts[3])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":2,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[4])
		require.Empty(t, parts[5])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Completing configuration with session.State"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","guild_id":"guild-123","guild":"guild-name","message":"Set GuildID"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Checking configuration 2/2"}`, parts[3])
		require.JSONEq(t, `{"level":"error","package":"welcome","channel in config":"my-channel","message":"Channel not found in Discord server"}`, parts[4])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":2,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[5])
		require.Empty(t, parts[6])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Completing configuration with session.State"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","guild_id":"guild-123","guild":"guild-name","message":"Set GuildID"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Set ChannelID"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Checking configuration 2/2"}`, parts[4])
		require.JSONEq(t, `{"level":"error","package":"welcome","message #":0,"emoji in config":"my-emoji-1","message":"Emoji not found in Discord server"}`, parts[5])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":2,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[6])
		require.Empty(t, parts[7])
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"discord_bot.welcome.validating_configuration"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Completing configuration with session.State"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","guild_id":"guild-123","guild":"guild-name","message":"Set GuildID"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Set ChannelID"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","emoji_id":"emoji-123","emoji":"my-emoji-1","message":"Set EmojiID"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Checking configuration 2/2"}`, parts[5])
		require.JSONEq(t, `{"level":"error","package":"welcome","message #":0,"role in config":"my role 1","message":"Role not found in Discord server"}`, parts[6])
		require.JSONEq(t, `{"level":"error","package":"welcome","step":2,"message":"discord_bot.welcome.configuration_validation_failed"}`, parts[7])
		require.Empty(t, parts[8])
	})
}
