package healthchecks

import (
	"context"

	"github.com/crazy-max/gohealthchecks"
	"github.com/rs/zerolog/log"
)

// Fail send a ping status with a message.
func (m *Manager) Fail() {
	err := m.client.Fail(
		context.Background(),
		gohealthchecks.PingingOptions{
			UUID: m.uuid,
			Logs: m.failedMessage,
		},
	)
	if err != nil {
		log.Error().Err(err).
			Msg("discord_bot.healthchecks.send_failed_message_failed")

		return
	}

	log.Info().
		Msg("discord_bot.healthchecks.send_failed_message")
}
