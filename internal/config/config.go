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

func LoadEnvVars() {
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
	} else if value == "" {
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
	case "INFO", "WARN":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "ERROR", "FATAL", "PANIC":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	return zerolog.GlobalLevel()
}
