package config

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
)

var (
	KubeConfigPath             string
	Port                       string
	MaxConsPerHost             int
	Namespace                  string
	ServerlessTTL              int
	DeploymentReadinessTimeout string
)

func LoadEnvVars() {
	KubeConfigPath = os.Getenv("KUBE_CONFIG_PATH")

	Port = parseString("PORT", "80")
	MaxConsPerHost = parseInt("MAX_CONS_PER_HOST", "10000")

	Namespace = os.Getenv("NAMESPACE")

	ServerlessTTL = parseInt("SERVERLESS_TTL_SECONDS", "30")
	DeploymentReadinessTimeout = parseString("DEPLOYMENT_READINESS_TIMEOUT_SECONDS", "30")
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
