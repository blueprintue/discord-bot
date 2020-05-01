package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-colorable"
	"github.com/rancoud/blueprintue-discord/configuration"
	"github.com/rancoud/blueprintue-discord/welcome"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const waitStateFilled = 10 * time.Millisecond
const configurationFilename = "config.json"

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetOutput(colorable.NewColorableStdout())
}

func main() {
	var err error

	logrus.Infof("[MAIN]\tRead configuration from file: %s", configurationFilename)
	config, err := configuration.ReadConfiguration(configurationFilename)
	if err != nil {
		logrus.Fatalf("[MAIN]\tCould not read configuration: %s", err)
	}

	logrus.Info("[MAIN]\tCreate discord session")
	session, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		logrus.Fatalf("[MAIN]\tCould not create discord session: %s", err)
	}

	logrus.Info("[MAIN]\tOpen discord session")
	err = session.Open()
	if err != nil {
		logrus.Fatalf("[MAIN]\tCould not open discord session: %s", err)
	}

	for {
		logrus.Info("[MAIN]\tPending session to be ready...")
		time.Sleep(waitStateFilled)
		if hasRequiredStateFieldsFilled(session) {
			break
		}
	}

	logrus.Info("[MAIN]\t--- Modules loading ---")

	logrus.Info("[MAIN]\t[WELCOME]\tCreate Welcome Manager")
	welcomeManager := welcome.NewWelcomeManager(session, config.Discord.Name, config.Modules.WelcomeConfiguration)

	logrus.Info("[MAIN]\t[WELCOME]\tRun Welcome Manager")
	err = welcomeManager.Run()
	if err != nil {
		logrus.Errorf("[MAIN]\t[WELCOME]\tRun Welcome Manager: %s", err)
		closeSessionDiscord(session)
		return
	}

	logrus.Info("[MAIN]\t--- Modules loaded ---")

	logrus.Info("[MAIN]\tBot is now running. Press CTRL+C to stop")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	closeSessionDiscord(session)
}

func hasRequiredStateFieldsFilled(session *discordgo.Session) bool {
	return session.State.Guilds[0].Channels != nil && session.State.User != nil
}

func closeSessionDiscord(session *discordgo.Session) {
	logrus.Info("[MAIN]\tClose discord session")
	err := session.Close()
	if err != nil {
		logrus.Fatalf("[MAIN]\tCould not close discord session %s", err)
	}
}