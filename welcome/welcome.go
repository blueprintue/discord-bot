package welcome

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	//"github.com/davecgh/go-spew/spew"
	//"github.com/rancoud/blueprintue-discord/helpers"
	//"log"
)
// todo: emoji twitch
// todo: emoji blueprintue
// todo: 2 messages Click the :twitch: below to join the twitch part -> donne le role twitch viewer
// todo: 2 messages Click the :blueprinUE: below to join the blueprintue part -> donne le role blueprintUE anonymous
// todo: restrict user with only verified email and mfa -> if not remove reaction and sent private message

/*
reflexion sur les roles
twitch viewer -> permet de lire les messages
blueprintue anonymous -> permet de lire les messages

blueprintue member -> ecrire

twitch sub -> écrire message

twitch follower -> écrire message
twitch sub -> écrire message

moderator -> moderation
vocal -> permet de join un vocal

*/
//const emojiPersonInLotusPosition = "\U0001F9D8"

type Configuration struct {
	Channel string `json:"channel"`
	Messages []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Color       int    `json:"color"`
		Role        string `json:"role"`
		Emoji       string `json:"emoji"`
	} `json:"messages"`
}

type Manager struct {
	session *discordgo.Session
	config Configuration
}

func NewWelcomeManager(
	session *discordgo.Session,
	config Configuration,
) *Manager {
	return &Manager{
		session: session,
		config: config,
	}
}

func (w *Manager) Run() {
	fmt.Println("Welcome -> Run")

	w.ReplaceMessages()
	//w.session.AddHandler(w.onMessageReactionAdd)
}

func (w *Manager) ReplaceMessages() {
	/*messages, err := w.session.ChannelMessages(w.channelID, 100, "", "", "")
	if err != nil {
		log.Fatalf("could not read channel messages from channelID %s: %s", w.channelID, err)
	}

	embedMessages := w.embedMessages
	for _, message := range messages {
		if message.Author.ID != w.session.State.User.ID {
			continue
		}

		if len(message.Embeds) == 0 {
			continue
		}

		for idxEmbedMessage, embedMessage := range embedMessages {
			if message.Embeds[0].Title == embedMessage.Title {
				w.updateRole(message.ID, embedMessage)

				embedMessages[idxEmbedMessage] = embedMessages[len(embedMessages)-1]
				embedMessages = embedMessages[:len(embedMessages)-1]

				break
			}
		}
	}

	for _, embedMessage := range embedMessages {
		w.addEmbedMessage(embedMessage)
	}*/
/*
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
			for _, reaction := range messagesFromBot[0].Reactions {
				//spew.Dump(reaction)
			}

			r, _ := helpers.MessageReactionsAll(w.session, w.channelID, messagesFromBot[0].ID, emojiPersonInLotusPosition)
			for _, reaction := range r {
				spew.Dump(reaction.Username)
			}
		}
	}*/
}
/*
func (w *Manager) updateRole(messageID string, embedMessage EmbedMessage) {
	r, _ := helpers.MessageReactionsAll(w.session, w.channelID, messageID, embedMessage.Emoji)
	for _, reaction := range r {
		spew.Dump(reaction.Username)
	}
}

func (w *Manager) addEmbedMessage(embedMessage EmbedMessage) {
	message, err := w.session.ChannelMessageSendEmbed(w.channelID, &discordgo.MessageEmbed{
		Title: embedMessage.Title,
		Description: embedMessage.Description,
		Color: embedMessage.Color,
	})
	if err != nil {
		log.Fatalf("could not send embed message to channelID %s: %s", w.channelID, err)
	}

	err = w.session.MessageReactionAdd(w.channelID, message.ID, embedMessage.Emoji)
	if err != nil {
		log.Fatalf("could not add reaction to welcome message to channelID %s - messageID %s: %s", w.channelID, message.ID, err)
	}
}

func (w *Manager) onMessageReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	if reaction.ChannelID != w.channelID {
		return
	}

	spew.Dump(reaction)
	if reaction.Emoji.Name == emojiPersonInLotusPosition {
		fmt.Println("onMessageReactionAdd")
		err := w.session.GuildMemberRoleAdd(w.session.State.Guilds[0].ID, reaction.UserID, "")
		if err != nil {
			log.Fatalf("could not add roleID %s to userID %s: %s", "", reaction.UserID, err)
		}
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
*/