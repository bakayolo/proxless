package commons

import "github.com/rs/zerolog/log"

var (
	routesMap = make(map[string]string)
)

func UpdateRoute(key, value string) {
	routesMap[key] = value
}

func DeleteRoute(keys ...string) {
	for _, key := range keys {
		delete(routesMap, key)
	}
}

func GetRoute(key string) string {
	if routesMap[key] == "" {
		log.Error().Msg("Service % not found in routes map")
	}
	return routesMap[key]
}
