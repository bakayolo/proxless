package store

import (
	"errors"
	"github.com/rs/zerolog/log"
)

type Route struct {
	Service   string
	Port      string
	Label     string
	Namespace string
}

var (
	routesMap = make(map[string]Route)
)

func UpdateRoute(key, service, port, label, namespace string) {
	routesMap[key] = Route{
		Service:   service,
		Port:      port,
		Label:     label,
		Namespace: namespace,
	}
}

func DeleteRoute(keys ...string) {
	for _, key := range keys {
		delete(routesMap, key)
	}
}

func GetRoute(key string) (Route, error) {
	if r, ok := routesMap[key]; ok {
		return r, nil
	}

	log.Error().Msgf("Service %s not found in routes map", key)
	return Route{}, errors.New("Service not found in routes map")
}
