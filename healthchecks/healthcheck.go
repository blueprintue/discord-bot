package healthchecks

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/crazy-max/gohealthchecks"
	"github.com/rs/zerolog/log"
)

type Configuration struct {
	BaseURL        string `json:"base_url"`
	UUID           string `json:"uuid"`
	StartedMessage string `json:"started_message"`
	FailedMessage  string `json:"failed_message"`
}

type Manager struct {
	client         *gohealthchecks.Client
	baseURL        *url.URL
	uuid           string
	startedMessage string
	failedMessage  string
}

func NewHealthchecksManager(
	config Configuration,
) *Manager {
	manager := &Manager{}

	log.Info().Msg("Checking configuration for Healthchecks")

	if !manager.hasValidConfigurationInFile(config) {
		return nil
	}

	return manager
}

func (m *Manager) hasValidConfigurationInFile(config Configuration) bool {
	baseRawURL := config.BaseURL
	if baseRawURL == "" {
		log.Info().
			Msg("BaseURL is empty, use default URL https://hc-ping.com/")

		baseRawURL = "https://hc-ping.com/"
	}

	baseURL, err := url.Parse(baseRawURL)
	if err != nil {
		log.Error().
			Err(err).
			Str("base_url", baseRawURL).
			Msg("BaseURL is invalid")

		return false
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	m.baseURL = baseURL

	if config.UUID == "" {
		log.Error().
			Msg("UUID is empty")

		return false
	}

	m.uuid = config.UUID

	m.startedMessage = config.StartedMessage
	if m.startedMessage == "" {
		log.Info().
			Msg(`StartedMessage is empty, use default "discord-bot started"`)

		m.startedMessage = "discord-bot started"
	}

	m.failedMessage = config.FailedMessage
	if m.failedMessage == "" {
		log.Info().
			Msg(`FailedMessage is empty, use default "discord-bot stopped"`)

		m.failedMessage = "discord-bot stopped"
	}

	return true
}

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
		log.Error().
			Err(err).
			Msg("Could not send Start HealthChecks client")

		return fmt.Errorf("%w", err)
	}

	return nil
}

func (m *Manager) Fail() {
	err := m.client.Fail(
		context.Background(),
		gohealthchecks.PingingOptions{
			UUID: m.uuid,
			Logs: m.failedMessage,
		},
	)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Could not send Fail HealthChecks client")
	}
}
