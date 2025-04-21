//nolint:paralleltest,dupl,funlen,dupword
package welcome_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blueprintue/discord-bot/welcome"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func TestHandlers_OnMessageReactionAdd(t *testing.T) {
	var bufferLogs bytes.Buffer
	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	err = session.State.GuildAdd(&discordgo.Guild{
		ID:       "guild-123",
		Name:     guildName,
		Channels: []*discordgo.Channel{{ID: "channel-123", Name: "my-channel"}},
		Emojis:   []*discordgo.Emoji{{ID: "emoji-123", Name: "my-emoji-1"}},
		Roles:    []*discordgo.Role{{ID: "role-123", Name: "my role 1"}},
		Members: []*discordgo.Member{
			{User: &discordgo.User{ID: "user-id-456"}},
			{User: &discordgo.User{ID: "bot-123"}},
			{User: &discordgo.User{ID: "user-id-789"}, Roles: []string{"role-123"}},
		},
	})
	require.NoError(t, err)

	session.State.User = &discordgo.User{
		ID: "bot-123",
	}

	welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
		Channel:   "my-channel",
		ChannelID: "channel-123",
		Messages: []welcome.Message{
			{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
		},
	})
	require.NotNil(t, welcomeManager)

	data1, err := json.Marshal([]*discordgo.Message{})
	require.NoError(t, err)

	recorder1 := httptest.NewRecorder()
	recorder1.Header().Add("Content-Type", "application/json")
	_, err = recorder1.Write(data1)
	require.NoError(t, err)

	expectedResponse1 := recorder1.Result()
	defer expectedResponse1.Body.Close()

	data2, err := json.Marshal(discordgo.Message{ID: "123", Reactions: []*discordgo.MessageReactions{{Emoji: &discordgo.Emoji{Name: "my-emoji-1"}}}})
	require.NoError(t, err)

	recorder2 := httptest.NewRecorder()
	recorder2.Header().Add("Content-Type", "application/json")
	_, err = recorder2.Write(data2)
	require.NoError(t, err)

	expectedResponse2 := recorder2.Result()
	defer expectedResponse2.Body.Close()

	recorder3 := httptest.NewRecorder()

	expectedResponse3 := recorder3.Result()
	defer expectedResponse3.Body.Close()

	recorder4 := httptest.NewRecorder()

	expectedResponse4 := recorder4.Result()
	defer expectedResponse4.Body.Close()

	session.Client = createClient(t,
		[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3, expectedResponse4},
		[]requestTest{
			{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
			{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
				body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null,"sticker_ids":null}`},
			{method: "PUT", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/123/reactions/my-emoji-1:emoji-123/@me"},
			{method: "PUT", host: "discord.com", uri: "/api/v9/guilds/guild-123/members/user-id-456/roles/role-123"},
		},
	)

	err = welcomeManager.Run()
	require.NoError(t, err)

	bufferLogs.Reset()

	t.Run("should stop process because second argument MessageReactionAdd is nil", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionAdd(nil, nil)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"error","message":"OnMessageReactionAdd - SKIP - Reaction is nil"}`, parts[0])
		require.Equal(t, ``, parts[1])
	})

	t.Run("should stop process because MessageReactionAdd.MessageReaction is nil", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionAdd(nil, &discordgo.MessageReactionAdd{MessageReaction: nil})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"error","message":"OnMessageReactionAdd - SKIP - Reaction is nil"}`, parts[0])
		require.Equal(t, ``, parts[1])
	})

	t.Run("should stop process because Channel is not not matching", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionAdd(nil, &discordgo.MessageReactionAdd{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-id",
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-id","message_id":"","message":"Incoming Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-id","message_id":"","message":"SKIP - Channel is not matching"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should stop process because Author is the bot", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionAdd(nil, &discordgo.MessageReactionAdd{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-123",
				UserID:    "bot-123",
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"","message":"Incoming Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"","user_id":"bot-123","message":"SKIP - User is the bot"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should stop process because Message ID is not matching", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionAdd(nil, &discordgo.MessageReactionAdd{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-123",
				UserID:    "user-id-456",
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"","message":"Incoming Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"","message":"SKIP - Message is not matching"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should stop process because Emoji.Name is not matching", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionAdd(nil, &discordgo.MessageReactionAdd{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-123",
				UserID:    "user-id-456",
				MessageID: "123",
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"123","message":"Incoming Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"123","emoji":"","message":"SKIP - Emoji is not matching"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should add role to user", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionAdd(nil, &discordgo.MessageReactionAdd{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-123",
				UserID:    "user-id-456",
				MessageID: "123",
				Emoji:     discordgo.Emoji{Name: "my-emoji-1"},
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"123","message":"Incoming Message Reaction Add"}`, parts[0])
		//nolint:lll
		require.JSONEq(t, `{"level":"info","role_id":"role-123","role":"my role 1","channel_id":"channel-123","message_id":"123","user_id":"user-id-456","message":"Adding Role to User"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})
}

func TestHandlers_OnMessageReactionAdd_Errors(t *testing.T) {
	var bufferLogs bytes.Buffer
	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	err = session.State.GuildAdd(&discordgo.Guild{
		ID:       "guild-123",
		Name:     guildName,
		Channels: []*discordgo.Channel{{ID: "channel-123", Name: "my-channel"}},
		Emojis:   []*discordgo.Emoji{{ID: "emoji-123", Name: "my-emoji-1"}},
		Roles:    []*discordgo.Role{{ID: "role-123", Name: "my role 1"}},
		Members: []*discordgo.Member{
			{User: &discordgo.User{ID: "user-id-456"}},
			{User: &discordgo.User{ID: "bot-123"}},
			{User: &discordgo.User{ID: "user-id-789"}, Roles: []string{"role-123"}},
		},
	})
	require.NoError(t, err)

	session.State.User = &discordgo.User{
		ID: "bot-123",
	}

	welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
		Channel:   "my-channel",
		ChannelID: "channel-123",
		Messages: []welcome.Message{
			{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
		},
	})
	require.NotNil(t, welcomeManager)

	data1, err := json.Marshal([]*discordgo.Message{})
	require.NoError(t, err)

	recorder1 := httptest.NewRecorder()
	recorder1.Header().Add("Content-Type", "application/json")
	_, err = recorder1.Write(data1)
	require.NoError(t, err)

	expectedResponse1 := recorder1.Result()
	defer expectedResponse1.Body.Close()

	data2, err := json.Marshal(discordgo.Message{ID: "123", Reactions: []*discordgo.MessageReactions{{Emoji: &discordgo.Emoji{Name: "my-emoji-1"}}}})
	require.NoError(t, err)

	recorder2 := httptest.NewRecorder()
	recorder2.Header().Add("Content-Type", "application/json")
	_, err = recorder2.Write(data2)
	require.NoError(t, err)

	expectedResponse2 := recorder2.Result()
	defer expectedResponse2.Body.Close()

	recorder3 := httptest.NewRecorder()

	expectedResponse3 := recorder3.Result()
	defer expectedResponse3.Body.Close()

	// request failed
	recorder4 := httptest.NewRecorder()
	//nolint:goconst
	recorder4.Result().Status = "500 Internal Server Error"
	recorder4.Result().StatusCode = 500

	expectedResponse4 := recorder4.Result()
	defer expectedResponse4.Body.Close()

	session.Client = createClient(t,
		[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3, expectedResponse4},
		[]requestTest{
			{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
			{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
				body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null,"sticker_ids":null}`},
			{method: "PUT", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/123/reactions/my-emoji-1:emoji-123/@me"},
			{method: "PUT", host: "discord.com", uri: "/api/v9/guilds/guild-123/members/user-id-456/roles/role-123"},
		},
	)

	err = welcomeManager.Run()
	require.NoError(t, err)

	bufferLogs.Reset()

	welcomeManager.OnMessageReactionAdd(nil, &discordgo.MessageReactionAdd{
		MessageReaction: &discordgo.MessageReaction{
			ChannelID: "channel-123",
			UserID:    "user-id-456",
			MessageID: "123",
			Emoji:     discordgo.Emoji{Name: "my-emoji-1"},
		},
	})

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"123","message":"Incoming Message Reaction Add"}`, parts[0])
	//nolint:lll
	require.JSONEq(t, `{"level":"info","role_id":"role-123","role":"my role 1","channel_id":"channel-123","message_id":"123","user_id":"user-id-456","message":"Adding Role to User"}`, parts[1])
	//nolint:lll
	require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","role_id":"role-123","role":"my role 1","channel_id":"channel-123","message_id":"123","user_id":"user-id-456","message":"Could not add Role to User"}`, parts[2])
	require.Equal(t, ``, parts[3])
}

