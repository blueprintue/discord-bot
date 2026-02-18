//nolint:paralleltest
package exporter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/blueprintue/discord-bot/exporter"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

const guildName = "guild-name"

func SkipTestNewExporterManager(t *testing.T) {
	var bufferLogs bytes.Buffer

	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	exporterManager := exporter.NewExporterManager(exporter.Configuration{
		Mode: "once",
	}, guildName, session)
	require.NotNil(t, exporterManager)

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.validating_configuration"}`, parts[0])
	require.JSONEq(t, `{"level":"info","mode":"once","message":"discord_bot.exporter.set_mode"}`, parts[1])
	require.JSONEq(t, `{"level":"info","help":"output_path is empty, use default './exports'","message":"discord_bot.exporter.use_default_output_path"}`, parts[2])
	// require.JSONEq(t, `{"level":"info","output_path":"","message":"discord_bot.exporter.set_output_path"}`, parts[3])
	// require.JSONEq(t, `{"level":"info","output_path_attachments":"","message":"discord_bot.exporter.set_output_path_attachments"}`, parts[4])
	// require.JSONEq(t, `{"level":"info","output_path_users":"","message":"discord_bot.exporter.set_output_path_users"}`, parts[5])
	require.JSONEq(t, `{"level":"info","help":"database_filename is empty, use default 'discord.db'","message":"discord_bot.exporter.use_default_database_filename"}`, parts[6])
	require.JSONEq(t, `{"level":"info","channels_excluded":[],"message":"discord_bot.exporter.set_channels_excluded"}`, parts[7])
	require.JSONEq(t, `{"level":"info","channels_included":[],"message":"discord_bot.exporter.set_channels_included"}`, parts[8])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.configuration_validated"}`, parts[9])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.creating_folders"}`, parts[10])
	// require.JSONEq(t, `{"level":"info","output_attachments_path":"","permission":488,"message":"discord_bot.exporter.creating_output_attachments_folder"}`, parts[11])
	// require.JSONEq(t, `{"level":"info","output_attachments_path":"","permission":488,"message":"discord_bot.exporter.output_attachments_folder_created"}`, parts[12])
	// require.JSONEq(t, `{"level":"info","output_users_path":"","permission":488,"message":"discord_bot.exporter.creating_output_users_folder"}`, parts[13])
	// require.JSONEq(t, `{"level":"info","output_users_path":"","permission":488,"message":"discord_bot.exporter.output_users_folder_created"}`, parts[14])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.folders_created"}`, parts[15])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.initializing_database"}`, parts[16])
	// require.JSONEq(t, `{"level":"info","database":"","message":"discord_bot.exporter.creating_database"}`, parts[17])
	// require.JSONEq(t, `{"level":"info","database":"","message":"discord_bot.exporter.database_created"}`, parts[18])
	// require.JSONEq(t, `{"level":"info","database":"","message":"discord_bot.exporter.checking_database_version"}`, parts[19])
	// require.JSONEq(t, `{"level":"info","database":"","version":"3.51.1","message":"discord_bot.exporter.database_version_checked"}`, parts[20])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.creating_guilds_table"}`, parts[21])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.guilds_table_created"}`, parts[22])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.creating_users_table"}`, parts[23])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.users_table_created"}`, parts[24])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.creating_channels_table"}`, parts[25])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.channels_table_created"}`, parts[26])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.creating_messages_table"}`, parts[27])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.messages_table_created"}`, parts[28])
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.database_initialized"}`, parts[29])
	require.Empty(t, parts[30])
}

func TestNewExporterManager_ErrorInvalidMode(t *testing.T) {
	var bufferLogs bytes.Buffer

	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	exporterManager := exporter.NewExporterManager(exporter.Configuration{
		Mode: "-",
	}, guildName, session)
	require.Nil(t, exporterManager)

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.validating_configuration"}`, parts[0])
	require.JSONEq(t, `{"level":"error","help":"Accepted values are 'once'","message":"discord_bot.exporter.configuration_invalid_mode"}`, parts[1])
	require.JSONEq(t, `{"level":"error","message":"discord_bot.exporter.configuration_validation_failed"}`, parts[2])
	require.Empty(t, parts[3])
}

func TestNewExporterManager_ErrorCollisionChannels(t *testing.T) {
	var bufferLogs bytes.Buffer

	log.Logger = zerolog.New(&bufferLogs).Level(zerolog.TraceLevel).With().Logger()

	session, err := discordgo.New("fake-token")
	require.NoError(t, err)

	exporterManager := exporter.NewExporterManager(exporter.Configuration{
		Mode:             "once",
		ChannelsIncluded: []string{"foo"},
		ChannelsExcluded: []string{"foo"},
	}, guildName, session)
	require.Nil(t, exporterManager)

	parts := strings.Split(bufferLogs.String(), "\n")
	require.JSONEq(t, `{"level":"info","message":"discord_bot.exporter.validating_configuration"}`, parts[0])
	require.JSONEq(t, `{"level":"info","mode":"once","message":"discord_bot.exporter.set_mode"}`, parts[1])
	require.JSONEq(t, `{"level":"info","help":"output_path is empty, use default './exports'","message":"discord_bot.exporter.use_default_output_path"}`, parts[2])
	// require.JSONEq(t, `{"level":"info","output_path":"","message":"discord_bot.exporter.set_output_path"}`, parts[3])
	// require.JSONEq(t, `{"level":"info","output_path_attachments":"","message":"discord_bot.exporter.set_output_path_attachments"}`, parts[4])
	// require.JSONEq(t, `{"level":"info","output_path_users":"","message":"discord_bot.exporter.set_output_path_users"}`, parts[5])
	require.JSONEq(t, `{"level":"info","help":"database_filename is empty, use default 'discord.db'","message":"discord_bot.exporter.use_default_database_filename"}`, parts[6])
	require.JSONEq(t, `{"level":"info","channels_excluded":["foo"],"message":"discord_bot.exporter.set_channels_excluded"}`, parts[7])
	require.JSONEq(t, `{"level":"error","channel":"foo","message":"discord_bot.exporter.collision_channel"}`, parts[8])
	require.JSONEq(t, `{"level":"error","message":"discord_bot.exporter.configuration_validation_failed"}`, parts[9])
	require.Empty(t, parts[10])
}
