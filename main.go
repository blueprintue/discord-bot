package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/rancoud/blueprintue-discord/logger"

	"github.com/bwmarrin/discordgo"
	"github.com/rancoud/blueprintue-discord/configuration"
	"github.com/rancoud/blueprintue-discord/welcome"
	"github.com/rs/zerolog/log"
)

const waitStateFilled = 10 * time.Millisecond
const configurationFilename = "config.json"

func main() {
	var err error

	log.Info().Msgf("Read configuration from file: %s", configurationFilename)
	config, err := configuration.ReadConfiguration(configurationFilename)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not read configuration")
	}

	logger.Configure(config)

	log.Info().Msg("Create discord session")
	session, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create discord session")
	}

	log.Info().Msg("Open discord session")
	err = session.Open()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not open discord session")
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

	log.Info().Msg("Run Welcome Manager")
	err = welcomeManager.Run()
	if err != nil {
		log.Error().Err(err).Msg("could not run Welcome Manager")
		closeSessionDiscord(session)
		return
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
	log.Info().Msg("Close discord session")
	err := session.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not close discord session")
	}
}
