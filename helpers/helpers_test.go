package helpers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blueprintue/discord-bot/helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

func TestMessageReactionsAll(t *testing.T) {
	t.Parallel()

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
	data1, err := json.Marshal(expectedUsers[:2])
	require.NoError(t, err)

	recorder1 := httptest.NewRecorder()
	recorder1.Header().Add("Content-Type", "application/json")
	_, err = recorder1.Write(data1)
	require.NoError(t, err)

	expectedResponse1 := recorder1.Result()
	defer expectedResponse1.Body.Close()

	// user 333
	data2, err := json.Marshal(expectedUsers[2:])
	require.NoError(t, err)

	recorder2 := httptest.NewRecorder()
	recorder2.Header().Add("Content-Type", "application/json")
	_, err = recorder2.Write(data2)
	require.NoError(t, err)

	expectedResponse2 := recorder2.Result()
	defer expectedResponse2.Body.Close()

	// no user
	recorder3 := httptest.NewRecorder()
	recorder3.Header().Add("Content-Type", "application/json")
	_, err = recorder3.WriteString("[]")
	require.NoError(t, err)

	expectedResponse3 := recorder3.Result()
	defer expectedResponse3.Body.Close()

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

//nolint:funlen,tparallel
func TestMessageReactionsAll_Errors(t *testing.T) {
	t.Parallel()

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

	//nolint:paralleltest
	t.Run("should return users and error because json unmarshal failed", func(t *testing.T) {
		// user 111 + user 222
		data1, err := json.Marshal(users[:2])
		require.NoError(t, err)

		recorder1 := httptest.NewRecorder()
		recorder1.Header().Add("Content-Type", "application/json")
		_, err = recorder1.Write(data1)
		require.NoError(t, err)

		expectedResponse1 := recorder1.Result()
		defer expectedResponse1.Body.Close()

		// invalid json
		recorder2 := httptest.NewRecorder()
		recorder2.Header().Add("Content-Type", "application/json")
		_, err = recorder2.WriteString("-")
		require.NoError(t, err)

		expectedResponse2 := recorder2.Result()
		defer expectedResponse2.Body.Close()

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

	//nolint:paralleltest,bodyclose
	t.Run("should return users and error because request failed", func(t *testing.T) {
		// user 111 + user 222
		data1, err := json.Marshal(users[:2])
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
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?limit=100"},
				{method: "GET", host: "discord.com", uri: "/api/v9/channels/channel-id/messages/message-id/reactions/:emoji:?after=222&limit=100"},
			},
		)

		actualUsers, err := helpers.MessageReactionsAll(session, channelID, messageID, emojiID)
		require.ErrorContains(t, err, "HTTP 500 Internal Server Error")
		require.Equal(t, users[:2], actualUsers)
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
}

func (r requestTest) assert(t *testing.T, req *http.Request) {
	t.Helper()

	require.Equal(t, r.method, req.Method)
	require.Equal(t, r.host, req.Host)
	require.Equal(t, r.uri, req.URL.RequestURI())
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
