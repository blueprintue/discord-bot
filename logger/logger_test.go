package logger_test

import (
	"os"
	"testing"

	"github.com/blueprintue/discord-bot/configuration"
	"github.com/blueprintue/discord-bot/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestConfigure(t *testing.T) {
	err := logger.Configure(configuration.Log{Filename: os.TempDir() + "/test.log", Level: "info"})
	require.NoError(t, err)

	require.Equal(t, "2006-01-02T15:04:05.000000000Z07:00", zerolog.TimeFieldFormat)
}

//nolint:paralleltest
func TestConfigure_Errors(t *testing.T) {
	err := logger.Configure(configuration.Log{Filename: os.TempDir(), Level: "info"})
	require.ErrorContains(t, err, "File /tmp is a directory")

	err = logger.Configure(configuration.Log{Filename: os.TempDir() + "/test.log", Level: "invalid"})
	require.ErrorContains(t, err, "Unknown Level String: 'invalid', defaulting to NoLevel")
}
