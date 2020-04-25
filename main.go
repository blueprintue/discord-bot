package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rancoud/blueprintue-discord/handlers"
	"github.com/rancoud/blueprintue-discord/welcome"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var token string
var waitStateFilled = 10 * time.Millisecond

func init() {
	flag.StringVar(&token, "t", "", "Bot token")
	flag.Parse()
}

func main() {
	session, err := discordgo.New("Bot " + token)
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

	session.AddHandler(handlers.Debug)

	welcomeManager := welcome.NewWelcomeManager("671032699879686171", session)
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