package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"kube-proxless/internal/utils"
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
	v := utils.ConvertStringParamsToInterface(params)
	log.Error().Err(err).Msgf(msg, v...)
}

func Panicf(err error, msg string, params ...string) {
	v := utils.ConvertStringParamsToInterface(params)
	log.Panic().Err(err).Msgf(msg, v...)
}

func Debugf(msg string, params ...string) {
	v := utils.ConvertStringParamsToInterface(params)
	log.Debug().Msgf(msg, v...)
}

func Infof(msg string, params ...string) {
	v := utils.ConvertStringParamsToInterface(params)
	log.Info().Msgf(msg, v...)
}
