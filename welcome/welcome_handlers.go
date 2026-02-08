package welcome

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

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
