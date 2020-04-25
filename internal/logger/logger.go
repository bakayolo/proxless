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

	log.Logger = log.With().Caller().Logger()
	log.Info().Msgf("logger initialized with %s settings", zerolog.GlobalLevel())
}

func Errorf(err error, msg string, params ...string) {
	v := convertStringParamsToInterface(params)
	log.Error().Err(err).Msgf(msg, v...)
}

func Panicf(err error, msg string, params ...string) {
	v := convertStringParamsToInterface(params)
	log.Panic().Err(err).Msgf(msg, v...)
}

func Debugf(msg string, params ...string) {
	v := convertStringParamsToInterface(params)
	log.Debug().Msgf(msg, v...)
}

func Infof(msg string, params ...string) {
	v := convertStringParamsToInterface(params)
	log.Info().Msgf(msg, v...)
}

func convertStringParamsToInterface(params []string) []interface{} {
	v := make([]interface{}, len(params))
	for i, p := range params {
		v[i] = p
	}
	return v
}
