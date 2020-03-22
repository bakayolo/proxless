package config

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
)

var (
	KubeConfigPath        string
	LogLevel              string
	Port                  string
	MaxConsPerHost        int
	Namespace             string
	ServerlessTTL         int
	ReadinessPollTimeout  string
	ReadinessPollInterval string
)

func LoadConfig() {
	KubeConfigPath = os.Getenv("KUBE_CONFIG_PATH")

	LogLevel = parseString("LOG_LEVEL", "DEBUG")

	Port = parseString("PORT", "80")
	MaxConsPerHost = parseInt("MAX_CONS_PER_HOST", "10000")

	Namespace = os.Getenv("NAMESPACE")

	ServerlessTTL = parseInt("SERVERLESS_TTL_SECONDS", "30")
	ReadinessPollTimeout = parseString("READINESS_POLL_TIMEOUT_SECONDS", "30")
	ReadinessPollInterval = parseString("READINESS_POLL_INTERVAL_SECONDS", "1")
}

func parseString(key, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" && defaultValue == "" {
		log.Panic().Msgf("Could not find env var: %v", key)
	} else if value == "" && defaultValue != "" {
		value = defaultValue
	}
	log.Info().Msgf("Successfully loaded env var: %v=%v", key, value)
	return value
}

func parseInt(key, defaultValue string) int {
	intValue, err := strconv.Atoi(parseString(key, defaultValue))
	if err != nil {
		log.Panic().Err(err).Msgf("%v should be an integer", key)
	}

	return intValue
}

func InitLogger() zerolog.Level {
	switch strings.ToUpper(LogLevel) {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "PANIC":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	return zerolog.GlobalLevel()
}
