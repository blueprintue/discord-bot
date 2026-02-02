//nolint:paralleltest
package healthchecks_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blueprintue/discord-bot/healthchecks"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFail(t *testing.T) {
	var bufferLogs bytes.Buffer

	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	currentRequestIdx := 0

	svr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if currentRequestIdx == 0 {
			assert.Equal(t, "/00000000-0000-0000-0000-000000000000/start", req.RequestURI)
			startedMessage, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, "starts", string(startedMessage))
		} else {
			assert.Equal(t, "/00000000-0000-0000-0000-000000000000/fail", req.RequestURI)
			failedMessage, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, "stops", string(failedMessage))
		}

		currentRequestIdx++

		res.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	healthchecksManager := healthchecks.NewHealthchecksManager(healthchecks.Configuration{
		BaseURL:        svr.URL,
		UUID:           "00000000-0000-0000-0000-000000000000",
		StartedMessage: "starts",
		FailedMessage:  "stops",
	})
	require.NotNil(t, healthchecksManager)

	err := healthchecksManager.Run()
	require.NoError(t, err)

	bufferLogs.Reset()

	healthchecksManager.Fail()

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","package":"healthchecks","message":"discord_bot.healthchecks.send_failed_message"}`, parts[0])
	require.Empty(t, parts[1])
}

func TestFail_Errors(t *testing.T) {
	var bufferLogs bytes.Buffer

	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	svr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	healthchecksManager := healthchecks.NewHealthchecksManager(healthchecks.Configuration{
		BaseURL:        svr.URL,
		UUID:           "00000000-0000-0000-0000-000000000000",
		StartedMessage: "starts",
		FailedMessage:  "stops",
	})
	require.NotNil(t, healthchecksManager)

	err := healthchecksManager.Run()
	require.Error(t, err)

	bufferLogs.Reset()

	healthchecksManager.Fail()

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"error","package":"healthchecks","error":"HTTP error 500","message":"discord_bot.healthchecks.send_failed_message_failed"}`, parts[0])
	require.Empty(t, parts[1])
}