func TestHandlers_OnMessageReactionRemove(t *testing.T) {
	var bufferLogs bytes.Buffer
	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	err = session.State.GuildAdd(&discordgo.Guild{
		ID:       "guild-123",
		Name:     guildName,
		Channels: []*discordgo.Channel{{ID: "channel-123", Name: "my-channel"}},
		Emojis:   []*discordgo.Emoji{{ID: "emoji-123", Name: "my-emoji-1"}},
		Roles:    []*discordgo.Role{{ID: "role-123", Name: "my role 1"}},
		Members: []*discordgo.Member{
			{User: &discordgo.User{ID: "user-id-456"}},
			{User: &discordgo.User{ID: "bot-123"}},
			{User: &discordgo.User{ID: "user-id-789"}, Roles: []string{"role-123"}},
		},
	})
	require.NoError(t, err)

	session.State.User = &discordgo.User{
		ID: "bot-123",
	}

	welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
		Channel:   "my-channel",
		ChannelID: "channel-123",
		Messages: []welcome.Message{
			{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
		},
	})
	require.NotNil(t, welcomeManager)

	data1, err := json.Marshal([]*discordgo.Message{})
	require.NoError(t, err)

	recorder1 := httptest.NewRecorder()
	recorder1.Header().Add("Content-Type", "application/json")
	_, err = recorder1.Write(data1)
	require.NoError(t, err)

	expectedResponse1 := recorder1.Result()
	defer expectedResponse1.Body.Close()

	data2, err := json.Marshal(discordgo.Message{ID: "123", Reactions: []*discordgo.MessageReactions{{Emoji: &discordgo.Emoji{Name: "my-emoji-1"}}}})
	require.NoError(t, err)

	recorder2 := httptest.NewRecorder()
	recorder2.Header().Add("Content-Type", "application/json")
	_, err = recorder2.Write(data2)
	require.NoError(t, err)

	expectedResponse2 := recorder2.Result()
	defer expectedResponse2.Body.Close()

	recorder3 := httptest.NewRecorder()

	expectedResponse3 := recorder3.Result()
	defer expectedResponse3.Body.Close()

	recorder4 := httptest.NewRecorder()

	expectedResponse4 := recorder4.Result()
	defer expectedResponse4.Body.Close()

	session.Client = createClient(t,
		[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3, expectedResponse4},
		[]requestTest{
			{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
			{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
				body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null,"sticker_ids":null}`},
			{method: "PUT", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/123/reactions/my-emoji-1:emoji-123/@me"},
			{method: "DELETE", host: "discord.com", uri: "/api/v9/guilds/guild-123/members/user-id-789/roles/role-123"},
		},
	)

	err = welcomeManager.Run()
	require.NoError(t, err)

	bufferLogs.Reset()

	t.Run("should stop process because second argument MessageReactionRemove is nil", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionRemove(nil, nil)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"error","message":"OnMessageReactionRemove - SKIP - Reaction is nil"}`, parts[0])
		require.Equal(t, ``, parts[1])
	})

	t.Run("should stop process because MessageReactionRemove.MessageReaction is nil", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionRemove(nil, &discordgo.MessageReactionRemove{MessageReaction: nil})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"error","message":"OnMessageReactionRemove - SKIP - Reaction is nil"}`, parts[0])
		require.Equal(t, ``, parts[1])
	})

	t.Run("should stop process because Channel is not not matching", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionRemove(nil, &discordgo.MessageReactionRemove{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-id",
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-id","message_id":"","message":"Incoming Message Reaction Remove"}`, parts[0])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-id","message_id":"","message":"SKIP - Channel is not matching"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should stop process because Author is the bot", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionRemove(nil, &discordgo.MessageReactionRemove{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-123",
				UserID:    "bot-123",
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"","message":"Incoming Message Reaction Remove"}`, parts[0])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"","user_id":"bot-123","message":"SKIP - User is the bot"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should stop process because Message ID is not matching", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionRemove(nil, &discordgo.MessageReactionRemove{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-123",
				UserID:    "user-id-789",
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"","message":"Incoming Message Reaction Remove"}`, parts[0])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"","message":"SKIP - Message is not matching"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should stop process because Emoji ID is not matching", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionRemove(nil, &discordgo.MessageReactionRemove{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-123",
				UserID:    "user-id-789",
				MessageID: "123",
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"123","message":"Incoming Message Reaction Remove"}`, parts[0])
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"123","emoji":"","message":"SKIP - Emoji is not matching"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})

	t.Run("should add role to user", func(t *testing.T) {
		bufferLogs.Reset()

		welcomeManager.OnMessageReactionRemove(nil, &discordgo.MessageReactionRemove{
			MessageReaction: &discordgo.MessageReaction{
				ChannelID: "channel-123",
				UserID:    "user-id-789",
				MessageID: "123",
				Emoji:     discordgo.Emoji{Name: "my-emoji-1"},
			},
		})

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"123","message":"Incoming Message Reaction Remove"}`, parts[0])
		//nolint:lll
		require.JSONEq(t, `{"level":"info","role_id":"role-123","role":"my role 1","channel_id":"channel-123","message_id":"123","user_id":"user-id-789","message":"Removing Role to User"}`, parts[1])
		require.Equal(t, ``, parts[2])
	})
}

