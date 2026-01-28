package exporter

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	limitChannelMessages           = 100
	limitLoopFetchChannelMessages  = 100
	permissionDirectory            = 0o750
	sleepBetweenAttachmentDownload = time.Second * 1
	sleepBetweenMessagesFetched    = time.Second * 5
)

// Run create sqlite database and export channels/messages/authors and attachments.
func (m *Manager) Run() {
	ctx := context.Background()

	log.Info().
		Msg("discord_bot.exporter.starting")

	hasChannelsIncluded := len(m.channelsIncluded) > 0

	for _, guild := range m.discordSession.State.Guilds {
		if guild.Icon != "" {
			m.downloadFile(guild.IconURL("4096"), path.Join(m.outputPath, "icon_guild_"+guild.ID+".png"))
		}

		saved := m.addOrUpdateGuild(ctx, translateGuild(guild))
		if !saved {
			continue
		}

		for idxChannel := range guild.Channels {
			if slices.Contains(m.channelsExcluded, guild.Channels[idxChannel].Name) {
				log.Info().
					Str("channel", guild.Channels[idxChannel].Name).
					Str("rule", "channels_excluded").
					Msg("discord_bot.exporter.skip_channel")

				continue
			}

			if hasChannelsIncluded && !slices.Contains(m.channelsIncluded, guild.Channels[idxChannel].Name) {
				log.Info().
					Str("channel", guild.Channels[idxChannel].Name).
					Str("rule", "channels_included").
					Msg("discord_bot.exporter.skip_channel")

				continue
			}

			saved := m.addOrUpdateChannel(ctx, translateChannel(guild.Channels[idxChannel]))
			if !saved {
				continue
			}

			m.fetchMessagesFromChannel(ctx, guild.ID, guild.Channels[idxChannel].ID, "")
		}
	}

	log.Info().
		Msg("discord_bot.exporter.stopped")
}

func (m *Manager) fetchMessagesFromChannel(ctx context.Context, guildID string, channelID string, startingID string) {
	maxLoops := 0

	for {
		maxLoops++

		messages, err := m.discordSession.ChannelMessages(channelID, limitChannelMessages, startingID, "", "")
		if err != nil {
			log.Error().Err(err).
				Str("channel_id", channelID).
				Msg("discord_bot.exporter.channel_messages_fetching_failed")

			return
		}

		if len(messages) == limitChannelMessages {
			startingID = messages[limitChannelMessages-1].ID
		} else {
			maxLoops += limitLoopFetchChannelMessages
		}

		for idxMessage := range messages {
			if messages[idxMessage].Author.Avatar != "" {
				m.downloadFile(messages[idxMessage].Author.AvatarURL(""), path.Join(m.outputPathUsers, messages[idxMessage].Author.Avatar+".png"))
			}

			m.addOrUpdateUser(ctx, translateUser(messages[idxMessage].Author))
			m.addOrUpdateMessage(ctx, translateMessage(messages[idxMessage], guildID))

			for idxAttachment := range messages[idxMessage].Attachments {
				//nolint:lll
				m.downloadFile(messages[idxMessage].Attachments[idxAttachment].URL, path.Join(m.outputPathAttachments, messages[idxMessage].Attachments[idxAttachment].ID+"_"+messages[idxMessage].Attachments[idxAttachment].Filename))

				log.Info().
					Msg("Sleep 1 second - attachment #" + strconv.Itoa(idxAttachment))

				time.Sleep(sleepBetweenAttachmentDownload)
			}
		}

		log.Info().
			Msg("Sleep 5 second - messages")

		time.Sleep(sleepBetweenMessagesFetched)

		if maxLoops >= limitLoopFetchChannelMessages {
			return
		}
	}
}

func (m *Manager) downloadFile(url string, filepath string) {
	_, ok := m.files[url]
	if ok {
		return
	}

	m.files[url] = struct{}{}

	log.Info().
		Str("URL", url).
		Str("filepath", filepath).
		Msg("discord_bot.exporter.downloading_file")

	//nolint:gosec,noctx
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).
			Str("URL", url).
			Str("filepath", filepath).
			Msg("discord_bot.exporter.file_downloading_failed")

		return
	}

	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Err(err).
			Str("URL", url).
			Str("filepath", filepath).
			Int("status_code", resp.StatusCode).
			Msg("discord_bot.exporter.file_downloading_failed")

		return
	}

	//nolint:gosec
	out, err := os.Create(filepath)
	if err != nil {
		log.Error().Err(err).
			Str("URL", url).
			Str("filepath", filepath).
			Str("reason", "failed to create file on disk").
			Msg("discord_bot.exporter.file_downloading_failed")

		return
	}

	//nolint:errcheck
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Error().Err(err).
			Str("URL", url).
			Str("filepath", filepath).
			Str("reason", "failed to copy file on disk").
			Msg("discord_bot.exporter.file_downloading_failed")
	}

	log.Info().
		Str("URL", url).
		Str("filepath", filepath).
		Msg("discord_bot.exporter.file_downloaded")
}
