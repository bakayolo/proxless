package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"strings"
)

func InitLogger() {
	logLevel := os.Getenv("LOG_LEVEL")

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	switch strings.ToUpper(logLevel) {
	case "VERBOSE":
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	case "DEBUG":
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "ERROR", "FATAL", "PANIC":
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// log.Logger = log.With().Caller().Logger()
	log.Info().Msgf("logger initialized with %s settings", zerolog.GlobalLevel())
}

func Errorf(err error, msg string, params ...interface{}) {
	log.Error().Err(err).Msgf(msg, params...)
}

func Panicf(err error, msg string, params ...interface{}) {
	log.Panic().Err(err).Msgf(msg, params...)
}

func Fatalf(err error, msg string, params ...interface{}) {
	log.Fatal().Err(err).Msgf(msg, params...)
}

func Debugf(msg string, params ...interface{}) {
	log.Debug().Msgf(msg, params...)
}

func Infof(msg string, params ...interface{}) {
	log.Info().Msgf(msg, params...)
}
