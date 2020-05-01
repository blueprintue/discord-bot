package welcome

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-colorable"
	"github.com/rancoud/blueprintue-discord/helpers"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetOutput(colorable.NewColorableStdout())
}

type Configuration struct {
	Channel string `json:"channel"`
	ChannelID string
	GuildID string
	Messages []Message `json:"messages"`
}

type Message struct {
	ID          string
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	Role        string `json:"role"`
	RoleID      string
	Emoji       string `json:"emoji"`
	EmojiID     string
}

type Manager struct {
	session *discordgo.Session
	guildName string
	config Configuration
}

func NewWelcomeManager(
	session *discordgo.Session,
	guildName string,
	config Configuration,
) *Manager {
	manager := &Manager{
		session: session,
		guildName: guildName,
		config: config,
	}

	manager.completeData()

	return manager
}

func (w *Manager) completeData() {
	for _, guild := range w.session.State.Guilds {
		if guild.Name == w.guildName {
			w.config.GuildID = guild.ID

			for _, channel := range guild.Channels {
				if channel.Name == w.config.Channel {
					w.config.ChannelID = channel.ID
					break
				}
			}

			for _, role := range guild.Roles {
				for idx := range w.config.Messages {
					if w.config.Messages[idx].Role == role.Name {
						w.config.Messages[idx].RoleID = role.ID
					}
				}
			}

			for _, emoji := range guild.Emojis {
				emojiRichEmbed := fmt.Sprintf("<:%s:%s>", emoji.Name, emoji.ID)
				emojiInText := ":" + emoji.Name + ":"
				for idx := range w.config.Messages {
					w.config.Messages[idx].Title = strings.ReplaceAll(w.config.Messages[idx].Title, emojiInText, emojiRichEmbed)
					w.config.Messages[idx].Description = strings.ReplaceAll(w.config.Messages[idx].Description, emojiInText, emojiRichEmbed)
					if w.config.Messages[idx].Emoji == emoji.Name {
						w.config.Messages[idx].EmojiID = emoji.ID
					}
				}
			}

			break
		}
	}
}

func (w *Manager) Run() {
	logrus.Info("[WELCOME]\tRun")

	w.ReplaceMessages()
	w.session.AddHandler(w.onMessageReactionAdd)
	w.session.AddHandler(w.onMessageReactionRemove)
}

func (w *Manager) ReplaceMessages() {
	messages, err := w.session.ChannelMessages(w.config.ChannelID, 100, "", "", "")
	if err != nil {
		log.Fatalf("[WELCOME]\tcould not read channel messages from channelID %s: %s", w.config.ChannelID, err)
	}

	var idxsMessageTreated []int
	for _, message := range messages {
		if message.Author.ID != w.session.State.User.ID {
			continue
		}

		if len(message.Embeds) == 0 {
			continue
		}

		for idxMessages := range w.config.Messages {
			if message.Embeds[0].Title == w.config.Messages[idxMessages].Title {
				w.config.Messages[idxMessages].ID = message.ID
				w.updateRoleBelongMessage(w.config.Messages[idxMessages])

				idxsMessageTreated = append(idxsMessageTreated, idxMessages)

				break
			}
		}
	}

	for idxMessages := range w.config.Messages {
		messageTreated := false
		for _, idxMessageTreated := range idxsMessageTreated {
			if idxMessageTreated == idxMessages {
				messageTreated = true
				break
			}
		}

		if !messageTreated {
			w.addEmbedMessage(&w.config.Messages[idxMessages])
			idxsMessageTreated = append(idxsMessageTreated, idxMessages)
		}
	}
}

func (w *Manager) updateRoleBelongMessage(message Message) {
	logrus.Infof("[WELCOME]\tupdateRoleBelongMessage for %s", message.Title)

	users, _ := helpers.MessageReactionsAll(w.session, w.config.ChannelID, message.ID, message.Emoji + ":" + message.EmojiID)
	for _, user := range users {
		var err error
		if w.isUserBot(user.ID) {
			continue
		}

		member, err := w.session.State.Member(w.config.GuildID, user.ID)
		if err != nil {
			log.Fatalf("[WELCOME]\tcould not find Member userID %s in GuildID %s: %s", user.ID, w.config.GuildID, err)
		}

		skipUser := false
		for _, role := range member.Roles {
			if role == message.RoleID {
				skipUser = true
				break
			}
		}

		if skipUser {
			continue
		}

		err = w.session.GuildMemberRoleAdd(w.config.GuildID, user.ID, message.RoleID)
		if err != nil {
			log.Fatalf("[WELCOME]\tcould not add roleID %s to userID %s: %s", message.RoleID, user.ID, err)
		}
	}
}

