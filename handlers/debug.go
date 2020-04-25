package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

func Debug(s *discordgo.Session, msg *discordgo.MessageCreate) {
	spew.Dump(msg)
}
