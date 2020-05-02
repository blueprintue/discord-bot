package welcome

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rancoud/blueprintue-discord/helpers"
	"github.com/sirupsen/logrus"
	"strings"
)

type Configuration struct {
	Channel string `json:"channel"`
	ChannelID string
	Guild string
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
	config Configuration
}

func NewWelcomeManager(
	session *discordgo.Session,
	guildName string,
	config Configuration,
) *Manager {
	config.Guild = guildName

	manager := &Manager{
		session: session,
		config: config,
	}

	logrus.Info("[WELCOME]\t[NewWelcomeManager]\tComplete configuration with session.State")
	manager.completeConfiguration()

	return manager
}

func (w *Manager) completeConfiguration() {
	for _, guild := range w.session.State.Guilds {
		if guild.Name == w.config.Guild {
			logrus.Infof("[WELCOME]\t[completeConfiguration]\tSet GuildID %s for Guild '%s'", guild.ID, w.config.Guild)
			w.config.GuildID = guild.ID

			for _, channel := range guild.Channels {
				if channel.Name == w.config.Channel {
					logrus.Infof("[WELCOME]\t[completeConfiguration]\tSet ChannelID %s for Channel '%s'", channel.ID, w.config.Channel)
					w.config.ChannelID = channel.ID
					break
				}
			}

			for _, role := range guild.Roles {
				for idx := range w.config.Messages {
					if w.config.Messages[idx].Role == role.Name {
						logrus.Infof("[WELCOME]\t[completeConfiguration]\tSet RoleID %s for Role '%s'", role.ID, w.config.Messages[idx].Role)
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
						logrus.Infof("[WELCOME]\t[completeConfiguration]\tSet EmojiID %s for Emoji '%s'", emoji.ID, w.config.Messages[idx].Emoji)
						w.config.Messages[idx].EmojiID = emoji.ID
					}
				}
			}

			break
		}
	}
}

func (w *Manager) Run() error {
	logrus.Info("[WELCOME]\t[Run]\tAdd Handler on Message Reaction Add")
	w.session.AddHandler(w.onMessageReactionAdd)

	logrus.Info("[WELCOME]\t[Run]\tAdd Handler on Message Reaction Remove")
	w.session.AddHandler(w.onMessageReactionRemove)

	logrus.Info("[WELCOME]\t[Run]\tAdd messages to channel")
	err := w.addMessagesToChannel()
	if err != nil {
		logrus.Errorf("[WELCOME]\t[Run]\tCould not add messages to channel: %s", err)
		return err
	}

	return nil
}

