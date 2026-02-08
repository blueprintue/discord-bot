package welcome

import (
	"fmt"
	"slices"

	"github.com/blueprintue/discord-bot/helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

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

	err = w.discordSession.MessageReactionAdd(w.channelID, messageSent.ID, message.Emoji+":"+message.EmojiID)
	if err != nil {
		log.Error().Err(err).
			Str("package", "welcome").
			Str("message_id", messageSent.ID).
			Str("emoji", message.Emoji+":"+message.EmojiID).
			Msg("discord_bot.welcome.reaction_adding_failed")

		return "", fmt.Errorf("%w", err)
	}

	return messageSent.ID, nil
}
