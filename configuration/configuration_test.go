package configuration_test

import (
	"testing"

	"github.com/blueprintue/discord-bot/configuration"
	"github.com/stretchr/testify/assert"
)

func TestReadConfiguration(t *testing.T) {
	t.Parallel()

	conf, err := configuration.ReadConfiguration("")
	assert.Error(t, err)
	assert.Nil(t, conf)
}