func (w *Manager) addMessagesToChannel() error {
	logrus.Infof("[WELCOME]\t[addMessagesToChannel]\tGet messages from channelID %s (%s)", w.config.ChannelID, w.config.Channel)
	messages, err := w.session.ChannelMessages(w.config.ChannelID, 100, "", "", "")
	if err != nil {
		logrus.Errorf("[WELCOME]\t[addMessagesToChannel]\tCould not read messages from channelID %s (%s): %s", w.config.ChannelID, w.config.Channel, err)
		return err
	}

	var idxsMessageTreated []int
	for _, message := range messages {
		if message.Author.ID != w.session.State.User.ID {
			logrus.Infof("[WELCOME]\t[addMessagesToChannel]\tSKIP - Message in Channel %s is not from bot", w.config.Channel)
			continue
		}

		if len(message.Embeds) == 0 {
			logrus.Infof("[WELCOME]\t[addMessagesToChannel]\tSKIP - Message in Channel %s is not an embed", w.config.Channel)
			continue
		}

		for idxMessages := range w.config.Messages {
			if message.Embeds[0].Title == w.config.Messages[idxMessages].Title {
				w.config.Messages[idxMessages].ID = message.ID

				logrus.Infof("[WELCOME]\t[addMessagesToChannel]\tMessage '%s' already sent -> update roles", w.config.Messages[idxMessages].Title)
				err := w.updateRoleBelongMessage(w.config.Messages[idxMessages])
				if err != nil {
					logrus.Errorf("[WELCOME]\t[addMessagesToChannel]\tCould not update role belong to Message '%s': %s", w.config.Messages[idxMessages].Title, err)
					return err
				}

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
			logrus.Infof("[WELCOME]\t[addMessagesToChannel]\tMessage %s missing - add message", w.config.Messages[idxMessages].Title)
			err := w.addMessage(&w.config.Messages[idxMessages])
			if err != nil {
				logrus.Errorf("[WELCOME]\t[addMessagesToChannel]\tCould not add Message %s: %s", w.config.Messages[idxMessages].Title, err)
				return err
			}

			idxsMessageTreated = append(idxsMessageTreated, idxMessages)
		}
	}

	return nil
}

func (w *Manager) updateRoleBelongMessage(message Message) error {
	logrus.Infof("[WELCOME]\t[updateRoleBelongMessage]\tGet all reactions from Message %s", message.Title)
	users, err := helpers.MessageReactionsAll(w.session, w.config.ChannelID, message.ID, message.Emoji + ":" + message.EmojiID)
	if err != nil {
		logrus.Errorf("[WELCOME]\t[updateRoleBelongMessage]\tCould not get all reactions from MessageID %s ChannelID %s (%s) Emoji %s: %s", message.ID, w.config.ChannelID, w.config.Channel, message.Emoji + ":" + message.EmojiID, err)
		return err
	}

	for _, user := range users {
		if w.isUserBot(user.ID) {
			logrus.Info("[WELCOME]\t[updateRoleBelongMessage]\tSKIP - User is the bot")
			continue
		}

		member, err := w.session.State.Member(w.config.GuildID, user.ID)
		if err != nil {
			logrus.Errorf("[WELCOME]\t[updateRoleBelongMessage]\tCould not find Member userID %s in GuildID %s: %s", user.ID, w.config.GuildID, err)
			return err
		}

		skipUser := false
		for _, role := range member.Roles {
			if role == message.RoleID {
				skipUser = true
				break
			}
		}

		if skipUser {
			logrus.Info("[WELCOME]\t[updateRoleBelongMessage]\tSKIP - User has already role")
			continue
		}

		logrus.Infof("[WELCOME]\t[updateRoleBelongMessage]\tAdd RoleID %s (%s) to UserID %s (%s)", message.RoleID, message.Role, user.ID, user.Username)
		err = w.session.GuildMemberRoleAdd(w.config.GuildID, user.ID, message.RoleID)
		if err != nil {
			logrus.Errorf("[WELCOME]\t[updateRoleBelongMessage]\tCould not add roleID %s to userID %s: %s", message.RoleID, user.ID, err)
			return err
		}
	}

	return nil
}

func (w *Manager) addMessage(message *Message) error {
	logrus.Infof("[WELCOME]\t[addMessage]\tSent Message %s to ChannelID %s (%s)", message.Title, w.config.ChannelID, w.config.Channel)
	messageSent, err := w.session.ChannelMessageSendEmbed(w.config.ChannelID, &discordgo.MessageEmbed{
		Title: message.Title,
		Description: message.Description,
		Color: message.Color,
	})
	if err != nil {
		logrus.Errorf("[WELCOME]\t[addMessage]\tCould not send Message to ChannelID %s: %s", w.config.ChannelID, err)
		return err
	}

	logrus.Infof("[WELCOME]\t[addMessage]\tMessage Sent with ID %s", messageSent.ID)
	message.ID = messageSent.ID

	logrus.Infof("[WELCOME]\t[addMessage]\tAdd Reaction to Message %s", message.Title)
	err = w.session.MessageReactionAdd(w.config.ChannelID, message.ID, message.Emoji + ":" + message.EmojiID)
	if err != nil {
		logrus.Errorf("[WELCOME]\t[addMessage]\tCould not add Reaction to MessageID %s: %s", message.ID, err)
		return err
	}

	return nil
}

func (w *Manager) onMessageReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	logrus.Info("[WELCOME]\t[onMessageReactionAdd]\tIncoming Message Reaction Add")
	if reaction.ChannelID != w.config.ChannelID {
		logrus.Info("[WELCOME]\t[onMessageReactionAdd]\tSKIP - Channel is not matching")
		return
	}

	if w.isUserBot(reaction.UserID) {
		logrus.Info("[WELCOME]\t[onMessageReactionAdd]\tSKIP - User is the bot")
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
		logrus.Info("[WELCOME]\t[onMessageReactionAdd]\tSKIP - Message is not matching")
		return
	}

	if reaction.Emoji.Name != w.config.Messages[idxMessageFound].Emoji {
		logrus.Info("[WELCOME]\t[onMessageReactionAdd]\tSKIP - Emoji is not matching")
		return
	}

	logrus.Infof("[WELCOME]\t[onMessageReactionAdd]\tAdd RoleID %s (%s) to UserID %s", w.config.Messages[idxMessageFound].RoleID, w.config.Messages[idxMessageFound].Role, reaction.UserID)
	err := w.session.GuildMemberRoleAdd(w.config.GuildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
	if err != nil {
		logrus.Errorf("[WELCOME]\t[onMessageReactionAdd]\tCould not add roleID %s to userID %s: %s", w.config.Messages[idxMessageFound].RoleID, reaction.UserID, err)
	}
}

func (w *Manager) onMessageReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	logrus.Info("[WELCOME]\t[onMessageReactionRemove]\tIncoming Message Reaction Remove")
	if reaction.ChannelID != w.config.ChannelID {
		logrus.Infof("[WELCOME]\t[onMessageReactionRemove]\tSKIP - Channel is not matching")
		return
	}

	if w.isUserBot(reaction.UserID) {
		logrus.Info("[WELCOME]\t[onMessageReactionRemove]\tSKIP - User is the bot")
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
		logrus.Info("[WELCOME]\t[onMessageReactionRemove]\tSKIP - Message is not matching")
		return
	}

	if reaction.Emoji.Name != w.config.Messages[idxMessageFound].Emoji {
		logrus.Info("[WELCOME]\t[onMessageReactionRemove]\tSKIP - Emoji is not matching")
		return
	}

	logrus.Infof("[WELCOME]\t[onMessageReactionRemove]\tRemove RoleID %s (%s) to UserID %s", w.config.Messages[idxMessageFound].RoleID, w.config.Messages[idxMessageFound].Role, reaction.UserID)
	err := w.session.GuildMemberRoleRemove(w.config.GuildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
	if err != nil {
		logrus.Errorf("[WELCOME]\t[onMessageReactionRemove]\tCould not remove roleID %s to userID %s: %s", w.config.Messages[idxMessageFound].RoleID, reaction.UserID, err)
	}
}

func (w *Manager) isUserBot(userID string) bool {
	return w.session.State.User.ID == userID
}

func (w *Manager) sendMessageToUser(userID string) error {
	st, err := w.session.UserChannelCreate(userID)
	if err != nil {
		logrus.Errorf("[WELCOME]\t[sendMessageToUser]\tCould not create Channel with UserID %s: %s", userID, err)
		return err
	}

	_, err = w.session.ChannelMessageSend(st.ID, "Sorry it's not possible to add role because your account is not verified and you don't have multi-factor authentication enabled")
	if err != nil {
		logrus.Errorf("[WELCOME]\t[sendMessageToUser]\tCould not send Message to UserID %s: %s", userID, err)
		return err
	}

	return nil
}