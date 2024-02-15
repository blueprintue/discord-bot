package welcome

import (
	"fmt"
	"strings"

	"github.com/blueprintue/discord-bot/helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const limitChannelMessages = 100

// Configuration is a struct.
type Configuration struct {
	Channel   string    `json:"channel"`
	ChannelID string    `json:"-"`
	Guild     string    `json:"-"`
	GuildID   string    `json:"-"`
	Messages  []Message `json:"messages"`
}

// Message is a struct.
type Message struct {
	ID                               string
	Title                            string `json:"title"`
	Description                      string `json:"description"`
	Color                            int    `json:"color"`
	Role                             string `json:"role"`
	RoleID                           string
	Emoji                            string `json:"emoji"`
	EmojiID                          string
	CanPurgeReactions                bool `json:"can_purge_reactions"`
	PurgeThresholdMembersReacted     int  `json:"purge_threshold_members_reacted"`
	PurgeBelowCountMembersNotInGuild int  `json:"purge_below_count_members_not_in_guild"`
}

// Manager is a struct.
type Manager struct {
	session *discordgo.Session
	config  Configuration
}

// NewWelcomeManager return a Manager.
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

	log.Info().Msg("Checking configuration 1/2")

	if !manager.hasValidConfigurationInFile() {
		return nil
	}

	log.Info().Msg("Completing configuration with session.State")
	manager.completeConfiguration()

	log.Info().Msg("Checking configuration 2/2")

	if !manager.hasValidConfigurationAgainstDiscordServer() {
		return nil
	}

	return manager
}

func (w *Manager) hasValidConfigurationInFile() bool {
	if w.config.Channel == "" {
		log.Error().
			Msg("Channel is empty")

		return false
	}

	for idx, message := range w.config.Messages {
		if message.Title == "" && message.Description == "" {
			log.Error().
				Int("message #", idx).
				Str("title", message.Title).
				Str("description", message.Description).
				Msg("Title and Description is empty")

			return false
		}

		if message.Emoji == "" {
			log.Error().
				Int("message #", idx).
				Msg("Emoji is empty")

			return false
		}

		if message.Role == "" {
			log.Error().
				Int("message #", idx).
				Msg("Role is empty")

			return false
		}
	}

	return true
}

//nolint:gocognit,cyclop
func (w *Manager) completeConfiguration() {
	for _, guild := range w.session.State.Guilds {
		if guild.Name == w.config.Guild {
			log.Info().
				Str("guild_id", guild.ID).
				Str("guild", w.config.Guild).
				Msg("Set GuildID")

			w.config.GuildID = guild.ID

			for _, channel := range guild.Channels {
				if channel.Name == w.config.Channel {
					log.Info().
						Str("channel_id", channel.ID).
						Str("channel", w.config.Channel).
						Msg("Set ChannelID")

					w.config.ChannelID = channel.ID

					break
				}
			}

			for _, role := range guild.Roles {
				for idx := range w.config.Messages {
					if w.config.Messages[idx].Role == role.Name {
						log.Info().
							Str("role_id", role.ID).
							Str("role", w.config.Messages[idx].Role).
							Msg("Set RoleID")

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
						log.Info().
							Str("emoji_id", emoji.ID).
							Str("emoji", w.config.Messages[idx].Emoji).
							Msg("Set EmojiID")

						w.config.Messages[idx].EmojiID = emoji.ID
					}
				}
			}

			break
		}
	}
}

func (w *Manager) hasValidConfigurationAgainstDiscordServer() bool {
	if w.config.ChannelID == "" {
		log.Error().
			Str("channel in config", w.config.Channel).
			Msg("Channel not found in Discord server")

		return false
	}

	for idx, message := range w.config.Messages {
		if message.EmojiID == "" {
			log.Error().
				Int("message #", idx).
				Str("emoji in config", message.Emoji).
				Msg("Emoji not found in Discord server")

			return false
		}

		if message.RoleID == "" {
			log.Error().
				Int("message #", idx).
				Str("role in config", message.Role).
				Msg("Role not found in Discord server")

			return false
		}
	}

	return true
}

// Run do the main task of Welcome.
func (w *Manager) Run() error {
	log.Info().Msg("Adding Handler on Message Reaction Add")
	w.session.AddHandler(w.onMessageReactionAdd)

	log.Info().Msg("Adding Handler on Message Reaction Remove")
	w.session.AddHandler(w.onMessageReactionRemove)

	log.Info().Msg("Adding messages to channel")

	err := w.addMessagesToChannel()
	if err != nil {
		log.Error().Err(err).Msg("Could not add messages to channel")

		return err
	}

	return nil
}

