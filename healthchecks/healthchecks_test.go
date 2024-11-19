//nolint:paralleltest
package healthchecks_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/blueprintue/discord-bot/healthchecks"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func TestNewHealthchecksManager(t *testing.T) {
	var bufferLogs bytes.Buffer
	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	healthchecksManager := healthchecks.NewHealthchecksManager(healthchecks.Configuration{
		BaseURL:        "https://example.com",
		UUID:           "00000000-0000-0000-0000-000000000000",
		StartedMessage: "starts",
		FailedMessage:  "stops",
	})
	require.NotNil(t, healthchecksManager)

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","message":"Checking configuration for Healthchecks"}`, parts[0])
	require.Equal(t, ``, parts[1])

	bufferLogs.Reset()

	healthchecksManager = healthchecks.NewHealthchecksManager(healthchecks.Configuration{
		UUID: "00000000-0000-0000-0000-000000000000",
	})
	require.NotNil(t, healthchecksManager)

	parts = strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","message":"Checking configuration for Healthchecks"}`, parts[0])
	require.JSONEq(t, `{"level":"info","message":"BaseURL is empty, use default URL https://hc-ping.com/"}`, parts[1])
	require.JSONEq(t, `{"level":"info","message":"StartedMessage is empty, use default \"discord-bot started\""}`, parts[2])
	require.JSONEq(t, `{"level":"info","message":"FailedMessage is empty, use default \"discord-bot stopped\""}`, parts[3])
	require.Equal(t, ``, parts[4])
}

//nolint:funlen,tparallel
func TestNewHealthchecksManager_ErrorHasValidConfigurationInFile(t *testing.T) {
	t.Parallel()

	type args struct {
		config healthchecks.Configuration
	}

	type want struct {
		logs []string
	}

	testCases := map[string]struct {
		args args
		want want
	}{
		"should return nil because uuid is empty": {
			args: args{
				config: healthchecks.Configuration{},
			},
			want: want{
				logs: []string{
					`{"level":"info","message":"Checking configuration for Healthchecks"}`,
					`{"level":"info","message":"BaseURL is empty, use default URL https://hc-ping.com/"}`,
					`{"level":"error","message":"UUID is empty"}`,
					``,
				},
			},
		},
		"should return nil because base_url is invalid": {
			args: args{
				config: healthchecks.Configuration{
					BaseURL: ":::::..:::::",
				},
			},
			want: want{
				logs: []string{
					`{"level":"info","message":"Checking configuration for Healthchecks"}`,
					`{"level":"error","error":"parse \":::::..:::::\": missing protocol scheme","base_url":":::::..:::::","message":"BaseURL is invalid"}`,
					``,
				},
			},
		},
	}

	for testCaseName, testCase := range testCases {
		t.Run(testCaseName, func(tt *testing.T) {
			var bufferLogs bytes.Buffer
			log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

			bufferLogs.Reset()

			healthchecksManager := healthchecks.NewHealthchecksManager(testCase.args.config)
			require.Nil(tt, healthchecksManager)

			parts := strings.Split(bufferLogs.String(), "\n")
			for idx := range testCase.want.logs {
				require.Equal(tt, testCase.want.logs[idx], parts[idx])
			}

			require.Equal(tt, len(testCase.want.logs), len(parts))
		})
	}
}
