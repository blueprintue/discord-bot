//nolint:paralleltest
package welcome_test

import (
	"bytes"
	"encoding/json"
	"io"
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

//nolint:funlen
func TestRun(t *testing.T) {
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

	welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
		Channel:   "my-channel",
		ChannelID: "channel-123",
		Messages: []welcome.Message{
			{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
		},
	}, guildName, session)
	require.NotNil(t, welcomeManager)

	bufferLogs.Reset()

	t.Run("should add message and add reaction because message is not found", func(t *testing.T) {
		bufferLogs.Reset()

		data1, err := json.Marshal([]*discordgo.Message{})
		require.NoError(t, err)

		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		_, err = recorder1.Write(data1)
		require.NoError(t, err)

		expectedResponse1 := recorder1.Result()
		defer expectedResponse1.Body.Close()

		data2, err := json.Marshal(discordgo.Message{ID: "123"})
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

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
					body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null,"sticker_ids":null}`},
				{method: "PUT", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/123/reactions/my-emoji-1:emoji-123/@me"},
			},
		)

		err = welcomeManager.Run()
		require.NoError(t, err)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding messages to channel"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message missing - add Message"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","message":"Sending Message to Channel"}`, parts[5])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"123","channel_id":"channel-123","channel":"my-channel","message":"Message Sent"}`, parts[6])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"123","message_title":"my title 1","emoji":"my-emoji-1:emoji-123","message":"Adding Reaction to Message"}`, parts[7])
		require.Empty(t, parts[8])
	})

	t.Run("should find message and update roles for users which added reaction on it", func(t *testing.T) {
		bufferLogs.Reset()

		data1, err := json.Marshal([]*discordgo.Message{
			{
				ID:      "101",
				Content: "this message is skipped because author is not bot",
				Author:  &discordgo.User{ID: "123"},
			},
			{
				ID:      "102",
				Content: "this message is skipped because embed is empty",
				Author:  &discordgo.User{ID: "bot-123"},
			},
			{
				ID:      "103",
				Content: "this message is skipped because embed is not same against config",
				Author:  &discordgo.User{ID: "bot-123"},
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "foo",
						Description: "bar",
						Color:       10,
					},
				},
			},
			{
				ID:      "104",
				Content: "this message is kept because embed is same against config",
				Author:  &discordgo.User{ID: "bot-123"},
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "my title 1",
						Description: "",
						Color:       0,
					},
				},
			},
		})
		require.NoError(t, err)

		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		_, err = recorder1.Write(data1)
		require.NoError(t, err)

		expectedResponse1 := recorder1.Result()
		defer expectedResponse1.Body.Close()

		data2, err := json.Marshal([]discordgo.User{
			{ID: "456", Username: "user not in discord"},
			{ID: "bot-123", Username: "user is bot"},
			{ID: "user-id-456", Username: "user lambda 456"},
			{ID: "user-id-789", Username: "user lambda 789 already has role"},
		})
		require.NoError(t, err)

		recorder2 := httptest.NewRecorder()
		recorder2.Header().Add("Content-Type", "application/json")
		_, err = recorder2.Write(data2)
		require.NoError(t, err)

		expectedResponse2 := recorder2.Result()
		defer expectedResponse2.Body.Close()

		data3, err := json.Marshal([]discordgo.User{})
		require.NoError(t, err)

		recorder3 := httptest.NewRecorder()
		recorder3.Header().Add("Content-Type", "application/json")
		_, err = recorder3.Write(data3)
		require.NoError(t, err)

		expectedResponse3 := recorder3.Result()
		defer expectedResponse3.Body.Close()

		recorder4 := httptest.NewRecorder()

		expectedResponse4 := recorder4.Result()
		defer expectedResponse4.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3, expectedResponse4},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123?after=user-id-789&limit=100"},
				{method: "PUT", host: "discord.com", uri: "/api/v9/guilds/guild-123/members/user-id-456/roles/role-123"},
			},
		)

		err = welcomeManager.Run()
		require.NoError(t, err)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding messages to channel"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"101","channel_id":"channel-123","channel":"my-channel","message":"SKIP - Message in Channel is not from bot"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"102","channel_id":"channel-123","channel":"my-channel","message":"SKIP - Message in Channel is not an embed"}`, parts[5])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message already sent -> update roles"}`, parts[6])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","emoji":"my-emoji-1:emoji-123","message":"Getting all Reactions from Message"}`, parts[7])
		require.JSONEq(t, `{"level":"error","error":"state cache not found","package":"welcome","user_id":"456","guild_id":"guild-123","message":"Could not find Member in Guild"}`, parts[8])
		require.JSONEq(t, `{"level":"info","package":"welcome","user_id":"bot-123","message":"SKIP - User is the bot"}`, parts[9])
		require.JSONEq(t, `{"level":"info","package":"welcome","role_id":"role-123","role":"my role 1","user_id":"user-id-456","username":"user lambda 456","message":"Adding Role to User"}`, parts[10])
		require.JSONEq(t, `{"level":"info","package":"welcome","user_id":"user-id-789","guild_id":"guild-123","message":"SKIP - User has already Role"}`, parts[11])
		require.JSONEq(t, `{"level":"info","package":"welcome","count_members_reacted":4,"count_members_not_found":1,"message":"Members not found in Guild"}`, parts[12])
		require.Empty(t, parts[13])
	})
}

