package store

import (
	"fmt"
	"github.com/rs/zerolog/log"
)

type route struct {
	service   string
	port      string
	label     string
	namespace string
}

var (
	routesMap = make(map[string]route)
)

func UpdateRoute(key, service, port, label, namespace string) {
	routesMap[key] = route{
		service:   service,
		port:      port,
		label:     label,
		namespace: namespace,
	}
}

func DeleteRoute(keys ...string) {
	for _, key := range keys {
		delete(routesMap, key)
	}
}

func GetRoute(key string) string {
	if r, ok := routesMap[key]; ok {
		return fmt.Sprintf("%s:%s", r.service, r.port)
	}

	log.Error().Msgf("Service %s not found in routes map", key)
	return ""
}
