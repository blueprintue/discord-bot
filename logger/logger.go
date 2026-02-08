// Package logger create logs folder, defines log level, how to rotate logs and how to format logs.
package logger

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/blueprintue/discord-bot/configuration"
	"github.com/ilya1st/rotatewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	formatTimestampLogger = "2006-01-02T15:04:05.000000000Z07:00"
	permissionDirectory   = 0750
)

// Configure configures logger.
//
//nolint:funlen
func Configure(confLog configuration.Log) error {
	var err error

	logFile := path.Clean(confLog.Filename)

	log.Info().
		Str("path", path.Dir(logFile)).
		Uint("permission", permissionDirectory).
		Msg("discord_bot.logger.creating_log_folder")

	err = os.MkdirAll(path.Dir(logFile), permissionDirectory)
	if err != nil {
		log.Error().Err(err).
			Str("path", path.Dir(logFile)).
			Uint("permission", permissionDirectory).
			Msg("discord_bot.logger.log_folder_creation_failed")

		return fmt.Errorf("%w", err)
	}

	log.Info().
		Str("path", path.Dir(logFile)).
		Uint("permission", permissionDirectory).
		Msg("discord_bot.logger.log_folder_created")

	log.Info().
		Str("log_file", logFile).
		Int("number_files_rotation", confLog.NumberFilesRotation).
		Msg("discord_bot.logger.creating_log_rotate_writer")

	rwriter, err := rotatewriter.NewRotateWriter(logFile, confLog.NumberFilesRotation)
	if err != nil {
		log.Error().Err(err).
			Str("log_file", logFile).
			Int("number_files_rotation", confLog.NumberFilesRotation).
			Msg("discord_bot.logger.log_rotate_writer_creation_failed")

		return fmt.Errorf("%w", err)
	}

	log.Info().
		Str("log_file", logFile).
		Int("number_files_rotation", confLog.NumberFilesRotation).
		Msg("discord_bot.logger.log_rotate_writer_created")

	sighupChan := make(chan os.Signal, 1)

	signal.Notify(sighupChan, syscall.SIGHUP)

	go func() {
		for {
			_, ok := <-sighupChan
			if !ok {
				return
			}

			errRotate := rwriter.Rotate(nil)
			if errRotate != nil {
				log.Error().Err(errRotate).
					Msg("discord_bot.logger.log_rotate_failed")
			}
		}
	}()

	log.Info().
		Str("log_level", confLog.Level).
		Msg("discord_bot.logger.parsing_log_level")

	logLevel, err := zerolog.ParseLevel(confLog.Level)
	if err != nil {
		log.Error().Err(err).
			Str("log_level", confLog.Level).
			Msg("discord_bot.logger.log_level_parsing_failed")

		return fmt.Errorf("%w", err)
	}

	log.Info().
		Str("log_level", confLog.Level).
		Msg("discord_bot.logger.log_level_parsed")

	zerolog.SetGlobalLevel(logLevel)

	zerolog.TimeFieldFormat = formatTimestampLogger

	log.Logger = zerolog.New(
		zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC1123,
			}, rwriter),
	).With().Timestamp().Caller().Logger()

	return nil
}
