package logger

import (
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

// Configure configures logger
func Configure(configuration *configuration.Configuration) {
	var err error

	logFile := path.Clean(configuration.Log.Filename)
	if err := os.MkdirAll(path.Dir(logFile), os.ModePerm); err != nil {
		log.Fatal().Err(err).Msgf("Cannot create log folder")
	}
	rwriter, err := rotatewriter.NewRotateWriter(logFile, 5)
	if err != nil {
		log.Fatal().Err(err).Msgf("Cannot create log file writer")
	}
	sighupChan := make(chan os.Signal, 1)
	signal.Notify(sighupChan, syscall.SIGHUP)
	go func() {
		for {
			_, ok := <-sighupChan
			if !ok {
				return
			}
			if err := rwriter.Rotate(nil); err != nil {
				log.Error().Err(err).Msgf("Cannot rotate log")
			}
		}
	}()

	log.Logger = zerolog.New(zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC1123,
		}, rwriter)).With().Timestamp().Caller().Logger()

	logLevel, err := zerolog.ParseLevel(configuration.Log.Level)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unknown log level")
	} else {
		zerolog.SetGlobalLevel(logLevel)
	}
}
