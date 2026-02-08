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
	require.JSONEq(t, `{"level":"info","message":"discord_bot.healthchecks.validating_configuration"}`, parts[0])
	require.JSONEq(t, `{"level":"info","base_url":"https://example.com/","message":"discord_bot.healthchecks.set_base_url"}`, parts[1])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.healthchecks.set_uuid"}`, parts[2])
	require.JSONEq(t, `{"level":"info","started_message":"starts","message":"discord_bot.healthchecks.set_started_message"}`, parts[3])
	require.JSONEq(t, `{"level":"info","failed_message":"stops","message":"discord_bot.healthchecks.set_failed_message"}`, parts[4])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.healthchecks.configuration_validated"}`, parts[5])
	require.Empty(t, parts[6])

	bufferLogs.Reset()

	healthchecksManager = healthchecks.NewHealthchecksManager(healthchecks.Configuration{
		UUID: "00000000-0000-0000-0000-000000000000",
	})
	require.NotNil(t, healthchecksManager)

	parts = strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","message":"discord_bot.healthchecks.validating_configuration"}`, parts[0])
	require.JSONEq(t, `{"level":"info","help":"BaseURL is empty, use default URL https://hc-ping.com/","message":"discord_bot.healthchecks.use_default_base_url"}`, parts[1])
	require.JSONEq(t, `{"level":"info","base_url":"https://hc-ping.com/","message":"discord_bot.healthchecks.set_base_url"}`, parts[2])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.healthchecks.set_uuid"}`, parts[3])
	require.JSONEq(t, `{"level":"info","help":"StartedMessage is empty, use default \"discord-bot started\"","message":"discord_bot.healthchecks.set_default_started_message"}`, parts[4])
	require.JSONEq(t, `{"level":"info","help":"FailedMessage is empty, use default \"discord-bot stopped\"","message":"discord_bot.healthchecks.set_default_failed_message"}`, parts[5])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.healthchecks.configuration_validated"}`, parts[6])
	require.Empty(t, parts[7])
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
					`{"level":"info","message":"discord_bot.healthchecks.validating_configuration"}`,
					`{"level":"info","help":"BaseURL is empty, use default URL https://hc-ping.com/","message":"discord_bot.healthchecks.use_default_base_url"}`,
					`{"level":"info","base_url":"https://hc-ping.com/","message":"discord_bot.healthchecks.set_base_url"}`,
					`{"level":"error","message":"discord_bot.healthchecks.empty_uuid"}`,
					`{"level":"error","message":"discord_bot.healthchecks.configuration_validation_failed"}`,
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
					`{"level":"info","message":"discord_bot.healthchecks.validating_configuration"}`,
					`{"level":"error","error":"parse \":::::..:::::\": missing protocol scheme","base_url":":::::..:::::","message":"discord_bot.healthchecks.base_url_parsing_failed"}`,
					`{"level":"error","message":"discord_bot.healthchecks.configuration_validation_failed"}`,
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

			require.Len(tt, parts, len(testCase.want.logs))
		})
	}
}
