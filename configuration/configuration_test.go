package configuration_test

import (
	"testing"

	"github.com/blueprintue/discord-bot/configuration"
	"github.com/stretchr/testify/require"
)

func TestReadConfiguration(t *testing.T) {
	t.Parallel()

	conf, err := configuration.ReadConfiguration("")
	require.Error(t, err)
	require.Nil(t, conf)
}
