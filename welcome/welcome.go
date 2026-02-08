// Package welcome defines configuration struct, add messages, update roles.
package welcome

import (
	"fmt"
	"strings"

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
	CanPurgeReactions                bool   `json:"can_purge_reactions"`
	PurgeThresholdMembersReacted     int    `json:"purge_threshold_members_reacted"`
	PurgeBelowCountMembersNotInGuild int    `json:"purge_below_count_members_not_in_guild"`
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
