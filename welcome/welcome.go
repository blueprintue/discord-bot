package welcome

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/rancoud/blueprintue-discord/helpers"
	"log"
)

const emojiPersonInLotusPosition = "\U0001F9D8"

type Manager struct {
	channelID string
	session *discordgo.Session
}

func NewWelcomeManager(
	channelID string,
	session *discordgo.Session,
) *Manager {
	return &Manager{
		channelID: channelID,
		session: session,
	}
}

func (w *Manager) Run() {
	fmt.Println("Welcome -> Run")

	w.ReplaceMessage()
	w.session.AddHandler(w.onMessageReactionAdd)
}

func (w *Manager) ReplaceMessage() {
	messages, err := w.session.ChannelMessages(w.channelID, 100, "", "", "")
	if err != nil {
		log.Fatalf("could not read channel messages from channelID %s: %s", w.channelID, err)
	}

	// check if bot has already sent messages
	var messagesFromBot []*discordgo.Message
	for _, message := range messages {
		if message.Author.ID != w.session.State.User.ID {
			continue
		}

		messagesFromBot = append(messagesFromBot, message)
	}

	if len(messagesFromBot) == 0 {
		message, err := w.session.ChannelMessageSendEmbed(w.channelID, &discordgo.MessageEmbed{
			Title: "Human Verification",
			Description: "Click the :person_in_lotus_position: below to show that you're human!",
		})
		if err != nil {
			log.Fatalf("could not send embed message to channelID %s: %s", w.channelID, err)
		}

		err = w.session.MessageReactionAdd(w.channelID, message.ID, emojiPersonInLotusPosition)
		if err != nil {
			log.Fatalf("could not add reaction to welcome message to channelID %s - messageID %s: %s", w.channelID, message.ID, err)
		}
	} else {
		spew.Dump(messagesFromBot)
		if messagesFromBot[0].Embeds[0].Title == "Human Verification" {
			/*for _, reaction := range messagesFromBot[0].Reactions {
				//spew.Dump(reaction)
			}*/

			r, _ := helpers.MessageReactionsAll(w.session, w.channelID, messagesFromBot[0].ID, emojiPersonInLotusPosition)
			for _, reaction := range r {
				spew.Dump(reaction.Username)
			}
		}
	}
}

func (w *Manager) onMessageReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	if reaction.ChannelID != w.channelID {
		return
	}

	spew.Dump(reaction)
	if reaction.Emoji.Name == emojiPersonInLotusPosition {
		fmt.Println("onMessageReactionAdd")
	}
}

func (w *Manager) onMessageReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	if reaction.ChannelID != w.channelID {
		return
	}

	spew.Dump(reaction)
	if reaction.Emoji.Name == emojiPersonInLotusPosition {
		fmt.Println("onMessageReactionRemove")
	}
}