package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/rancoud/blueprintue-discord/configuration"
	"github.com/rancoud/blueprintue-discord/welcome"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var waitStateFilled = 10 * time.Millisecond

func main() {
	config, err := configuration.ReadConfiguration("config.json")
	if err != nil {
		log.Fatalf("could not read configuration: %s", err)
	}

	session, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatalf("could not create session: %s", err)
	}

	err = session.Open()
	if err != nil {
		log.Fatalf("could not open session: %s", err)
	}

	for {
		time.Sleep(waitStateFilled)
		if hasRequiredStateFieldsFilled(session) {
			break
		}
	}

	spew.Dump(session.State)

	err = session.Close()

	//session.AddHandler(handlers.Debug)

	welcomeManager := welcome.NewWelcomeManager(session, config.Modules.WelcomeConfiguration)
	welcomeManager.Run()

	fmt.Println("Bot is now running. Press CTRL+C to stop")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err = session.Close()
	if err != nil {
		log.Fatalf("could not close session %s", err)
	}
}

func hasRequiredStateFieldsFilled(session *discordgo.Session) bool {
	return session.State.Guilds[0].Channels != nil && session.State.User != nil
}