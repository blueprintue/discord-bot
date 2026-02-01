// Package main is the entry point.
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/blueprintue/discord-bot/configuration"
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

//nolint:funlen,cyclop
func main() {
	var err error

	log.Info().
		Str("module", "main").
		Str("version", version).
		Msg("discord_bot.starting")

	log.Info().
		Str("module", "main").
		Str("configuration_file", configurationFilename).
		Msg("discord_bot.reading_configuration")

	config, err := configuration.ReadConfiguration(os.DirFS("."), configurationFilename)
	if err != nil {
		log.Fatal().
			Str("module", "main").
			Str("configuration_file", configurationFilename).
			Err(err).
			Msg("discord_bot.configuration_read_failed")
	}

	log.Info().
		Str("module", "main").
		Str("configuration_file", configurationFilename).
		Msg("discord_bot.configuration_read")

	log.Info().
		Str("module", "main").
		Msg("discord_bot.configuring_logger")

	err = logger.Configure(config.Log)
	if err != nil {
		log.Fatal().
			Str("module", "main").
			Err(err).
			Msg("discord_bot.logger_configured_failed")
	}

	log.Info().
		Str("module", "main").
		Msg("discord_bot.logger_configured")

	log.Info().
		Str("module", "main").
		Msg("discord_bot.creating_discord_session")

	discordSession, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatal().
			Str("module", "main").
			Err(err).
			Msg("discord_bot.discord_session_creation_failed")
	}

	discordSession.Identify.Intents = discordgo.IntentsAll

	log.Info().
		Str("module", "main").
		Msg("discord_bot.discord_session_created")

	log.Info().
		Str("module", "main").
		Msg("discord_bot.opening_discord_session")

	err = discordSession.Open()
	if err != nil {
		log.Fatal().
			Str("module", "main").
			Err(err).
			Msg("discord_bot.discord_session_open_failed")
	}

	timeoutChan := time.After(timeoutStateFilled)

pending_discord_session_open_completely:
	for {
		select {
		case <-timeoutChan:
			log.Fatal().
				Str("module", "main").
				Err(err).
				Msg("discord_bot.discord_session_open_completely_failed")
		default:
			log.Info().
				Str("module", "main").
				Msg("discord_bot.pending_discord_session_open_completely")
	
			time.Sleep(waitStateFilled)
	
			if hasRequiredStateFieldsFilled(discordSession) {
				break pending_discord_session_open_completely
			}
		}
	}

	log.Info().
		Str("module", "main").
		Msg("discord_bot.discord_session_opened")

	log.Info().
		Str("module", "main").
		Msg("discord_bot.welcome.creating")

	welcomeManager := welcome.NewWelcomeManager(discordSession, config.Discord.Name, config.Modules.WelcomeConfiguration)
	if welcomeManager == nil {
		log.Error().
			Str("module", "main").
			Msg("discord_bot.welcome.creation_failed")
	} else {
		log.Info().
			Str("module", "main").
			Msg("discord_bot.welcome.created")

		log.Info().
			Str("module", "main").
			Msg("discord_bot.welcome.starting")

		err = welcomeManager.Run()
		if err != nil {
			log.Error().
				Str("module", "main").
				Err(err).
				Msg("discord_bot.welcome.start_failed")

			closeSessionDiscord(discordSession)

			return
		}

		log.Info().
			Str("module", "main").
			Msg("discord_bot.welcome.started")
	}

	log.Info().
		Str("module", "main").
		Msg("discord_bot.healthchecks.creating")

	healthchecksManager := healthchecks.NewHealthchecksManager(config.Modules.HealthcheckConfiguration)
	if healthchecksManager == nil {
		log.Error().
			Str("module", "main").
			Msg("discord_bot.healthchecks.creation_failed")
	} else {
		log.Info().
			Str("module", "main").
			Msg("discord_bot.healthchecks.created")

		log.Info().
			Str("module", "main").
			Msg("discord_bot.healthchecks.starting")

		err = healthchecksManager.Run()
		if err != nil {
			log.Error().
				Str("module", "main").
				Err(err).
				Msg("discord_bot.healthchecks.start_failed")
		} else {
			log.Info().
				Str("module", "main").
				Msg("discord_bot.healthchecks.started")
		}
	}

	log.Info().
		Str("module", "main").
		Str("help", "Press CTRL+C to stop").
		Msg("discord_bot.started")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-sc

	closeSessionDiscord(discordSession)

	if healthchecksManager != nil {
		healthchecksManager.Fail()
	}

	log.Warn().
		Str("module", "main").
		Any("signal", sig).
		Msgf("discord_bot.closed")

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
		Str("module", "main").
		Msg("discord_bot.closing")

	err := discordSession.Close()
	if err != nil {
		log.Fatal().
			Str("module", "main").
			Err(err).
			Msg("discord_bot.discord_session_close_failed")
	}
}
