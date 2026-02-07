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
	guildName      string
	guildID        string
	channelName    string
	channelID      string
	messages       []Message
}

// NewWelcomeManager return a Manager.
func NewWelcomeManager(
	config Configuration,
	guildName string,
	discordSession *discordgo.Session,
) *Manager {
	manager := &Manager{
		discordSession: discordSession,
		guildName:      guildName,
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

	manager.completeConfiguration(config)

	if !manager.hasValidConfigurationAgainstDiscordServer(config) {
		log.Error().
			Str("package", "welcome").
			Int("step", stepTwoConfiguration).
			Msg("discord_bot.welcome.configuration_validation_failed")

		return nil
	}

	log.Info().
		Str("package", "welcome").
		Msg("discord_bot.welcome.configuration_validated")

	return manager
}

func hasValidConfigurationInFile(config Configuration) bool {
	if config.Channel == "" {
		log.Error().
			Str("package", "welcome").
			Msg("discord_bot.welcome.configuration_empty_channel")

		return false
	}

	if len(config.Messages) == 0 {
		log.Error().
			Str("package", "welcome").
			Msg("discord_bot.welcome.configuration_empty_messages")

		return false
	}

	for idx, message := range config.Messages {
		if message.Title == "" && message.Description == "" {
			log.Error().
				Str("package", "welcome").
				Int("message index", idx).
				Str("title", message.Title).
				Str("description", message.Description).
				Msg("discord_bot.welcome.configuration_empty_title_description_message")

			return false
		}

		if message.Emoji == "" {
			log.Error().
				Str("package", "welcome").
				Int("message index", idx).
				Msg("discord_bot.welcome.configuration_empty_emoji_message")

			return false
		}

		if message.Role == "" {
			log.Error().
				Str("package", "welcome").
				Int("message index", idx).
				Msg("discord_bot.welcome.configuration_empty_role_message")

			return false
		}
	}

	return true
}

//nolint:cyclop,funlen
func (w *Manager) completeConfiguration(config Configuration) {
	w.messages = append(make([]Message, 0, len(config.Messages)), config.Messages...)

	for _, guild := range w.discordSession.State.Guilds {
		if guild.Name != w.guildName {
			continue
		}

		log.Info().
			Str("package", "welcome").
			Str("guild_id", guild.ID).
			Str("guild", w.guildName).
			Msg("discord_bot.welcome.set_guild_id")

		w.guildID = guild.ID

		for _, channel := range guild.Channels {
			if channel.Name != config.Channel {
				continue
			}

			log.Info().
				Str("package", "welcome").
				Str("channel_id", channel.ID).
				Str("channel", config.Channel).
				Msg("discord_bot.welcome.set_channel_id")

			w.channelID = channel.ID
			w.channelName = config.Channel

			break
		}

		for _, role := range guild.Roles {
			for idx := range w.messages {
				if w.messages[idx].Role != role.Name {
					continue
				}

				log.Info().
					Str("package", "welcome").
					Int("message index", idx).
					Str("role_id", role.ID).
					Str("role", w.messages[idx].Role).
					Msg("discord_bot.welcome.set_role_id")

				w.messages[idx].RoleID = role.ID
			}
		}

		for _, emoji := range guild.Emojis {
			emojiRichEmbed := fmt.Sprintf("<:%s:%s>", emoji.Name, emoji.ID)
			emojiInText := ":" + emoji.Name + ":"

			for idx := range w.messages {
				w.messages[idx].Title = strings.ReplaceAll(w.messages[idx].Title, emojiInText, emojiRichEmbed)
				w.messages[idx].Description = strings.ReplaceAll(w.messages[idx].Description, emojiInText, emojiRichEmbed)

				if w.messages[idx].Emoji != emoji.Name {
					continue
				}

				log.Info().
					Str("package", "welcome").
					Int("message index", idx).
					Str("emoji_id", emoji.ID).
					Str("emoji", w.messages[idx].Emoji).
					Msg("discord_bot.welcome.set_emoji_id")

				w.messages[idx].EmojiID = emoji.ID
			}
		}

		break
	}
}

func (w *Manager) hasValidConfigurationAgainstDiscordServer(config Configuration) bool {
	if w.guildID == "" {
		log.Error().
			Str("package", "welcome").
			Str("guild", w.guildName).
			Msg("discord_bot.welcome.configuration_guild_missed")

		return false
	}

	if w.channelID == "" {
		log.Error().
			Str("package", "welcome").
			Str("channel", config.Channel).
			Msg("discord_bot.welcome.configuration_channel_missed")

		return false
	}

	for idx, message := range w.messages {
		if message.EmojiID == "" {
			log.Error().
				Str("package", "welcome").
				Int("message index", idx).
				Str("emoji", message.Emoji).
				Msg("discord_bot.welcome.configuration_emoji_missed")

			return false
		}

		if message.RoleID == "" {
			log.Error().
				Str("package", "welcome").
				Int("message index", idx).
				Str("role", message.Role).
				Msg("discord_bot.welcome.configuration_role_missed")

			return false
		}
	}

	return true
}

// Run do the main task of Welcome.
func (w *Manager) Run() error {
	log.Info().
		Str("package", "welcome").
		Msg("discord_bot.welcome.add_handler_on_message_reaction_add")

	w.discordSession.AddHandler(w.OnMessageReactionAdd)

	log.Info().
		Str("package", "welcome").
		Msg("discord_bot.welcome.add_handler_on_message_reaction_remove")

	w.discordSession.AddHandler(w.OnMessageReactionRemove)

	log.Info().
		Str("package", "welcome").
		Msg("discord_bot.welcome.adding_messages")

	err := w.addMessagesToChannel()
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Msg("discord_bot.welcome.messages_adding_failed")

		return err
	}
	
	log.Info().
		Str("package", "welcome").
		Msg("discord_bot.welcome.messages_added")

	return nil
}

