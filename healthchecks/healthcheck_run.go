package healthchecks

import (
	"context"
	"fmt"

	"github.com/crazy-max/gohealthchecks"
	"github.com/rs/zerolog/log"
)

// Run creates a client and start the monitoring by sending a ping status with a message.
func (m *Manager) Run() error {
	m.client = gohealthchecks.NewClient(
		&gohealthchecks.ClientOptions{
			BaseURL: m.baseURL,
		},
	)

	err := m.client.Start(
		context.Background(),
		gohealthchecks.PingingOptions{
			UUID: m.uuid,
			Logs: m.startedMessage,
		},
	)
	if err != nil {
		log.Error().Err(err).
			Msg("discord_bot.healthchecks.send_started_message_failed")

		return fmt.Errorf("%w", err)
	}

	log.Info().
		Msg("discord_bot.healthchecks.send_started_message")

	return nil
}
