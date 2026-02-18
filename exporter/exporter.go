// Package exporter defines configuration struct and how to export discord messages.
package exporter

import (
	"database/sql"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const (
	modeOnce string = "once"
)

// Configuration contains exporter parameters.
type Configuration struct {
	Mode             string   `json:"mode"`
	OutputPath       string   `json:"output_path"`
	DatabaseFilename string   `json:"database_filename"`
	ChannelsIncluded []string `json:"channels_included"`
	ChannelsExcluded []string `json:"channels_excluded"`
}

// Manager is a struct.
type Manager struct {
	files                 map[string]struct{}
	db                    *sql.DB
	discordSession        *discordgo.Session
	guildName             string
	mode                  string
	outputPath            string
	outputPathAttachments string
	outputPathUsers       string
	databaseFilename      string
	channelsIncluded      []string
	channelsExcluded      []string
}

// NewExporterManager checks configuration and returns a manager.
func NewExporterManager(
	config Configuration,
	guildName string,
	discordSession *discordgo.Session,
) *Manager {
	manager := &Manager{
		discordSession: discordSession,
		guildName:      guildName,
		files:          map[string]struct{}{},
	}

	log.Info().
		Msg("discord_bot.exporter.validating_configuration")

	if !manager.hasValidConfigurationInFile(config) {
		log.Error().
			Msg("discord_bot.exporter.configuration_validation_failed")

		return nil
	}

	log.Info().
		Msg("discord_bot.exporter.configuration_validated")

	log.Info().
		Msg("discord_bot.exporter.creating_folders")

	created := createFolders(manager.outputPathAttachments, manager.outputPathUsers)
	if !created {
		log.Error().
			Msg("discord_bot.exporter.folders_creating_failed")

		return nil
	}

	log.Info().
		Msg("discord_bot.exporter.folders_created")

	log.Info().
		Msg("discord_bot.exporter.initializing_database")

	db := initializeDatabase(manager.outputPath, manager.databaseFilename)
	if db == nil {
		log.Error().
			Msg("discord_bot.exporter.database_initialized_failed")

		return nil
	}

	log.Info().
		Msg("discord_bot.exporter.database_initialized")

	manager.db = db

	return manager
}

//nolint:funlen
func (m *Manager) hasValidConfigurationInFile(config Configuration) bool {
	if !slices.Contains([]string{modeOnce}, config.Mode) {
		log.Error().
			Str("help", "Accepted values are 'once'").
			Msg("discord_bot.exporter.configuration_invalid_mode")

		return false
	}

	m.mode = config.Mode

	log.Info().
		Str("mode", m.mode).
		Msg("discord_bot.exporter.set_mode")

	outputPath := strings.TrimSpace(config.OutputPath)
	if outputPath == "" {
		outputPath = "./exports"

		log.Info().
			Str("help", "output_path is empty, use default './exports'").
			Msg("discord_bot.exporter.use_default_output_path")
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		log.Error().Err(err).
			Str("output_path", outputPath).
			Msg("discord_bot.exporter.configuration_invalid_output_path")

		return false
	}

	m.outputPath = absPath

	log.Info().
		Str("output_path", m.outputPath).
		Msg("discord_bot.exporter.set_output_path")

	m.outputPathAttachments = path.Join(m.outputPath, "attachments")

	log.Info().
		Str("output_path_attachments", m.outputPathAttachments).
		Msg("discord_bot.exporter.set_output_path_attachments")

	m.outputPathUsers = path.Join(m.outputPath, "users")

	log.Info().
		Str("output_path_users", m.outputPathUsers).
		Msg("discord_bot.exporter.set_output_path_users")

	m.databaseFilename = strings.TrimSpace(config.DatabaseFilename)
	if m.databaseFilename == "" {
		log.Info().
			Str("help", "database_filename is empty, use default 'discord.db'").
			Msg("discord_bot.exporter.use_default_database_filename")

		m.databaseFilename = "discord.db"
	} else {
		log.Info().
			Str("database_filename", m.databaseFilename).
			Msg("discord_bot.exporter.set_database_filename")
	}

	channelSeen := make(map[string]struct{})

	for idx := range config.ChannelsExcluded {
		channel := strings.TrimSpace(config.ChannelsExcluded[idx])
		if channel == "" {
			continue
		}

		m.channelsExcluded = append(m.channelsExcluded, channel)

		channelSeen[channel] = struct{}{}
	}

	log.Info().
		Strs("channels_excluded", m.channelsExcluded).
		Msg("discord_bot.exporter.set_channels_excluded")

	for idx := range config.ChannelsIncluded {
		channel := strings.TrimSpace(config.ChannelsIncluded[idx])
		if channel == "" {
			continue
		}

		_, exists := channelSeen[channel]
		if exists {
			log.Error().
				Str("channel", channel).
				Msg("discord_bot.exporter.collision_channel")

			return false
		}

		m.channelsIncluded = append(m.channelsIncluded, channel)
	}

	log.Info().
		Strs("channels_included", m.channelsIncluded).
		Msg("discord_bot.exporter.set_channels_included")

	return true
}

func createFolders(outputPathAttachments string, outputPathUsers string) bool {
	log.Info().
		Str("output_attachments_path", outputPathAttachments).
		Uint("permission", permissionDirectory).
		Msg("discord_bot.exporter.creating_output_attachments_folder")

	err := os.MkdirAll(outputPathAttachments, permissionDirectory)
	if err != nil {
		log.Error().Err(err).
			Str("output_attachments_path", outputPathAttachments).
			Uint("permission", permissionDirectory).
			Msg("discord_bot.exporter.output_attachments_folder_creation_failed")

		return false
	}

	log.Info().
		Str("output_attachments_path", outputPathAttachments).
		Uint("permission", permissionDirectory).
		Msg("discord_bot.exporter.output_attachments_folder_created")

	log.Info().
		Str("output_users_path", outputPathUsers).
		Uint("permission", permissionDirectory).
		Msg("discord_bot.exporter.creating_output_users_folder")

	err = os.MkdirAll(outputPathUsers, permissionDirectory)
	if err != nil {
		log.Error().Err(err).
			Str("output_users_path", outputPathUsers).
			Uint("permission", permissionDirectory).
			Msg("discord_bot.exporter.output_users_folder_creation_failed")

		return false
	}

	log.Info().
		Str("output_users_path", outputPathUsers).
		Uint("permission", permissionDirectory).
		Msg("discord_bot.exporter.output_users_folder_created")

	return true
}
