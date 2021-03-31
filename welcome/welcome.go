package welcome

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/bwmarrin/discordgo"
	"github.com/rancoud/blueprintue-discord/helpers"
)

// Configuration is a struct
type Configuration struct {
	Channel   string `json:"channel"`
	ChannelID string
	Guild     string
	GuildID   string
	Messages  []Message `json:"messages"`
}

// Message is a struct
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

// Manager is a struct
type Manager struct {
	session *discordgo.Session
	config  Configuration
}

// NewWelcomeManager return a Manager
func NewWelcomeManager(
	session *discordgo.Session,
	guildName string,
	config Configuration,
) *Manager {
	config.Guild = guildName

	manager := &Manager{
		session: session,
		config:  config,
	}

	log.Info().Msg("Complete configuration with session.State")
	manager.completeConfiguration()

	return manager
}

func (w *Manager) completeConfiguration() {
	for _, guild := range w.session.State.Guilds {
		if guild.Name == w.config.Guild {
			log.Info().Msgf("Set GuildID %s for Guild '%s'", guild.ID, w.config.Guild)
			w.config.GuildID = guild.ID

			for _, channel := range guild.Channels {
				if channel.Name == w.config.Channel {
					log.Info().Msgf("Set ChannelID %s for Channel '%s'", channel.ID, w.config.Channel)
					w.config.ChannelID = channel.ID
					break
				}
			}

			for _, role := range guild.Roles {
				for idx := range w.config.Messages {
					if w.config.Messages[idx].Role == role.Name {
						log.Info().Msgf("Set RoleID %s for Role '%s'", role.ID, w.config.Messages[idx].Role)
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
						log.Info().Msgf("Set EmojiID %s for Emoji '%s'", emoji.ID, w.config.Messages[idx].Emoji)
						w.config.Messages[idx].EmojiID = emoji.ID
					}
				}
			}

			break
		}
	}
}

// Run do the main task of Welcome
func (w *Manager) Run() error {
	log.Info().Msg("Add Handler on Message Reaction Add")
	w.session.AddHandler(w.onMessageReactionAdd)

	log.Info().Msg("Add Handler on Message Reaction Remove")
	w.session.AddHandler(w.onMessageReactionRemove)

	log.Info().Msg("Add messages to channel")
	err := w.addMessagesToChannel()
	if err != nil {
		log.Error().Err(err).Msg("Could not add messages to channel")
		return err
	}

	return nil
}