//nolint:funlen,cyclop
func (w *Manager) addMessagesToChannel() error {
	log.Info().
		Str("package", "welcome").
		Str("channel_id", w.channelID).
		Str("channel", w.channelName).
		Msg("discord_bot.welcome.fetching_messages")

	messages, err := w.discordSession.ChannelMessages(w.channelID, limitChannelMessages, "", "", "")
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("channel_id", w.channelID).
			Str("channel", w.channelName).
			Msg("discord_bot.welcome.messages_fetching_failed")

		return fmt.Errorf("%w", err)
	}

	log.Info().
		Str("package", "welcome").
		Str("channel_id", w.channelID).
		Str("channel", w.channelName).
		Msg("discord_bot.welcome.messages_fetched")

	var idxsMessageTreated []int

	for _, message := range messages {
		if message.Author.ID != w.discordSession.State.User.ID {
			continue
		}

		if len(message.Embeds) == 0 {
			continue
		}

		for idxMessage := range w.messages {
			if w.isSameMessageAgainstConfig(message.Embeds[0], w.messages[idxMessage]) {
				w.messages[idxMessage].ID = message.ID

				err := w.updateUserRoleBelongMessage(w.messages[idxMessage])
				if err != nil {
					log.Error().Err(err).
						Str("package", "welcome").
						Str("message_title", w.messages[idxMessage].Title).
						Msg("discord_bot.welcome.role_updating_failed")

					return err
				}

				idxsMessageTreated = append(idxsMessageTreated, idxMessage)

				break
			}
		}
	}

	for idxMessage := range w.messages {
		messageTreated := slices.Contains(idxsMessageTreated, idxMessage)

		if !messageTreated {
			log.Info().
				Str("package", "welcome").
				Str("message_title", w.messages[idxMessage].Title).
				Msg("discord_bot.welcome.adding_missed_messages")

			messageID, err := w.addMessage(w.messages[idxMessage])
			if err != nil {
				log.Error().Err(err).
					Str("package", "welcome").
					Str("message_title", w.messages[idxMessage].Title).
					Msg("discord_bot.welcome.messages_missed_adding_failed")

				return err
			}

			w.messages[idxMessage].ID = messageID

			log.Info().
				Str("package", "welcome").
				Str("message_title", w.messages[idxMessage].Title).
				Msg("discord_bot.welcome.missed_messages_added")

			idxsMessageTreated = append(idxsMessageTreated, idxMessage)
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
func (w *Manager) updateUserRoleBelongMessage(message Message) error {
	log.Info().
		Str("package", "welcome").
		Str("message_id", message.ID).
		Str("message_title", message.Title).
		Str("channel_id", w.channelID).
		Str("channel", w.channelName).
		Str("emoji", message.Emoji+":"+message.EmojiID).
		Msg("discord_bot.welcome.fetching_reactions_message")

	users, err := helpers.MessageReactionsAll(w.discordSession, w.channelID, message.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("message_id", message.ID).
			Str("channel_id", w.channelID).
			Str("channel", w.channelName).
			Str("emoji", message.Emoji+":"+message.EmojiID).
			Msg("discord_bot.welcome.reactions_message_fetching_failed")

		return fmt.Errorf("%w", err)
	}

	membersNotInGuild := []string{}

	for _, user := range users {
		if w.isUserBot(user.ID) {
			continue
		}

		member, err := w.discordSession.State.Member(w.guildID, user.ID)
		if err != nil {
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
			Msg("discord_bot.welcome.adding_user_role_adding")

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
			Msg("discord_bot.welcome.members_not_in_guild")

		if message.CanPurgeReactions &&
			len(users) >= message.PurgeThresholdMembersReacted &&
			len(membersNotInGuild) <= message.PurgeBelowCountMembersNotInGuild {
			log.Info().
				Str("package", "welcome").
				Msg("discord_bot.welcome.purge_reactions")

			for idx := range membersNotInGuild {
				log.Info().
					Str("package", "welcome").
					Str("message_id", message.ID).
					Str("emoji", message.Emoji+":"+message.EmojiID).
					Str("user_id", membersNotInGuild[idx]).
					Msg("discord_bot.welcome.removing_reaction")

				err = w.discordSession.MessageReactionRemove(w.channelID, message.ID, message.Emoji+":"+message.EmojiID, membersNotInGuild[idx])
				if err != nil {
					log.Error().Err(err).
						Str("package", "welcome").
						Str("message_id", message.ID).
						Str("emoji", message.Emoji+":"+message.EmojiID).
						Str("user_id", membersNotInGuild[idx]).
						Msg("discord_bot.welcome.reaction_removing_failed")
				}
			}
		}
	}

	return nil
}

func (w *Manager) addMessage(message Message) (string, error) {
	log.Info().
		Str("package", "welcome").
		Str("message_title", message.Title).
		Str("channel_id", w.channelID).
		Str("channel", w.channelName).
		Msg("discord_bot.welcome.adding_message")

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
			Msg("discord_bot.welcome.message_adding_failed")

		return "", fmt.Errorf("%w", err)
	}

	log.Info().
		Str("package", "welcome").
		Str("message_id", messageSent.ID).
		Str("channel_id", w.channelID).
		Str("channel", w.channelName).
		Msg("discord_bot.welcome.message_added")

	log.Info().
		Str("package", "welcome").
		Str("message_id", messageSent.ID).
		Str("message_title", message.Title).
		Str("emoji", message.Emoji+":"+message.EmojiID).
		Msg("discord_bot.welcome.adding_reaction")

	err = w.discordSession.MessageReactionAdd(w.channelID, message.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("message_id", message.ID).
			Str("emoji", message.Emoji+":"+message.EmojiID).
			Msg("discord_bot.welcome.reaction_adding_failed")

		return messageSent.ID, fmt.Errorf("%w", err)
	}

	return messageSent.ID, nil
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
		Str("role_id", w.messages[idxMessageFound].RoleID).
		Str("role", w.messages[idxMessageFound].Role).
		Str("channel_id", reaction.ChannelID).
		Str("message_id", reaction.MessageID).
		Str("user_id", reaction.UserID).
		Msg("discord_bot.welcome.user_role_adding")

	err := w.discordSession.GuildMemberRoleAdd(w.guildID, reaction.UserID, w.messages[idxMessageFound].RoleID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("role_id", w.messages[idxMessageFound].RoleID).
			Str("role", w.messages[idxMessageFound].Role).
			Str("channel_id", reaction.ChannelID).
			Str("message_id", reaction.MessageID).
			Str("user_id", reaction.UserID).
			Msg("discord_bot.welcome.user_role_adding_failed")

		return
	}

	log.Info().
		Str("package", "welcome").
		Str("role_id", w.messages[idxMessageFound].RoleID).
		Str("role", w.messages[idxMessageFound].Role).
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
		Str("role_id", w.messages[idxMessageFound].RoleID).
		Str("role", w.messages[idxMessageFound].Role).
		Str("channel_id", reaction.ChannelID).
		Str("message_id", reaction.MessageID).
		Str("user_id", reaction.UserID).
		Msg("discord_bot.welcome.user_role_removing")

	err := w.discordSession.GuildMemberRoleRemove(w.guildID, reaction.UserID, w.messages[idxMessageFound].RoleID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("role_id", w.messages[idxMessageFound].RoleID).
			Str("role", w.messages[idxMessageFound].Role).
			Str("channel_id", reaction.ChannelID).
			Str("message_id", reaction.MessageID).
			Str("user_id", reaction.UserID).
			Msg("discord_bot.welcome.user_role_removing_failed")
		
		return
	}

	log.Info().
		Str("package", "welcome").
		Str("role_id", w.messages[idxMessageFound].RoleID).
		Str("role", w.messages[idxMessageFound].Role).
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

	for idxMessage := range w.messages {
		if messageReaction.MessageID == w.messages[idxMessage].ID {
			idxMessageFound = idxMessage

			break
		}
	}

	if idxMessageFound == -1 {
		return -1, false
	}

	if messageReaction.Emoji.Name != w.messages[idxMessageFound].Emoji {
		return -1, false
	}
	
	return idxMessageFound, true
}

func (w *Manager) isUserBot(userID string) bool {
	return w.discordSession.State.User.ID == userID
}