func TestHandlers_OnMessageReactionRemove_Errors(t *testing.T) {
	var bufferLogs bytes.Buffer
	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	err = session.State.GuildAdd(&discordgo.Guild{
		ID:       "guild-123",
		Name:     guildName,
		Channels: []*discordgo.Channel{{ID: "channel-123", Name: "my-channel"}},
		Emojis:   []*discordgo.Emoji{{ID: "emoji-123", Name: "my-emoji-1"}},
		Roles:    []*discordgo.Role{{ID: "role-123", Name: "my role 1"}},
		Members: []*discordgo.Member{
			{User: &discordgo.User{ID: "user-id-456"}},
			{User: &discordgo.User{ID: "bot-123"}},
			{User: &discordgo.User{ID: "user-id-789"}, Roles: []string{"role-123"}},
		},
	})
	require.NoError(t, err)

	session.State.User = &discordgo.User{
		ID: "bot-123",
	}

	welcomeManager := welcome.NewWelcomeManager(session, guildName, welcome.Configuration{
		Channel:   "my-channel",
		ChannelID: "channel-123",
		Messages: []welcome.Message{
			{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
		},
	})
	require.NotNil(t, welcomeManager)

	data1, err := json.Marshal([]*discordgo.Message{})
	require.NoError(t, err)

	recorder1 := httptest.NewRecorder()
	recorder1.Header().Add("Content-Type", "application/json")
	_, err = recorder1.Write(data1)
	require.NoError(t, err)

	expectedResponse1 := recorder1.Result()
	defer expectedResponse1.Body.Close()

	data2, err := json.Marshal(discordgo.Message{ID: "123", Reactions: []*discordgo.MessageReactions{{Emoji: &discordgo.Emoji{Name: "my-emoji-1"}}}})
	require.NoError(t, err)

	recorder2 := httptest.NewRecorder()
	recorder2.Header().Add("Content-Type", "application/json")
	_, err = recorder2.Write(data2)
	require.NoError(t, err)

	expectedResponse2 := recorder2.Result()
	defer expectedResponse2.Body.Close()

	recorder3 := httptest.NewRecorder()

	expectedResponse3 := recorder3.Result()
	defer expectedResponse3.Body.Close()

	// request failed
	recorder4 := httptest.NewRecorder()
	recorder4.Result().Status = "500 Internal Server Error"
	recorder4.Result().StatusCode = 500

	expectedResponse4 := recorder4.Result()
	defer expectedResponse4.Body.Close()

	session.Client = createClient(t,
		[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3, expectedResponse4},
		[]requestTest{
			{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
			{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
				body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null,"sticker_ids":null}`},
			{method: "PUT", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/123/reactions/my-emoji-1:emoji-123/@me"},
			{method: "DELETE", host: "discord.com", uri: "/api/v9/guilds/guild-123/members/user-id-789/roles/role-123"},
		},
	)

	err = welcomeManager.Run()
	require.NoError(t, err)

	bufferLogs.Reset()

	welcomeManager.OnMessageReactionRemove(nil, &discordgo.MessageReactionRemove{
		MessageReaction: &discordgo.MessageReaction{
			ChannelID: "channel-123",
			UserID:    "user-id-789",
			MessageID: "123",
			Emoji:     discordgo.Emoji{Name: "my-emoji-1"},
		},
	})

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","channel_id":"channel-123","message_id":"123","message":"Incoming Message Reaction Remove"}`, parts[0])
	//nolint:lll
	require.JSONEq(t, `{"level":"info","role_id":"role-123","role":"my role 1","channel_id":"channel-123","message_id":"123","user_id":"user-id-789","message":"Removing Role to User"}`, parts[1])
	//nolint:lll
	require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","role_id":"role-123","role":"my role 1","channel_id":"channel-123","message_id":"123","user_id":"user-id-789","message":"Could not remove Role to User"}`, parts[2])
	require.Equal(t, ``, parts[3])
}