func (w *Manager) addMessagesToChannel() error {
	log.Info().Msgf("Get messages from channelID %s (%s)", w.config.ChannelID, w.config.Channel)
	messages, err := w.session.ChannelMessages(w.config.ChannelID, 100, "", "", "")
	if err != nil {
		log.Error().Err(err).Msgf("Could not read messages from channelID %s (%s)", w.config.ChannelID, w.config.Channel)
		return err
	}

	var idxsMessageTreated []int
	for _, message := range messages {
		if message.Author.ID != w.session.State.User.ID {
			log.Info().Msgf("SKIP - Message in Channel %s is not from bot", w.config.Channel)
			continue
		}

		if len(message.Embeds) == 0 {
			log.Info().Msgf("SKIP - Message in Channel %s is not an embed", w.config.Channel)
			continue
		}

		for idxMessages := range w.config.Messages {
			if message.Embeds[0].Title == w.config.Messages[idxMessages].Title {
				w.config.Messages[idxMessages].ID = message.ID

				log.Info().Msgf("Message '%s' already sent -> update roles", w.config.Messages[idxMessages].Title)
				err := w.updateRoleBelongMessage(w.config.Messages[idxMessages])
				if err != nil {
					log.Error().Err(err).Msgf("Could not update role belong to Message '%s'", w.config.Messages[idxMessages].Title)
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
			log.Info().Msgf("Message %s missing - add message", w.config.Messages[idxMessages].Title)
			err := w.addMessage(&w.config.Messages[idxMessages])
			if err != nil {
				log.Error().Err(err).Msgf("Could not add Message %s", w.config.Messages[idxMessages].Title)
				return err
			}

			idxsMessageTreated = append(idxsMessageTreated, idxMessages)
		}
	}

	return nil
}

func (w *Manager) updateRoleBelongMessage(message Message) error {
	log.Info().Msgf("Get all reactions from Message %s", message.Title)
	users, err := helpers.MessageReactionsAll(w.session, w.config.ChannelID, message.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).Msgf("Could not get all reactions from MessageID %s ChannelID %s (%s) Emoji %s", message.ID, w.config.ChannelID, w.config.Channel, message.Emoji+":"+message.EmojiID)
		return err
	}

	for _, user := range users {
		if w.isUserBot(user.ID) {
			log.Info().Msg("SKIP - User is the bot")
			continue
		}

		member, err := w.session.State.Member(w.config.GuildID, user.ID)
		if err != nil {
			log.Error().Err(err).Msgf("Could not find Member userID %s in GuildID %s", user.ID, w.config.GuildID)
			continue
		}

		skipUser := false
		for _, role := range member.Roles {
			if role == message.RoleID {
				skipUser = true
				break
			}
		}

		if skipUser {
			log.Info().Msg("SKIP - User has already role")
			continue
		}

		log.Info().Msgf("Add RoleID %s (%s) to UserID %s (%s)", message.RoleID, message.Role, user.ID, user.Username)
		err = w.session.GuildMemberRoleAdd(w.config.GuildID, user.ID, message.RoleID)
		if err != nil {
			log.Error().Err(err).Msgf("Could not add roleID %s to userID %s", message.RoleID, user.ID)
			return err
		}
	}

	return nil
}

func (w *Manager) addMessage(message *Message) error {
	log.Info().Msgf("Sent Message %s to ChannelID %s (%s)", message.Title, w.config.ChannelID, w.config.Channel)
	messageSent, err := w.session.ChannelMessageSendEmbed(w.config.ChannelID, &discordgo.MessageEmbed{
		Title:       message.Title,
		Description: message.Description,
		Color:       message.Color,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Could not send Message to ChannelID %s: %s", w.config.ChannelID)
		return err
	}

	log.Info().Msgf("Message Sent with ID %s", messageSent.ID)
	message.ID = messageSent.ID

	log.Info().Msgf("Add Reaction to Message %s", message.Title)
	err = w.session.MessageReactionAdd(w.config.ChannelID, message.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).Msgf("Could not add Reaction to MessageID %s: %s", message.ID)
		return err
	}

	return nil
}

func (w *Manager) onMessageReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	log.Info().Msg("Incoming Message Reaction Add")
	if reaction.ChannelID != w.config.ChannelID {
		log.Info().Msg("SKIP - Channel is not matching")
		return
	}

	if w.isUserBot(reaction.UserID) {
		log.Info().Msg("SKIP - User is the bot")
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
		log.Info().Msg("SKIP - Message is not matching")
		return
	}

	if reaction.Emoji.Name != w.config.Messages[idxMessageFound].Emoji {
		log.Info().Msg("SKIP - Emoji is not matching")
		return
	}

	log.Info().Msgf("Add RoleID %s (%s) to UserID %s", w.config.Messages[idxMessageFound].RoleID, w.config.Messages[idxMessageFound].Role, reaction.UserID)
	err := w.session.GuildMemberRoleAdd(w.config.GuildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
	if err != nil {
		log.Error().Err(err).Msgf("Could not add roleID %s to userID %s", w.config.Messages[idxMessageFound].RoleID, reaction.UserID)
	}
}

func (w *Manager) onMessageReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	log.Info().Msg("Incoming Message Reaction Remove")
	if reaction.ChannelID != w.config.ChannelID {
		log.Info().Msgf("SKIP - Channel is not matching")
		return
	}

	if w.isUserBot(reaction.UserID) {
		log.Info().Msg("SKIP - User is the bot")
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
		log.Info().Msg("SKIP - Message is not matching")
		return
	}

	if reaction.Emoji.Name != w.config.Messages[idxMessageFound].Emoji {
		log.Info().Msg("SKIP - Emoji is not matching")
		return
	}

	log.Info().Msgf("Remove RoleID %s (%s) to UserID %s", w.config.Messages[idxMessageFound].RoleID, w.config.Messages[idxMessageFound].Role, reaction.UserID)
	err := w.session.GuildMemberRoleRemove(w.config.GuildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
	if err != nil {
		log.Error().Err(err).Msgf("Could not remove roleID %s to userID %s: %s", w.config.Messages[idxMessageFound].RoleID, reaction.UserID)
	}
}

func (w *Manager) isUserBot(userID string) bool {
	return w.session.State.User.ID == userID
}

func (w *Manager) sendMessageToUser(userID string) error {
	st, err := w.session.UserChannelCreate(userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Could not create Channel")
		return err
	}

	_, err = w.session.ChannelMessageSend(st.ID, "Sorry it's not possible to add role because your account is not verified and you don't have multi-factor authentication enabled")
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msgf("Could not send Message")
		return err
	}

	return nil
}