func (w *Manager) addEmbedMessage(message *Message) {
	logrus.Infof("[WELCOME]\taddEmbedMessage %s", message.Title)

	messageSent, err := w.session.ChannelMessageSendEmbed(w.config.ChannelID, &discordgo.MessageEmbed{
		Title: message.Title,
		Description: message.Description,
		Color: message.Color,
	})
	if err != nil {
		log.Fatalf("[WELCOME]\tcould not send embed message to channelID %s: %s", w.config.ChannelID, err)
	}

	message.ID = messageSent.ID

	err = w.session.MessageReactionAdd(w.config.ChannelID, message.ID, message.Emoji + ":" + message.EmojiID)
	if err != nil {
		log.Fatalf("[WELCOME]\tcould not add reaction to welcome message to channelID %s - messageID %s: %s", w.config.ChannelID, message.ID, err)
	}
}

func (w *Manager) onMessageReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	if reaction.ChannelID != w.config.ChannelID {
		return
	}

	if w.isUserBot(reaction.UserID) {
		return
	}

	idxMessageFound := -1
	for idxMessage := range w.config.Messages {
		if reaction.MessageID == w.config.Messages[idxMessage].ID {
			idxMessageFound = idxMessage
			break
		}
	}

	if idxMessageFound == -1 {
		return
	}

	/*member, err := w.session.State.Member(w.config.GuildID, reaction.UserID)
	if err != nil {
		log.Fatalf("[WELCOME]\tcould not find Member userID %s in GuildID %s: %s", reaction.UserID, w.config.GuildID, err)
	}*/

	/*if !member.User.MFAEnabled || !member.User.Verified {
		err = w.session.MessageReactionRemove(w.config.ChannelID, reaction.MessageID, w.config.Messages[idxMessageFound].Emoji + ":" + w.config.Messages[idxMessageFound].EmojiID, reaction.UserID)
		if err != nil {
			log.Fatalf("[WELCOME]\tcould not find Member userID %s in GuildID %s: %s", reaction.UserID, w.config.GuildID, err)
		}
		st, err := w.session.UserChannelCreate(reaction.UserID)
		if err != nil {
			log.Fatalf("[WELCOME]\tcould not create channel with User userID %s: %s", reaction.UserID, err)
		}
		_, err = w.session.ChannelMessageSend(st.ID, "Sorry it's not possible to add you role because your account is not verified and don't have multi-factor authentication enabled")
		if err != nil {
			log.Fatalf("[WELCOME]\tcould not create channel with User userID %s: %s", reaction.UserID, err)
		}
		return
	}*/

	if reaction.Emoji.Name == w.config.Messages[idxMessageFound].Emoji {
		logrus.Info("[WELCOME]\tonMessageReactionAdd")
		err := w.session.GuildMemberRoleAdd(w.config.GuildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
		if err != nil {
			log.Fatalf("[WELCOME]\tcould not add roleID %s to userID %s: %s", w.config.Messages[idxMessageFound].RoleID, reaction.UserID, err)
		}
	}
}

func (w *Manager) onMessageReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	if reaction.ChannelID != w.config.ChannelID {
		return
	}

	if w.isUserBot(reaction.UserID) {
		return
	}

	idxMessageFound := -1
	for idxMessage := range w.config.Messages {
		if reaction.MessageID == w.config.Messages[idxMessage].ID {
			idxMessageFound = idxMessage
			break
		}
	}

	if idxMessageFound == -1 {
		return
	}

	if reaction.Emoji.Name == w.config.Messages[idxMessageFound].Emoji {
		logrus.Info("[WELCOME]\tonMessageReactionRemove")
		err := w.session.GuildMemberRoleRemove(w.config.GuildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
		if err != nil {
			log.Fatalf("[WELCOME]\tcould not remove roleID %s to userID %s: %s", w.config.Messages[idxMessageFound].RoleID, reaction.UserID, err)
		}
	}
}

func (w *Manager) isUserBot(userID string) bool {
	return w.session.State.User.ID == userID
}