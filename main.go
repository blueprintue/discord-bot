package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/blueprintue/discord-bot/configuration"
	"github.com/blueprintue/discord-bot/logger"
	"github.com/blueprintue/discord-bot/welcome"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const (
	waitStateFilled       = 250 * time.Millisecond
	configurationFilename = "config.json"
)

var version = "edge"

//nolint:funlen
func main() {
	var err error

	log.Info().Str("version", version).Msg("Starting discord-bot")

	log.Info().Msgf("Read configuration from file: %s", configurationFilename)

	config, err := configuration.ReadConfiguration(configurationFilename)
	if err != nil {
		log.Fatal().Err(err).Msg("Error on configuration")
	}

	logger.Configure(config)

	log.Info().Msg("Create discordgo session")

	session, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create discordgo session")
	}

	log.Info().Msg("Open discordgo session")

	err = session.Open()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not open discordgo session")
	}

	for {
		log.Info().Msg("Pending session to be ready...")
		time.Sleep(waitStateFilled)

		if hasRequiredStateFieldsFilled(session) {
			break
		}
	}

	log.Info().Msg("Create Welcome Manager")

	welcomeManager := welcome.NewWelcomeManager(session, config.Discord.Name, config.Modules.WelcomeConfiguration)
	if welcomeManager == nil {
		log.Error().Msg("Could not start Welcome Manager")
	} else {
		log.Info().Msg("Run Welcome Manager")

		err = welcomeManager.Run()
		if err != nil {
			log.Error().Err(err).Msg("Could not run Welcome Manager")

			closeSessionDiscord(session)

			return
		}
	}

	log.Info().Msg("Bot is now running. Press CTRL+C to stop")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	closeSessionDiscord(session)
}

func hasRequiredStateFieldsFilled(session *discordgo.Session) bool {
	return session.State.Guilds[0].Channels != nil && session.State.User != nil
}

func closeSessionDiscord(session *discordgo.Session) {
	log.Info().Msg("Close discordgo session")

	err := session.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not close discordgo session")
	}
}
