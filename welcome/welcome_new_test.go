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

const guildName = "guild-name"

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

	welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
		Channel:   "my-channel",
		ChannelID: "channel-123",
		Messages: []welcome.Message{
			{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
		},
	})
	require.NotNil(t, welcomeManager)

	parts := strings.Split(bufferLogs.String(), "\n")
	require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
	require.Equal(t, `{"level":"info","message":"Completing configuration with session.State"}`, parts[1])
	require.Equal(t, `{"level":"info","guild_id":"guild-123","guild":"guild-name","message":"Set GuildID"}`, parts[2])
	require.Equal(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"Set ChannelID"}`, parts[3])
	require.Equal(t, `{"level":"info","role_id":"role-123","role":"my role 1","message":"Set RoleID"}`, parts[4])
	require.Equal(t, `{"level":"info","emoji_id":"emoji-123","emoji":"my-emoji-1","message":"Set EmojiID"}`, parts[5])
	require.Equal(t, `{"level":"info","message":"Checking configuration 2/2"}`, parts[6])
	require.Equal(t, ``, parts[7])
}

//nolint:funlen
func TestNewWelcomeManager_ErrorHasValidConfigurationInFile(t *testing.T) {
	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	t.Run("should return nil because channel is empty", func(t *testing.T) {
		var bufferLogs bytes.Buffer
		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"error","message":"Channel is empty"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should return nil because messages is empty array", func(t *testing.T) {
		var bufferLogs bytes.Buffer
		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel: "my-channel",
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"error","message":"Messages is empty"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should return nil because message.title and message.description are empty", func(t *testing.T) {
		var bufferLogs bytes.Buffer
		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel:  "my-channel",
			Messages: []welcome.Message{{}},
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"error","message #":0,"title":"","description":"","message":"Title and Description is empty"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should return nil because message.emoji is empty", func(t *testing.T) {
		var bufferLogs bytes.Buffer
		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel:  "my-channel",
			Messages: []welcome.Message{{Title: "my title"}},
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"error","message #":0,"message":"Emoji is empty"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should return nil because message.role is empty", func(t *testing.T) {
		var bufferLogs bytes.Buffer
		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel:  "my-channel",
			Messages: []welcome.Message{{Title: "my title", Emoji: "my-emoji"}},
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"error","message #":0,"message":"Role is empty"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should return nil because message.role is empty on the second message", func(t *testing.T) {
		var bufferLogs bytes.Buffer
		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel: "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", Role: "my role 1"},
				{Title: "my title 2", Emoji: "my-emoji-2"},
			},
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"error","message #":1,"message":"Role is empty"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})
}

//nolint:funlen
func TestNewWelcomeManager_ErrorHasValidConfigurationAgainstDiscordServer(t *testing.T) {
	t.Run("should return nil because guild not found in Discord server", func(t *testing.T) {
		var bufferLogs bytes.Buffer
		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		session, err := discordgo.New("fake-token")
		require.NoError(t, err)

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel: "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", Role: "my role 1"},
			},
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"info","message":"Completing configuration with session.State"}`, parts[1])
		require.Equal(t, `{"level":"info","message":"Checking configuration 2/2"}`, parts[2])
		require.Equal(t, `{"level":"error","guild in config":"guild-name","message":"Guild not found in Discord server"}`, parts[3])
		require.Equal(t, ``, parts[4])
	})

	t.Run("should return nil because channel not found in Discord server", func(t *testing.T) {
		var bufferLogs bytes.Buffer
		log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

		session, err := discordgo.New("fake-token")
		require.NoError(t, err)

		session.State.Guilds = append(session.State.Guilds, &discordgo.Guild{ID: "guild-123", Name: guildName})

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel: "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", Role: "my role 1"},
			},
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"info","message":"Completing configuration with session.State"}`, parts[1])
		require.Equal(t, `{"level":"info","guild_id":"guild-123","guild":"guild-name","message":"Set GuildID"}`, parts[2])
		require.Equal(t, `{"level":"info","message":"Checking configuration 2/2"}`, parts[3])
		require.Equal(t, `{"level":"error","channel in config":"my-channel","message":"Channel not found in Discord server"}`, parts[4])
		require.Equal(t, ``, parts[5])
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

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel: "my-channel",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", Role: "my role 1"},
			},
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"info","message":"Completing configuration with session.State"}`, parts[1])
		require.Equal(t, `{"level":"info","guild_id":"guild-123","guild":"guild-name","message":"Set GuildID"}`, parts[2])
		require.Equal(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"Set ChannelID"}`, parts[3])
		require.Equal(t, `{"level":"info","message":"Checking configuration 2/2"}`, parts[4])
		require.Equal(t, `{"level":"error","message #":0,"emoji in config":"my-emoji-1","message":"Emoji not found in Discord server"}`, parts[5])
		require.Equal(t, ``, parts[6])
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

		welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
			Channel:   "my-channel",
			ChannelID: "channel-123",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
			},
		})
		require.Nil(t, welcomeManager)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Checking configuration 1/2"}`, parts[0])
		require.Equal(t, `{"level":"info","message":"Completing configuration with session.State"}`, parts[1])
		require.Equal(t, `{"level":"info","guild_id":"guild-123","guild":"guild-name","message":"Set GuildID"}`, parts[2])
		require.Equal(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"Set ChannelID"}`, parts[3])
		require.Equal(t, `{"level":"info","emoji_id":"emoji-123","emoji":"my-emoji-1","message":"Set EmojiID"}`, parts[4])
		require.Equal(t, `{"level":"info","message":"Checking configuration 2/2"}`, parts[5])
		require.Equal(t, `{"level":"error","message #":0,"role in config":"my role 1","message":"Role not found in Discord server"}`, parts[6])
		require.Equal(t, ``, parts[7])
	})
}