//nolint:funlen,maintidx
func TestRun_Errors(t *testing.T) {
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

	welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
		Channel:   "my-channel",
		ChannelID: "channel-123",
		Messages: []welcome.Message{
			{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1"},
		},
	}, guildName, session)
	require.NotNil(t, welcomeManager)

	bufferLogs.Reset()

	t.Run("should return error because fetching messages from channel (ChannelMessages) return error", func(t *testing.T) {
		bufferLogs.Reset()

		// request failed
		recorder1 := httptest.NewRecorder()
		recorder1.Result().Status = internalServerError
		recorder1.Result().StatusCode = 500

		expectedResponse1 := recorder1.Result()
		defer expectedResponse1.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
			},
		)

		err = welcomeManager.Run()
		require.Error(t, err)
		require.ErrorContains(t, err, "HTTP 500 Internal Server Error")

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding messages to channel"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Could not read Messages from Channel"}`, parts[4])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message":"Could not add messages to channel"}`, parts[5])
		require.Empty(t, parts[6])
	})

	t.Run("should return error because adding message in channel (ChannelMessageSendEmbed) return error", func(t *testing.T) {
		bufferLogs.Reset()

		data1, err := json.Marshal([]*discordgo.Message{})
		require.NoError(t, err)

		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		_, err = recorder1.Write(data1)
		require.NoError(t, err)

		expectedResponse1 := recorder1.Result()
		defer expectedResponse1.Body.Close()

		// request failed
		recorder2 := httptest.NewRecorder()
		recorder2.Result().Status = internalServerError
		recorder2.Result().StatusCode = 500

		expectedResponse2 := recorder2.Result()
		defer expectedResponse2.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
					body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null,"sticker_ids":null}`},
			},
		)

		err = welcomeManager.Run()
		require.Error(t, err)
		require.ErrorContains(t, err, "HTTP 500 Internal Server Error")

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding messages to channel"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message missing - add Message"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","message":"Sending Message to Channel"}`, parts[5])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","message":"Could not send Message to Channel"}`, parts[6])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message_title":"my title 1","message":"Could not add Message"}`, parts[7])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message":"Could not add messages to channel"}`, parts[8])
		require.Empty(t, parts[9])
	})

	t.Run("should return error because adding reaction on message in channel (MessageReactionAdd) return error", func(t *testing.T) {
		bufferLogs.Reset()

		data1, err := json.Marshal([]*discordgo.Message{})
		require.NoError(t, err)

		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		_, err = recorder1.Write(data1)
		require.NoError(t, err)

		expectedResponse1 := recorder1.Result()
		defer expectedResponse1.Body.Close()

		data2, err := json.Marshal(discordgo.Message{ID: "123"})
		require.NoError(t, err)

		recorder2 := httptest.NewRecorder()
		recorder2.Header().Add("Content-Type", "application/json")
		_, err = recorder2.Write(data2)
		require.NoError(t, err)

		expectedResponse2 := recorder2.Result()
		defer expectedResponse2.Body.Close()

		// request failed
		recorder3 := httptest.NewRecorder()
		recorder3.Result().Status = internalServerError
		recorder3.Result().StatusCode = 500

		expectedResponse3 := recorder3.Result()
		defer expectedResponse3.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
					body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null,"sticker_ids":null}`},
				{method: "PUT", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/123/reactions/my-emoji-1:emoji-123/@me"},
			},
		)

		err = welcomeManager.Run()
		require.Error(t, err)
		require.ErrorContains(t, err, "HTTP 500 Internal Server Error")

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding messages to channel"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message missing - add Message"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","message":"Sending Message to Channel"}`, parts[5])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"123","channel_id":"channel-123","channel":"my-channel","message":"Message Sent"}`, parts[6])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"123","message_title":"my title 1","emoji":"my-emoji-1:emoji-123","message":"Adding Reaction to Message"}`, parts[7])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message_id":"123","emoji":"my-emoji-1:emoji-123","message":"Could not add Reaction to Message"}`, parts[8])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message_title":"my title 1","message":"Could not add Message"}`, parts[9])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message":"Could not add messages to channel"}`, parts[10])
		require.Empty(t, parts[11])
	})

	t.Run("should return error because fetching all reactions from message (MessageReactionsAll) return error", func(t *testing.T) {
		bufferLogs.Reset()

		data1, err := json.Marshal([]*discordgo.Message{
			{
				ID:      "104",
				Content: "this message is kept because embed is same against config",
				Author:  &discordgo.User{ID: "bot-123"},
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "my title 1",
						Description: "",
						Color:       0,
					},
				},
			},
		})
		require.NoError(t, err)

		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		_, err = recorder1.Write(data1)
		require.NoError(t, err)

		expectedResponse1 := recorder1.Result()
		defer expectedResponse1.Body.Close()

		recorder2 := httptest.NewRecorder()
		recorder2.Header().Add("Content-Type", "application/json")
		_, err = recorder2.WriteString("-")
		require.NoError(t, err)

		expectedResponse2 := recorder2.Result()
		defer expectedResponse2.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123?limit=100"},
			},
		)

		err = welcomeManager.Run()
		require.Error(t, err)
		require.ErrorContains(t, err, "json unmarshal")

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding messages to channel"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message already sent -> update roles"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","emoji":"my-emoji-1:emoji-123","message":"Getting all Reactions from Message"}`, parts[5])
		require.JSONEq(t, `{"level":"error","error":"json unmarshal","package":"welcome","message_id":"104","channel_id":"channel-123","channel":"my-channel","emoji":"my-emoji-1:emoji-123","message":"Could not get all Reactions"}`, parts[6])
		require.JSONEq(t, `{"level":"error","error":"json unmarshal","package":"welcome","message_title":"my title 1","message":"Could not update role belong to Message"}`, parts[7])
		require.JSONEq(t, `{"level":"error","error":"json unmarshal","package":"welcome","message":"Could not add messages to channel"}`, parts[8])
		require.Empty(t, parts[9])
	})

	t.Run("should return error because adding role to user (GuildMemberRoleAdd) return error", func(t *testing.T) {
		bufferLogs.Reset()

		data1, err := json.Marshal([]*discordgo.Message{
			{
				ID:      "104",
				Content: "this message is kept because embed is same against config",
				Author:  &discordgo.User{ID: "bot-123"},
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "my title 1",
						Description: "",
						Color:       0,
					},
				},
			},
		})
		require.NoError(t, err)

		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		_, err = recorder1.Write(data1)
		require.NoError(t, err)

		expectedResponse1 := recorder1.Result()
		defer expectedResponse1.Body.Close()

		data2, err := json.Marshal([]discordgo.User{
			{ID: "user-id-456", Username: "user lambda 456"},
		})
		require.NoError(t, err)

		recorder2 := httptest.NewRecorder()
		recorder2.Header().Add("Content-Type", "application/json")
		_, err = recorder2.Write(data2)
		require.NoError(t, err)

		expectedResponse2 := recorder2.Result()
		defer expectedResponse2.Body.Close()

		data3, err := json.Marshal([]discordgo.User{})
		require.NoError(t, err)

		recorder3 := httptest.NewRecorder()
		recorder3.Header().Add("Content-Type", "application/json")
		_, err = recorder3.Write(data3)
		require.NoError(t, err)

		expectedResponse3 := recorder3.Result()
		defer expectedResponse3.Body.Close()

		// request failed
		recorder4 := httptest.NewRecorder()
		recorder4.Result().Status = internalServerError
		recorder4.Result().StatusCode = 500

		expectedResponse4 := recorder4.Result()
		defer expectedResponse4.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3, expectedResponse4},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123?after=user-id-456&limit=100"},
				{method: "PUT", host: "discord.com", uri: "/api/v9/guilds/guild-123/members/user-id-456/roles/role-123"},
			},
		)

		err = welcomeManager.Run()
		require.Error(t, err)
		require.ErrorContains(t, err, "HTTP 500 Internal Server Error")

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding messages to channel"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message already sent -> update roles"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","emoji":"my-emoji-1:emoji-123","message":"Getting all Reactions from Message"}`, parts[5])
		require.JSONEq(t, `{"level":"info","package":"welcome","role_id":"role-123","role":"my role 1","user_id":"user-id-456","username":"user lambda 456","message":"Adding Role to User"}`, parts[6])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","role_id":"role-123","role":"my role 1","user_id":"user-id-456","username":"user lambda 456","message":"discord_bot.welcome.user_role_adding_failed"}`, parts[7])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message_title":"my title 1","message":"Could not update role belong to Message"}`, parts[8])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message":"Could not add messages to channel"}`, parts[9])
		require.Empty(t, parts[10])
	})
}

type mockRoundTripper struct {
	idxResponse     int
	test            *testing.T
	responsesMocked []*http.Response
	requestsTest    []requestTest
}

type requestTest struct {
	method string
	host   string
	uri    string
	body   string
}

func (r requestTest) assert(t *testing.T, req *http.Request) {
	t.Helper()

	require.Equal(t, r.method, req.Method)
	require.Equal(t, r.host, req.Host)
	require.Equal(t, r.uri, req.URL.RequestURI())

	if req.Method == http.MethodPost {
		bodyData, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.Equal(t, r.body, string(bodyData))
	}
}

func (rt *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.requestsTest[rt.idxResponse].assert(rt.test, req)

	resp := rt.responsesMocked[rt.idxResponse]

	rt.idxResponse++

	return resp, nil
}

func createClient(t *testing.T, responses []*http.Response, requests []requestTest) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &mockRoundTripper{
			idxResponse:     0,
			test:            t,
			responsesMocked: responses,
			requestsTest:    requests,
		},
	}
}
