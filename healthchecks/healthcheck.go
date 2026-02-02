// Package healthchecks defines configuration struct and how to ping healthchecks for status.
package healthchecks

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/crazy-max/gohealthchecks"
	"github.com/rs/zerolog/log"
)

// Configuration contains healthchecks parameters.
type Configuration struct {
	BaseURL        string `json:"base_url"`
	UUID           string `json:"uuid"`
	StartedMessage string `json:"started_message"`
	FailedMessage  string `json:"failed_message"`
}

// Manager is a struct.
type Manager struct {
	client         *gohealthchecks.Client
	baseURL        *url.URL
	uuid           string
	startedMessage string
	failedMessage  string
}

// NewHealthchecksManager checks configuration and returns a manager.
func NewHealthchecksManager(
	config Configuration,
) *Manager {
	manager := &Manager{}

	log.Info().
		Str("package", "healthchecks").
		Msg("discord_bot.healthchecks.validating_configuration")

	if !manager.hasValidConfigurationInFile(config) {
		log.Error().
			Str("package", "healthchecks").
			Msg("discord_bot.healthchecks.configuration_validation_failed")

		return nil
	}

	log.Info().
		Str("package", "healthchecks").
		Msg("discord_bot.healthchecks.configuration_validated")

	return manager
}

//nolint:funlen
func (m *Manager) hasValidConfigurationInFile(config Configuration) bool {
	baseRawURL := config.BaseURL
	if baseRawURL == "" {
		log.Info().
			Str("package", "healthchecks").
			Str("help", "BaseURL is empty, use default URL https://hc-ping.com/").
			Msg("discord_bot.healthchecks.use_default_base_url")

		baseRawURL = "https://hc-ping.com/"
	}

	baseURL, err := url.Parse(baseRawURL)
	if err != nil {
		log.Error().Err(err).
			Str("package", "healthchecks").
			Str("base_url", baseRawURL).
			Msg("discord_bot.healthchecks.base_url_parsing_failed")

		return false
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	log.Info().
		Str("package", "healthchecks").
		Str("base_url", baseURL.String()).
		Msg("discord_bot.healthchecks.set_base_url")

	m.baseURL = baseURL

	if config.UUID == "" {
		log.Error().
			Str("package", "healthchecks").
			Msg("discord_bot.healthchecks.empty_uuid")

		return false
	}

	log.Info().
		Str("package", "healthchecks").
		Msg("discord_bot.healthchecks.set_uuid")

	m.uuid = config.UUID

	m.startedMessage = config.StartedMessage
	if m.startedMessage == "" {
		log.Info().
			Str("package", "healthchecks").
			Str("help", `StartedMessage is empty, use default "discord-bot started"`).
			Msg("discord_bot.healthchecks.set_default_started_message")

		m.startedMessage = "discord-bot started"
	} else {
		log.Info().
			Str("package", "healthchecks").
			Str("started_message", m.startedMessage).
			Msg("discord_bot.healthchecks.set_started_message")
	}

	m.failedMessage = config.FailedMessage
	if m.failedMessage == "" {
		log.Info().
			Str("package", "healthchecks").
			Str("help", `FailedMessage is empty, use default "discord-bot stopped"`).
			Msg("discord_bot.healthchecks.set_default_failed_message")

		m.failedMessage = "discord-bot stopped"
	} else {
		log.Info().
			Str("package", "healthchecks").
			Str("failed_message", m.failedMessage).
			Msg("discord_bot.healthchecks.set_failed_message")
	}

	return true
}

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
			Str("package", "healthchecks").
			Msg("discord_bot.healthchecks.send_started_message_failed")

		return fmt.Errorf("%w", err)
	}

	log.Info().
		Str("package", "healthchecks").
		Msg("discord_bot.healthchecks.send_started_message")

	return nil
}

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
			Str("package", "healthchecks").
			Msg("discord_bot.healthchecks.send_failed_message_failed")
		
		return
	}

	log.Info().
		Str("package", "healthchecks").
		Msg("discord_bot.healthchecks.send_failed_message")
}