//nolint:funlen,cyclop
func (w *Manager) addMessagesToChannel() error {
	log.Info().
		Str("channel_id", w.config.ChannelID).
		Str("channel", w.config.Channel).
		Msg("Getting Messages from Channel")

	messages, err := w.session.ChannelMessages(w.config.ChannelID, limitChannelMessages, "", "", "")

	if err != nil {
		log.Error().Err(err).
			Str("channel_id", w.config.ChannelID).
			Str("channel", w.config.Channel).
			Msg("Could not read Messages from Channel")

		return fmt.Errorf("%w", err)
	}

	var idxsMessageTreated []int

	for _, message := range messages {
		if message.Author.ID != w.session.State.User.ID {
			log.Info().
				Str("message_id", message.ID).
				Str("channel_id", w.config.ChannelID).
				Str("channel", w.config.Channel).
				Msg("SKIP - Message in Channel is not from bot")

			continue
		}

		if len(message.Embeds) == 0 {
			log.Info().
				Str("message_id", message.ID).
				Str("channel_id", w.config.ChannelID).
				Str("channel", w.config.Channel).
				Msg("SKIP - Message in Channel is not an embed")

			continue
		}

		for idxMessages := range w.config.Messages {
			if message.Embeds[0].Title == w.config.Messages[idxMessages].Title {
				w.config.Messages[idxMessages].ID = message.ID

				log.Info().
					Str("message_title", w.config.Messages[idxMessages].Title).
					Msg("Message already sent -> update roles")

				err := w.updateRoleBelongMessage(w.config.Messages[idxMessages])
				if err != nil {
					log.Error().Err(err).
						Str("message_title", w.config.Messages[idxMessages].Title).
						Msg("Could not update role belong to Message")

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
			log.Info().
				Str("message_title", w.config.Messages[idxMessages].Title).
				Msg("Message missing - add Message")

			err := w.addMessage(&w.config.Messages[idxMessages])
			if err != nil {
				log.Error().Err(err).
					Str("message_title", w.config.Messages[idxMessages].Title).
					Msg("Could not add Message")

				return err
			}

			idxsMessageTreated = append(idxsMessageTreated, idxMessages)
		}
	}

	return nil
}

//nolint:funlen,cyclop
func (w *Manager) updateRoleBelongMessage(message Message) error {
	log.Info().
		Str("message_id", message.ID).
		Str("message_title", message.Title).
		Str("channel_id", w.config.ChannelID).
		Str("channel", w.config.Channel).
		Str("emoji", message.Emoji+":"+message.EmojiID).
		Msg("Getting all Reactions from Message")

	users, err := helpers.MessageReactionsAll(w.session, w.config.ChannelID, message.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).
			Str("message_id", message.ID).
			Str("channel_id", w.config.ChannelID).
			Str("channel", w.config.Channel).
			Str("emoji", message.Emoji+":"+message.EmojiID).
			Msg("Could not get all Reactions")

		return fmt.Errorf("%w", err)
	}

	membersNotInGuild := []string{}

	for _, user := range users {
		if w.isUserBot(user.ID) {
			log.Info().
				Str("user_id", user.ID).
				Msg("SKIP - User is the bot")

			continue
		}

		member, err := w.session.State.Member(w.config.GuildID, user.ID)
		if err != nil {
			log.Error().Err(err).
				Str("user_id", user.ID).
				Str("guild_id", w.config.GuildID).
				Msg("Could not find Member in Guild")

			membersNotInGuild = append(membersNotInGuild, user.ID)

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
			log.Info().
				Str("user_id", user.ID).
				Str("guild_id", w.config.GuildID).
				Msg("SKIP - User has already Role")

			continue
		}

		log.Info().
			Str("role_id", message.RoleID).
			Str("role", message.Role).
			Str("user_id", user.ID).
			Str("username", user.Username).
			Msg("Add Role to User")

		err = w.session.GuildMemberRoleAdd(w.config.GuildID, user.ID, message.RoleID)
		if err != nil {
			log.Error().Err(err).
				Str("role_id", message.RoleID).
				Str("role", message.Role).
				Str("user_id", user.ID).
				Str("username", user.Username).
				Msg("Could not add Role to User")

			return fmt.Errorf("%w", err)
		}
	}

	if len(membersNotInGuild) > 0 {
		log.Info().
			Int("count_members_reacted", len(users)).
			Int("count_members_not_found", len(membersNotInGuild)).
			Msg("Members not found in Guild")

		// use constant from config file
		if message.CanPurgeReactions &&
			len(users) > message.PurgeThresholdMembersReacted &&
			len(membersNotInGuild) < message.PurgeBelowCountMembersNotInGuild {
			log.Info().
				Msg("Do purge")

			for idx := range membersNotInGuild {
				err = w.session.MessageReactionRemove(w.config.ChannelID, message.ID, message.Emoji+":"+message.EmojiID, membersNotInGuild[idx])
				if err != nil {
					log.Error().Err(err).
						Str("role_id", message.RoleID).
						Str("role", message.Role).
						Str("user_id", membersNotInGuild[idx]).
						Msg("Could not remove Reaction")
				}
			}
		}
	}

	return nil
}

func (w *Manager) addMessage(message *Message) error {
	log.Info().
		Str("message_title", message.Title).
		Str("channel_id", w.config.ChannelID).
		Str("channel", w.config.Channel).
		Msg("Sending Message to Channel")

	messageSent, err := w.session.ChannelMessageSendEmbed(w.config.ChannelID, &discordgo.MessageEmbed{
		Title:       message.Title,
		Description: message.Description,
		Color:       message.Color,
	})

	if err != nil {
		log.Error().Err(err).
			Str("message_title", message.Title).
			Str("channel_id", w.config.ChannelID).
			Str("channel", w.config.Channel).
			Msg("Could not send Message to Channel")

		return fmt.Errorf("%w", err)
	}

	log.Info().
		Str("message_id", messageSent.ID).
		Str("channel_id", w.config.ChannelID).
		Str("channel", w.config.Channel).
		Msg("Message Sent")

	message.ID = messageSent.ID

	log.Info().
		Str("message_id", messageSent.ID).
		Str("message_title", message.Title).
		Str("emoji", message.Emoji+":"+message.EmojiID).
		Msg("Adding Reaction to Message")

	err = w.session.MessageReactionAdd(w.config.ChannelID, message.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).
			Str("message_id", message.ID).
			Str("emoji", message.Emoji+":"+message.EmojiID).
			Msg("Could not add Reaction to Message")

		return fmt.Errorf("%w", err)
	}

	return nil
}

