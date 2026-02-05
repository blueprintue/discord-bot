//nolint:paralleltest,dupl
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

//nolint:funlen,maintidx
func TestRunPurge(t *testing.T) {
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

	t.Run("should not purge because CanPurgeReactions is false", func(t *testing.T) {
		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel:   "my-channel",
			ChannelID: "channel-123",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1",
					CanPurgeReactions: false, PurgeThresholdMembersReacted: 1, PurgeBelowCountMembersNotInGuild: 10},
			},
		}, guildName, session)
		require.NotNil(t, welcomeManager)
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message already sent -> update roles"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","emoji":"my-emoji-1:emoji-123","message":"Getting all Reactions from Message"}`, parts[5])
		require.JSONEq(t, `{"level":"error","package":"welcome","error":"state cache not found","user_id":"456","guild_id":"guild-123","message":"Could not find Member in Guild"}`, parts[6])
		require.JSONEq(t, `{"level":"info","package":"welcome","role_id":"role-123","role":"my role 1","user_id":"user-id-456","username":"user lambda 456","message":"Adding Role to User"}`, parts[7])
		require.JSONEq(t, `{"level":"info","package":"welcome","count_members_reacted":4,"count_members_not_found":1,"message":"Members not found in Guild"}`, parts[8])
		require.Empty(t, parts[9])
	})

	t.Run("should not purge because PurgeThresholdMembersReacted is not equal or greater than members reacted", func(t *testing.T) {
		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel:   "my-channel",
			ChannelID: "channel-123",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1",
					CanPurgeReactions: true, PurgeThresholdMembersReacted: 5, PurgeBelowCountMembersNotInGuild: 10},
			},
		}, guildName, session)
		require.NotNil(t, welcomeManager)
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message already sent -> update roles"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","emoji":"my-emoji-1:emoji-123","message":"Getting all Reactions from Message"}`, parts[5])
		require.JSONEq(t, `{"level":"error","package":"welcome","error":"state cache not found","user_id":"456","guild_id":"guild-123","message":"Could not find Member in Guild"}`, parts[6])
		require.JSONEq(t, `{"level":"info","package":"welcome","role_id":"role-123","role":"my role 1","user_id":"user-id-456","username":"user lambda 456","message":"Adding Role to User"}`, parts[7])
		require.JSONEq(t, `{"level":"info","package":"welcome","count_members_reacted":4,"count_members_not_found":1,"message":"Members not found in Guild"}`, parts[8])
		require.Empty(t, parts[9])
	})

	t.Run("should not purge because count members not in discord is equal or greater than PurgeBelowCountMembersNotInGuild", func(t *testing.T) {
		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel:   "my-channel",
			ChannelID: "channel-123",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1",
					CanPurgeReactions: true, PurgeThresholdMembersReacted: 4, PurgeBelowCountMembersNotInGuild: 0},
			},
		}, guildName, session)
		require.NotNil(t, welcomeManager)
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
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message already sent -> update roles"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","emoji":"my-emoji-1:emoji-123","message":"Getting all Reactions from Message"}`, parts[5])
		require.JSONEq(t, `{"level":"error","error":"state cache not found","package":"welcome","user_id":"456","guild_id":"guild-123","message":"Could not find Member in Guild"}`, parts[6])
		require.JSONEq(t, `{"level":"info","package":"welcome","role_id":"role-123","role":"my role 1","user_id":"user-id-456","username":"user lambda 456","message":"Adding Role to User"}`, parts[7])
		require.JSONEq(t, `{"level":"info","package":"welcome","count_members_reacted":4,"count_members_not_found":1,"message":"Members not found in Guild"}`, parts[8])
		require.Empty(t, parts[9])
	})

	t.Run("should do purge", func(t *testing.T) {
		welcomeManager := welcome.NewWelcomeManager(welcome.Configuration{
			Channel:   "my-channel",
			ChannelID: "channel-123",
			Messages: []welcome.Message{
				{Title: "my title 1", Emoji: "my-emoji-1", EmojiID: "emoji-123", Role: "my role 1",
					CanPurgeReactions: true, PurgeThresholdMembersReacted: 1, PurgeBelowCountMembersNotInGuild: 10},
			},
		}, guildName, session)
		require.NotNil(t, welcomeManager)
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
			{ID: "456", Username: "user 1 not in discord"},
			{ID: "678", Username: "user 2 not in discord"},
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

		recorder5 := httptest.NewRecorder()

		expectedResponse5 := recorder5.Result()
		defer expectedResponse5.Body.Close()

		// request failed
		recorder6 := httptest.NewRecorder()

		recorder6.Result().Status = internalServerError
		recorder6.Result().StatusCode = 500

		expectedResponse6 := recorder6.Result()
		defer expectedResponse6.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3, expectedResponse4, expectedResponse5, expectedResponse6},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123?after=user-id-789&limit=100"},
				{method: "PUT", host: "discord.com", uri: "/api/v9/guilds/guild-123/members/user-id-456/roles/role-123"},
				{method: "DELETE", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123/456"},
				{method: "DELETE", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/104/reactions/my-emoji-1:emoji-123/678"},
			},
		)

		err = welcomeManager.Run()
		require.NoError(t, err)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Adding messages to channel"}`, parts[2])
		require.JSONEq(t, `{"level":"info","package":"welcome","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_title":"my title 1","message":"Message already sent -> update roles"}`, parts[4])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","emoji":"my-emoji-1:emoji-123","message":"Getting all Reactions from Message"}`, parts[5])
		require.JSONEq(t, `{"level":"error","error":"state cache not found","package":"welcome","user_id":"456","guild_id":"guild-123","message":"Could not find Member in Guild"}`, parts[6])
		require.JSONEq(t, `{"level":"error","error":"state cache not found","package":"welcome","user_id":"678","guild_id":"guild-123","message":"Could not find Member in Guild"}`, parts[7])
		require.JSONEq(t, `{"level":"info","package":"welcome","role_id":"role-123","role":"my role 1","user_id":"user-id-456","username":"user lambda 456","message":"Adding Role to User"}`, parts[8])
		require.JSONEq(t, `{"level":"info","package":"welcome","count_members_reacted":5,"count_members_not_found":2,"message":"Members not found in Guild"}`, parts[9])
		require.JSONEq(t, `{"level":"info","package":"welcome","message":"Do purge"}`, parts[10])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","emoji":"my-emoji-1:emoji-123","user_id":"456","message":"Removing Reaction on Message for User"}`, parts[11])
		require.JSONEq(t, `{"level":"info","package":"welcome","message_id":"104","emoji":"my-emoji-1:emoji-123","user_id":"678","message":"Removing Reaction on Message for User"}`, parts[12])
		require.JSONEq(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","package":"welcome","message_id":"104","emoji":"my-emoji-1:emoji-123","user_id":"678","message":"Could not remove Reaction"}`, parts[13])
		require.Empty(t, parts[14])
	})
}
