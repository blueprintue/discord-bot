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
	"github.com/stretchr/testify/require"
)

func TestFail(t *testing.T) {
	var bufferLogs bytes.Buffer
	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	currentRequestIdx := 0

	svr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if currentRequestIdx == 0 {
			require.Equal(t, "/00000000-0000-0000-0000-000000000000/start", req.RequestURI)
			startedMessage, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			require.Equal(t, "starts", string(startedMessage))
		} else {
			require.Equal(t, "/00000000-0000-0000-0000-000000000000/fail", req.RequestURI)
			failedMessage, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			require.Equal(t, "stops", string(failedMessage))
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
	require.Equal(t, ``, parts[0])
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
	require.Equal(t, `{"level":"error","error":"HTTP error 500","message":"Could not send Fail HealthChecks client"}`, parts[0])
	require.Equal(t, ``, parts[1])
}