func (w *Manager) onMessageReactionAdd(_ *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	log.Info().
		Str("channel_id", reaction.ChannelID).
		Msg("Incoming Message Reaction Add")

	if reaction.ChannelID != w.config.ChannelID {
		log.Info().Msg("SKIP - Channel is not matching")

		return
	}

	if w.isUserBot(reaction.UserID) {
		log.Info().
			Str("channel_id", reaction.ChannelID).
			Msg("SKIP - User is the bot")

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
		log.Info().
			Str("channel_id", reaction.ChannelID).
			Msg("SKIP - Message is not matching")

		return
	}

	if reaction.Emoji.Name != w.config.Messages[idxMessageFound].Emoji {
		log.Info().
			Str("channel_id", reaction.ChannelID).
			Msg("SKIP - Emoji is not matching")

		return
	}

	log.Info().
		Str("role_id", w.config.Messages[idxMessageFound].RoleID).
		Str("role", w.config.Messages[idxMessageFound].Role).
		Str("channel_id", reaction.ChannelID).
		Str("user_id", reaction.UserID).
		Msg("Adding Role to User")

	err := w.session.GuildMemberRoleAdd(w.config.GuildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
	if err != nil {
		log.Error().Err(err).
			Str("role_id", w.config.Messages[idxMessageFound].RoleID).
			Str("role", w.config.Messages[idxMessageFound].Role).
			Str("channel_id", reaction.ChannelID).
			Str("user_id", reaction.UserID).
			Msg("Could not add Role to User")
	}
}

func (w *Manager) onMessageReactionRemove(_ *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	log.Info().Msg("Incoming Message Reaction Remove")

	if reaction.ChannelID != w.config.ChannelID {
		log.Info().
			Str("channel_id", reaction.ChannelID).
			Msg("SKIP - Channel is not matching")

		return
	}

	if w.isUserBot(reaction.UserID) {
		log.Info().
			Str("user_id", reaction.UserID).
			Msg("SKIP - User is the bot")

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
		log.Info().
			Str("emoji", reaction.Emoji.Name).
			Msg("SKIP - Emoji is not matching")

		return
	}

	log.Info().
		Str("role_id", w.config.Messages[idxMessageFound].RoleID).
		Str("role", w.config.Messages[idxMessageFound].Role).
		Str("user_id", reaction.UserID).
		Msg("Removing Role to User")

	err := w.session.GuildMemberRoleRemove(w.config.GuildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
	if err != nil {
		log.Error().Err(err).
			Str("role_id", w.config.Messages[idxMessageFound].RoleID).
			Str("role", w.config.Messages[idxMessageFound].Role).
			Str("user_id", reaction.UserID).
			Msg("Could not remove Role to User")
	}
}

func (w *Manager) isUserBot(userID string) bool {
	return w.session.State.User.ID == userID
}
