// Package welcome defines configuration struct, add messages, update roles.
package welcome

import (
	"fmt"
	"slices"
	"strings"

	"github.com/blueprintue/discord-bot/helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const (
	limitChannelMessages int = 100
	stepOneConfiguration int = 1
	stepTwoConfiguration int = 2
)

// Configuration is a struct.
type Configuration struct {
	Channel   string    `json:"channel"`
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
	discordSession *discordgo.Session
	config Configuration

	guildName   string
	guildID     string
	channelName string
	channelID   string
	messages    []Message
}

// NewWelcomeManager return a Manager.
func NewWelcomeManager(
	config Configuration,
	guildName string,
	discordSession *discordgo.Session,
) *Manager {
	manager := &Manager{
		discordSession: discordSession,
		config:  config,
		guildName: guildName,
	}

	log.Info().
		Str("package", "welcome").
		Msg("discord_bot.welcome.validating_configuration")

	if !hasValidConfigurationInFile(config) {
		log.Error().
			Str("package", "welcome").
			Int("step", stepOneConfiguration).
			Msg("discord_bot.welcome.configuration_validation_failed")

		return nil
	}

	log.Info().
		Str("package", "welcome").
		Msg("Completing configuration with session.State")

	manager.completeConfiguration(config)

	log.Info().
		Str("package", "welcome").
		Msg("Checking configuration 2/2")

	if !manager.hasValidConfigurationAgainstDiscordServer() {
		log.Error().
			Str("package", "welcome").
			Int("step", stepTwoConfiguration).
			Msg("discord_bot.welcome.configuration_validation_failed")

		return nil
	}

	return manager
}

func hasValidConfigurationInFile(config Configuration) bool {
	if config.Channel == "" {
		log.Error().
			Str("package", "welcome").
			Msg("discord_bot.welcome.empty_channel")

		return false
	}

	if len(config.Messages) == 0 {
		log.Error().
			Str("package", "welcome").
			Msg("discord_bot.welcome.empty_messages")

		return false
	}

	for idx, message := range config.Messages {
		if message.Title == "" && message.Description == "" {
			log.Error().
				Str("package", "welcome").
				Int("message index", idx).
				Str("title", message.Title).
				Str("description", message.Description).
				Msg("discord_bot.welcome.empty_title_description_message")

			return false
		}

		if message.Emoji == "" {
			log.Error().
				Str("package", "welcome").
				Int("message index", idx).
				Msg("discord_bot.welcome.empty_emoji_message")

			return false
		}

		if message.Role == "" {
			log.Error().
				Str("package", "welcome").
				Int("message index", idx).
				Msg("discord_bot.welcome.empty_role_message")

			return false
		}
	}

	return true
}

//nolint:cyclop,funlen
func (w *Manager) completeConfiguration(config Configuration) {
	for _, guild := range w.discordSession.State.Guilds {
		if guild.Name != w.guildName {
			continue
		}

		log.Info().
			Str("package", "welcome").
			Str("guild_id", guild.ID).
			Str("guild", w.guildName).
			Msg("Set GuildID")

		w.guildID = guild.ID

		for _, channel := range guild.Channels {
			if channel.Name != config.Channel {
				continue
			}

			log.Info().
				Str("package", "welcome").
				Str("channel_id", channel.ID).
				Str("channel", config.Channel).
				Msg("Set ChannelID")

			w.channelID = channel.ID
			w.channelName = config.Channel

			break
		}

		for _, role := range guild.Roles {
			for idx := range w.config.Messages {
				if config.Messages[idx].Role != role.Name {
					continue
				}

				log.Info().
					Str("package", "welcome").
					Str("role_id", role.ID).
					Str("role", w.config.Messages[idx].Role).
					Msg("Set RoleID")

				w.config.Messages[idx].RoleID = role.ID
			}
		}

		for _, emoji := range guild.Emojis {
			emojiRichEmbed := fmt.Sprintf("<:%s:%s>", emoji.Name, emoji.ID)
			emojiInText := ":" + emoji.Name + ":"

			for idx := range w.config.Messages {
				config.Messages[idx].Title = strings.ReplaceAll(w.config.Messages[idx].Title, emojiInText, emojiRichEmbed)
				config.Messages[idx].Description = strings.ReplaceAll(w.config.Messages[idx].Description, emojiInText, emojiRichEmbed)

				if config.Messages[idx].Emoji != emoji.Name {
					continue
				}

				log.Info().
					Str("package", "welcome").
					Str("emoji_id", emoji.ID).
					Str("emoji", w.config.Messages[idx].Emoji).
					Msg("Set EmojiID")

				w.config.Messages[idx].EmojiID = emoji.ID
			}
		}

		break
	}
}

