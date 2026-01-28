// Package main is the entry point.
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/blueprintue/discord-bot/configuration"
	"github.com/blueprintue/discord-bot/exporter"
	"github.com/blueprintue/discord-bot/healthchecks"
	"github.com/blueprintue/discord-bot/logger"
	"github.com/blueprintue/discord-bot/welcome"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const (
	waitStateFilled       = 250 * time.Millisecond
	timeoutStateFilled    = 10 * time.Second
	configurationFilename = "config.json"
)

var version = "edge"

//nolint:funlen
func main() {
	var err error

	log.Info().
		Str("version", version).
		Msg("discord_bot.main.starting")

	log.Info().
		Str("configuration_file", configurationFilename).
		Msg("discord_bot.main.reading_configuration")

	config, err := configuration.ReadConfiguration(os.DirFS("."), configurationFilename)
	if err != nil {
		log.Fatal().Err(err).
			Str("configuration_file", configurationFilename).
			Msg("discord_bot.main.configuration_read_failed")
	}

	log.Info().
		Str("configuration_file", configurationFilename).
		Msg("discord_bot.main.configuration_read")

	log.Info().
		Msg("discord_bot.main.configuring_logger")

	err = logger.Configure(config.Log)
	if err != nil {
		log.Fatal().Err(err).
			Msg("discord_bot.main.logger_configured_failed")
	}

	log.Info().
		Msg("discord_bot.main.logger_configured")

	log.Info().
		Msg("discord_bot.main.creating_discord_session")

	discordSession, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatal().Err(err).
			Msg("discord_bot.main.discord_session_creation_failed")
	}

	discordSession.Identify.Intents = discordgo.IntentsAll

	log.Info().
		Msg("discord_bot.main.discord_session_created")

	log.Info().
		Msg("discord_bot.main.opening_discord_session")

	err = discordSession.Open()
	if err != nil {
		log.Fatal().Err(err).
			Msg("discord_bot.main.discord_session_open_failed")
	}

	timeoutChan := time.After(timeoutStateFilled)

pending_discord_session_open_completely:
	for {
		select {
		case <-timeoutChan:
			log.Fatal().Err(err).
				Msg("discord_bot.main.discord_session_open_completely_failed")
		default:
			log.Info().
				Msg("discord_bot.main.pending_discord_session_open_completely")

			time.Sleep(waitStateFilled)

			if hasRequiredStateFieldsFilled(discordSession) {
				break pending_discord_session_open_completely
			}
		}
	}

	startModuleExporter(config.Modules.ExporterConfiguration, config.Discord.Name, discordSession)

	log.Info().
		Msg("discord_bot.main.discord_session_opened")

	healthchecksManager := startModuleHealthchecks(config.Modules.HealthcheckConfiguration)

	startModuleWelcome(config.Modules.WelcomeConfiguration, config.Discord.Name, discordSession)

	log.Info().
		Str("help", "Press CTRL+C to stop").
		Msg("discord_bot.main.started")

	sighupChan := make(chan os.Signal, 1)
	signal.Notify(sighupChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-sighupChan

	closeSessionDiscord(discordSession)

	if healthchecksManager != nil {
		healthchecksManager.Fail()
	}

	log.Warn().
		Any("signal", sig).
		Msg("discord_bot.main.closed")

	os.Exit(0)
}

func hasRequiredStateFieldsFilled(discordSession *discordgo.Session) bool {
	return discordSession != nil &&
		discordSession.State != nil &&
		discordSession.State.User != nil &&
		len(discordSession.State.Guilds) > 0 &&
		discordSession.State.Guilds[0] != nil &&
		len(discordSession.State.Guilds[0].Channels) > 0 &&
		discordSession.State.Guilds[0].Channels[0] != nil
}

func closeSessionDiscord(discordSession *discordgo.Session) {
	log.Info().
		Msg("discord_bot.main.closing")

	err := discordSession.Close()
	if err != nil {
		log.Fatal().Err(err).
			Msg("discord_bot.main.discord_session_close_failed")
	}
}

func startModuleExporter(configuration *exporter.Configuration, guildName string, discordSession *discordgo.Session) {
	if configuration == nil {
		log.Info().
			Msg("discord_bot.main.exporter.skipped")

		return
	}

	log.Info().
		Msg("discord_bot.main.exporter.creating")

	exporterManager := exporter.NewExporterManager(*configuration, guildName, discordSession)
	if exporterManager == nil {
		log.Error().
			Msg("discord_bot.main.exporter.creating_failed")

		return
	}

	log.Info().
		Msg("discord_bot.main.exporter.created")

	exporterManager.Run()
}

func startModuleHealthchecks(configuration *healthchecks.Configuration) *healthchecks.Manager {
	if configuration == nil {
		log.Info().
			Msg("discord_bot.main.healthchecks.skipped")

		return nil
	}

	log.Info().
		Msg("discord_bot.main.healthchecks.creating")

	healthchecksManager := healthchecks.NewHealthchecksManager(*configuration)
	if healthchecksManager == nil {
		log.Error().
			Msg("discord_bot.main.healthchecks.creation_failed")

		return nil
	}

	log.Info().
		Msg("discord_bot.main.healthchecks.created")

	log.Info().
		Msg("discord_bot.main.healthchecks.starting")

	err := healthchecksManager.Run()
	if err != nil {
		log.Error().Err(err).
			Msg("discord_bot.main.healthchecks.start_failed")

		return nil
	}

	log.Info().
		Msg("discord_bot.main.healthchecks.started")

	return healthchecksManager
}

func startModuleWelcome(configuration *welcome.Configuration, guildName string, discordSession *discordgo.Session) {
	if configuration == nil {
		log.Info().
			Msg("discord_bot.main.welcome.skipped")

		return
	}

	log.Info().
		Msg("discord_bot.main.welcome.creating")

	welcomeManager := welcome.NewWelcomeManager(*configuration, guildName, discordSession)
	if welcomeManager == nil {
		log.Error().
			Msg("discord_bot.main.welcome.creation_failed")

		return
	}

	log.Info().
		Msg("discord_bot.main.welcome.created")

	log.Info().
		Msg("discord_bot.main.welcome.starting")

	err := welcomeManager.Run()
	if err != nil {
		log.Error().Err(err).
			Msg("discord_bot.main.welcome.start_failed")

		return
	}

	log.Info().
		Msg("discord_bot.main.welcome.started")
}
