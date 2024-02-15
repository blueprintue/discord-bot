package helpers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blueprintue/discord-bot/helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
}

func (r requestTest) assert(t *testing.T, req *http.Request) {
	assert.Equal(t, r.method, req.Method)
	assert.Equal(t, r.host, req.Host)
	assert.Equal(t, r.uri, req.URL.RequestURI())
}

func (rt *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.requestsTest[rt.idxResponse].assert(rt.test, req)

	resp := rt.responsesMocked[rt.idxResponse]

	rt.idxResponse++

	return resp, nil
}

func TestMessageReactionsAll(t *testing.T) {
	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	expectedUsers := []*discordgo.User{
		{ID: "111"},
		{ID: "222"},
		{ID: "333"},
	}

	channelID := "channel-id"
	messageID := "message-id"
	emojiID := ":emoji:"

	// user 111 + user 222
	data1, _ := json.Marshal(expectedUsers[:2])
	recorder1 := httptest.NewRecorder()
	recorder1.Header().Add("Content-Type", "application/json")
	recorder1.WriteString(string(data1))
	expectedResponse1 := recorder1.Result()

	// user 333
	data2, _ := json.Marshal(expectedUsers[2:])
	recorder2 := httptest.NewRecorder()
	recorder2.Header().Add("Content-Type", "application/json")
	recorder2.WriteString(string(data2))
	expectedResponse2 := recorder2.Result()

	// no user
	recorder3 := httptest.NewRecorder()
	recorder3.Header().Add("Content-Type", "application/json")
	recorder3.WriteString("[]")
	expectedResponse3 := recorder3.Result()

	session.Client = createClient(t,
		[]*http.Response{expectedResponse1, expectedResponse2, expectedResponse3},
		[]requestTest{
			{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?limit=100"},
			{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?after=222&limit=100"},
			{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?after=333&limit=100"},
		},
	)

	actualUsers, err := helpers.MessageReactionsAll(session, channelID, messageID, emojiID)
	require.NoError(t, err)
	require.Equal(t, expectedUsers, actualUsers)
}

func TestMessageReactionsAllErrors(t *testing.T) {
	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	users := []*discordgo.User{
		{ID: "111"},
		{ID: "222"},
		{ID: "333"},
	}

	channelID := "channel-id"
	messageID := "message-id"
	emojiID := ":emoji:"

	t.Run("should return users and error because json unmarshal failed", func(t *testing.T) {
		// user 111 + user 222
		data1, _ := json.Marshal(users[:2])
		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		recorder1.WriteString(string(data1))
		expectedResponse1 := recorder1.Result()

		// invalid json
		recorder2 := httptest.NewRecorder()
		recorder2.Header().Add("Content-Type", "application/json")
		recorder2.WriteString("-")
		expectedResponse2 := recorder2.Result()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?after=222&limit=100"},
			},
		)

		actualUsers, err := helpers.MessageReactionsAll(session, channelID, messageID, emojiID)
		require.ErrorIs(t, err, discordgo.ErrJSONUnmarshal)
		require.Equal(t, users[:2], actualUsers)
	})

	t.Run("should return users and error because request failed", func(t *testing.T) {
		// user 111 + user 222
		data1, _ := json.Marshal(users[:2])
		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		recorder1.WriteString(string(data1))
		expectedResponse1 := recorder1.Result()

		// request failed
		recorder2 := httptest.NewRecorder()
		recorder2.Result().Status = "500 Internal Server Error"
		recorder2.Result().StatusCode = 500
		expectedResponse2 := recorder2.Result()

		session.Client = createClient(t,
			[]*http.Response{expectedResponse1, expectedResponse2},
			[]requestTest{
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?after=222&limit=100"},
			},
		)

		actualUsers, err := helpers.MessageReactionsAll(session, channelID, messageID, emojiID)
		require.ErrorContains(t, err, "HTTP 500 Internal Server Error")
		require.Equal(t, users[:2], actualUsers)
	})
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
