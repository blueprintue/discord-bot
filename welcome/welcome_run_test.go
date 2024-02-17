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

	session.State.Guilds = append(session.State.Guilds, &discordgo.Guild{
		ID:       "guild-123",
		Name:     guildName,
		Channels: []*discordgo.Channel{{ID: "channel-123", Name: "my-channel"}},
		Emojis:   []*discordgo.Emoji{{ID: "emoji-123", Name: "my-emoji-1"}},
		Roles:    []*discordgo.Role{{ID: "role-123", Name: "my role 1"}},
	})

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
					body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null}`},
				{method: "PUT", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/123/reactions/my-emoji-1:emoji-123/@me"},
			},
		)

		err = welcomeManager.Run()
		require.NoError(t, err)

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.Equal(t, `{"level":"info","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.Equal(t, `{"level":"info","message":"Adding messages to channel"}`, parts[2])
		require.Equal(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.Equal(t, `{"level":"info","message_title":"my title 1","message":"Message missing - add Message"}`, parts[4])
		//nolint:lll
		require.Equal(t, `{"level":"info","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","message":"Sending Message to Channel"}`, parts[5])
		require.Equal(t, `{"level":"info","message_id":"123","channel_id":"channel-123","channel":"my-channel","message":"Message Sent"}`, parts[6])
		//nolint:lll
		require.Equal(t, `{"level":"info","message_id":"123","message_title":"my title 1","emoji":"my-emoji-1:emoji-123","message":"Adding Reaction to Message"}`, parts[7])
		require.Equal(t, ``, parts[8])
	})
}

//nolint:funlen
func TestRun_Errors(t *testing.T) {
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

	bufferLogs.Reset()

	//nolint:bodyclose,goconst
	t.Run("should return error because fetching messages from channel (ChannelMessages) return error", func(t *testing.T) {
		bufferLogs.Reset()

		// request failed
		recorder1 := httptest.NewRecorder()
		recorder1.Result().Status = "500 Internal Server Error"
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
		require.Equal(t, `{"level":"info","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.Equal(t, `{"level":"info","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.Equal(t, `{"level":"info","message":"Adding messages to channel"}`, parts[2])
		require.Equal(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		//nolint:lll
		require.Equal(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","channel_id":"channel-123","channel":"my-channel","message":"Could not read Messages from Channel"}`, parts[4])
		require.Equal(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","message":"Could not add messages to channel"}`, parts[5])
		require.Equal(t, ``, parts[6])
	})

	//nolint:bodyclose
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
		recorder2.Result().Status = "500 Internal Server Error"
		recorder2.Result().StatusCode = 500

		expectedResponse2 := recorder2.Result()
		defer expectedResponse2.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
					body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null}`},
			},
		)

		err = welcomeManager.Run()
		require.Error(t, err)
		require.ErrorContains(t, err, "HTTP 500 Internal Server Error")

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.Equal(t, `{"level":"info","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.Equal(t, `{"level":"info","message":"Adding messages to channel"}`, parts[2])
		require.Equal(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.Equal(t, `{"level":"info","message_title":"my title 1","message":"Message missing - add Message"}`, parts[4])
		//nolint:lll
		require.Equal(t, `{"level":"info","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","message":"Sending Message to Channel"}`, parts[5])
		//nolint:lll
		require.Equal(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","message":"Could not send Message to Channel"}`, parts[6])
		require.Equal(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","message_title":"my title 1","message":"Could not add Message"}`, parts[7])
		require.Equal(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","message":"Could not add messages to channel"}`, parts[8])
		require.Equal(t, ``, parts[9])
	})

	//nolint:bodyclose
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
		recorder3.Result().Status = "500 Internal Server Error"
		recorder3.Result().StatusCode = 500

		expectedResponse3 := recorder3.Result()
		defer expectedResponse3.Body.Close()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-123/messages?limit=100"},
				{method: "POST", host: "discord.com", uri: "/api/v9/channels/channel-123/messages",
					body: `{"embeds":[{"type":"rich","title":"my title 1"}],"tts":false,"components":null}`},
				{method: "PUT", host: "discord.com", uri: "/api/v9/channels/channel-123/messages/123/reactions/my-emoji-1:emoji-123/@me"},
			},
		)

		err = welcomeManager.Run()
		require.Error(t, err)
		require.ErrorContains(t, err, "HTTP 500 Internal Server Error")

		parts := strings.Split(bufferLogs.String(), "\n")
		require.Equal(t, `{"level":"info","message":"Adding Handler on Message Reaction Add"}`, parts[0])
		require.Equal(t, `{"level":"info","message":"Adding Handler on Message Reaction Remove"}`, parts[1])
		require.Equal(t, `{"level":"info","message":"Adding messages to channel"}`, parts[2])
		require.Equal(t, `{"level":"info","channel_id":"channel-123","channel":"my-channel","message":"Getting Messages from Channel"}`, parts[3])
		require.Equal(t, `{"level":"info","message_title":"my title 1","message":"Message missing - add Message"}`, parts[4])
		//nolint:lll
		require.Equal(t, `{"level":"info","message_title":"my title 1","channel_id":"channel-123","channel":"my-channel","message":"Sending Message to Channel"}`, parts[5])
		require.Equal(t, `{"level":"info","message_id":"123","channel_id":"channel-123","channel":"my-channel","message":"Message Sent"}`, parts[6])
		//nolint:lll
		require.Equal(t, `{"level":"info","message_id":"123","message_title":"my title 1","emoji":"my-emoji-1:emoji-123","message":"Adding Reaction to Message"}`, parts[7])
		//nolint:lll
		require.Equal(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","message_id":"123","emoji":"my-emoji-1:emoji-123","message":"Could not add Reaction to Message"}`, parts[8])
		require.Equal(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","message_title":"my title 1","message":"Could not add Message"}`, parts[9])
		require.Equal(t, `{"level":"error","error":"HTTP 500 Internal Server Error, ","message":"Could not add messages to channel"}`, parts[10])
		require.Equal(t, ``, parts[11])
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