func (w *Manager) hasValidConfigurationAgainstDiscordServer() bool {
	if w.guildID == "" {
		log.Error().
			Str("package", "welcome").
			Str("guild in config", w.guildName).
			Msg("Guild not found in Discord server")

		return false
	}

	if w.channelID == "" {
		log.Error().
			Str("package", "welcome").
			Str("channel in config", w.config.Channel).
			Msg("Channel not found in Discord server")

		return false
	}

	for idx, message := range w.config.Messages {
		if message.EmojiID == "" {
			log.Error().
				Str("package", "welcome").
				Int("message #", idx).
				Str("emoji in config", message.Emoji).
				Msg("Emoji not found in Discord server")

			return false
		}

		if message.RoleID == "" {
			log.Error().
				Str("package", "welcome").
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
	log.Info().
		Str("package", "welcome").
		Msg("Adding Handler on Message Reaction Add")

	w.discordSession.AddHandler(w.OnMessageReactionAdd)

	log.Info().
		Str("package", "welcome").
		Msg("Adding Handler on Message Reaction Remove")

	w.discordSession.AddHandler(w.OnMessageReactionRemove)

	log.Info().
		Str("package", "welcome").
		Msg("Adding messages to channel")

	err := w.addMessagesToChannel()
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Msg("Could not add messages to channel")

		return err
	}

	return nil
}

//nolint:funlen,cyclop
func (w *Manager) addMessagesToChannel() error {
	log.Info().
		Str("package", "welcome").
		Str("channel_id", w.channelID).
		Str("channel", w.channelName).
		Msg("Getting Messages from Channel")

	messages, err := w.discordSession.ChannelMessages(w.channelID, limitChannelMessages, "", "", "")
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("channel_id", w.channelID).
			Str("channel", w.channelName).
			Msg("Could not read Messages from Channel")

		return fmt.Errorf("%w", err)
	}

	var idxsMessageTreated []int

	for _, message := range messages {
		if message.Author.ID != w.discordSession.State.User.ID {
			log.Info().
				Str("package", "welcome").
				Str("message_id", message.ID).
				Str("channel_id", w.channelID).
				Str("channel", w.channelName).
				Msg("SKIP - Message in Channel is not from bot")

			continue
		}

		if len(message.Embeds) == 0 {
			log.Info().
				Str("package", "welcome").
				Str("message_id", message.ID).
				Str("channel_id", w.channelID).
				Str("channel", w.channelName).
				Msg("SKIP - Message in Channel is not an embed")

			continue
		}

		for idxMessages := range w.config.Messages {
			if w.isSameMessageAgainstConfig(message.Embeds[0], w.config.Messages[idxMessages]) {
				w.config.Messages[idxMessages].ID = message.ID

				log.Info().
					Str("package", "welcome").
					Str("message_title", w.config.Messages[idxMessages].Title).
					Msg("Message already sent -> update roles")

				err := w.updateRoleBelongMessage(w.config.Messages[idxMessages])
				if err != nil {
					log.Error().Err(err).
						Str("package", "welcome").
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
		messageTreated := slices.Contains(idxsMessageTreated, idxMessages)

		if !messageTreated {
			log.Info().
				Str("package", "welcome").
				Str("message_title", w.config.Messages[idxMessages].Title).
				Msg("Message missing - add Message")

			err := w.addMessage(&w.config.Messages[idxMessages])
			if err != nil {
				log.Error().Err(err).
					Str("package", "welcome").
					Str("message_title", w.config.Messages[idxMessages].Title).
					Msg("Could not add Message")

				return err
			}

			idxsMessageTreated = append(idxsMessageTreated, idxMessages)
		}
	}

	return nil
}

func (w *Manager) isSameMessageAgainstConfig(messageFromDiscord *discordgo.MessageEmbed, messageFromConfig Message) bool {
	return messageFromDiscord.Title == messageFromConfig.Title &&
		messageFromDiscord.Description == messageFromConfig.Description &&
		messageFromDiscord.Color == messageFromConfig.Color
}

//nolint:funlen,cyclop
func (w *Manager) updateRoleBelongMessage(message Message) error {
	log.Info().
		Str("package", "welcome").
		Str("message_id", message.ID).
		Str("message_title", message.Title).
		Str("channel_id", w.channelID).
		Str("channel", w.channelName).
		Str("emoji", message.Emoji+":"+message.EmojiID).
		Msg("Getting all Reactions from Message")

	users, err := helpers.MessageReactionsAll(w.discordSession, w.channelID, message.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("message_id", message.ID).
			Str("channel_id", w.channelID).
			Str("channel", w.channelName).
			Str("emoji", message.Emoji+":"+message.EmojiID).
			Msg("Could not get all Reactions")

		return fmt.Errorf("%w", err)
	}

	membersNotInGuild := []string{}

	for _, user := range users {
		if w.isUserBot(user.ID) {
			continue
		}

		member, err := w.discordSession.State.Member(w.guildID, user.ID)
		if err != nil {
			log.Error().Err(err).
				Str("package", "welcome").
				Str("user_id", user.ID).
				Str("guild_id", w.guildID).
				Msg("Could not find Member in Guild")

			membersNotInGuild = append(membersNotInGuild, user.ID)

			continue
		}

		skipUser := slices.Contains(member.Roles, message.RoleID)

		if skipUser {
			continue
		}

		log.Info().
			Str("package", "welcome").
			Str("role_id", message.RoleID).
			Str("role", message.Role).
			Str("user_id", user.ID).
			Str("username", user.Username).
			Msg("Adding Role to User")

		err = w.discordSession.GuildMemberRoleAdd(w.guildID, user.ID, message.RoleID)
		if err != nil {
			log.Error().Err(err).
				Str("package", "welcome").
				Str("role_id", message.RoleID).
				Str("role", message.Role).
				Str("user_id", user.ID).
				Str("username", user.Username).
				Msg("discord_bot.welcome.user_role_adding_failed")

			return fmt.Errorf("%w", err)
		}
	}

	if len(membersNotInGuild) > 0 {
		log.Info().
			Str("package", "welcome").
			Int("count_members_reacted", len(users)).
			Int("count_members_not_found", len(membersNotInGuild)).
			Msg("Members not found in Guild")

		if message.CanPurgeReactions &&
			len(users) >= message.PurgeThresholdMembersReacted &&
			len(membersNotInGuild) <= message.PurgeBelowCountMembersNotInGuild {
			log.Info().
				Str("package", "welcome").
				Msg("Do purge")

			for idx := range membersNotInGuild {
				log.Info().
					Str("package", "welcome").
					Str("message_id", message.ID).
					Str("emoji", message.Emoji+":"+message.EmojiID).
					Str("user_id", membersNotInGuild[idx]).
					Msg("Removing Reaction on Message for User")

				err = w.discordSession.MessageReactionRemove(w.channelID, message.ID, message.Emoji+":"+message.EmojiID, membersNotInGuild[idx])
				if err != nil {
					log.Error().Err(err).
						Str("package", "welcome").
						Str("message_id", message.ID).
						Str("emoji", message.Emoji+":"+message.EmojiID).
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
		Str("package", "welcome").
		Str("message_title", message.Title).
		Str("channel_id", w.channelID).
		Str("channel", w.channelName).
		Msg("Sending Message to Channel")

	messageSent, err := w.discordSession.ChannelMessageSendEmbed(w.channelID, &discordgo.MessageEmbed{
		Title:       message.Title,
		Description: message.Description,
		Color:       message.Color,
	})
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("message_title", message.Title).
			Str("channel_id", w.channelID).
			Str("channel", w.channelName).
			Msg("Could not send Message to Channel")

		return fmt.Errorf("%w", err)
	}

	log.Info().
		Str("package", "welcome").
		Str("message_id", messageSent.ID).
		Str("channel_id", w.channelID).
		Str("channel", w.config.Channel).
		Msg("Message Sent")

	message.ID = messageSent.ID

	log.Info().
		Str("package", "welcome").
		Str("message_id", messageSent.ID).
		Str("message_title", message.Title).
		Str("emoji", message.Emoji+":"+message.EmojiID).
		Msg("Adding Reaction to Message")

	err = w.discordSession.MessageReactionAdd(w.channelID, message.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("message_id", message.ID).
			Str("emoji", message.Emoji+":"+message.EmojiID).
			Msg("Could not add Reaction to Message")

		return fmt.Errorf("%w", err)
	}

	return nil
}

// OnMessageReactionAdd is public for tests, never call it directly
//
//nolint:dupl
func (w *Manager) OnMessageReactionAdd(_ *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	log.Debug().
		Str("package", "welcome").
		Msg("discord_bot.welcome.event_message_reaction_add_received")

	if reaction == nil || reaction.MessageReaction == nil {
		return
	}

	idxMessageFound, found := w.isMessageReactionMatching(reaction.MessageReaction)
	if !found {
		return
	}

	log.Info().
		Str("package", "welcome").
		Str("role_id", w.config.Messages[idxMessageFound].RoleID).
		Str("role", w.config.Messages[idxMessageFound].Role).
		Str("channel_id", reaction.ChannelID).
		Str("message_id", reaction.MessageID).
		Str("user_id", reaction.UserID).
		Msg("discord_bot.welcome.user_role_adding")

	err := w.discordSession.GuildMemberRoleAdd(w.guildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("role_id", w.config.Messages[idxMessageFound].RoleID).
			Str("role", w.config.Messages[idxMessageFound].Role).
			Str("channel_id", reaction.ChannelID).
			Str("message_id", reaction.MessageID).
			Str("user_id", reaction.UserID).
			Msg("discord_bot.welcome.user_role_adding_failed")

		return
	}

	log.Info().
		Str("package", "welcome").
		Str("role_id", w.config.Messages[idxMessageFound].RoleID).
		Str("role", w.config.Messages[idxMessageFound].Role).
		Str("channel_id", reaction.ChannelID).
		Str("message_id", reaction.MessageID).
		Str("user_id", reaction.UserID).
		Msg("discord_bot.welcome.user_role_added")
}

// OnMessageReactionRemove is public for tests, never call it directly
//
//nolint:dupl
func (w *Manager) OnMessageReactionRemove(_ *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	log.Debug().
		Str("package", "welcome").
		Msg("discord_bot.welcome.event_message_reaction_remove_received")

	if reaction == nil || reaction.MessageReaction == nil {
		return
	}

	idxMessageFound, found := w.isMessageReactionMatching(reaction.MessageReaction)
	if !found {
		return
	}

	log.Info().
		Str("package", "welcome").
		Str("role_id", w.config.Messages[idxMessageFound].RoleID).
		Str("role", w.config.Messages[idxMessageFound].Role).
		Str("channel_id", reaction.ChannelID).
		Str("message_id", reaction.MessageID).
		Str("user_id", reaction.UserID).
		Msg("discord_bot.welcome.user_role_removing")

	err := w.discordSession.GuildMemberRoleRemove(w.guildID, reaction.UserID, w.config.Messages[idxMessageFound].RoleID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("role_id", w.config.Messages[idxMessageFound].RoleID).
			Str("role", w.config.Messages[idxMessageFound].Role).
			Str("channel_id", reaction.ChannelID).
			Str("message_id", reaction.MessageID).
			Str("user_id", reaction.UserID).
			Msg("discord_bot.welcome.user_role_removing_failed")
		
		return
	}

	log.Info().
		Str("package", "welcome").
		Str("role_id", w.config.Messages[idxMessageFound].RoleID).
		Str("role", w.config.Messages[idxMessageFound].Role).
		Str("channel_id", reaction.ChannelID).
		Str("message_id", reaction.MessageID).
		Str("user_id", reaction.UserID).
		Msg("discord_bot.welcome.user_role_removed")
}

func (w *Manager) isMessageReactionMatching(messageReaction *discordgo.MessageReaction) (int, bool) {
	if messageReaction.ChannelID != w.channelID {
		return -1, false
	}

	if w.isUserBot(messageReaction.UserID) {
		return -1, false
	}

	idxMessageFound := -1

	for idxMessage := range w.config.Messages {
		if messageReaction.MessageID == w.config.Messages[idxMessage].ID {
			idxMessageFound = idxMessage

			break
		}
	}

	if idxMessageFound == -1 {
		return -1, false
	}

	if messageReaction.Emoji.Name != w.config.Messages[idxMessageFound].Emoji {
		return -1, false
	}
	
	return idxMessageFound, true
}

func (w *Manager) isUserBot(userID string) bool {
	return w.discordSession.State.User.ID == userID
}
