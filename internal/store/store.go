package store

<<<<<<< HEAD
import (
	"errors"
	"github.com/rs/zerolog/log"
)
=======
import "time"
>>>>>>> 🚧 Autoscale the deployment after N seconds

func UpdateStore(identifier, service, port, label, namespace string, domains []string) {
	if route, err := getRoute(identifier); err != nil { // new route
		updateRoute(identifier, Route{
			Service:   service,
			Port:      port,
			Label:     label,
			Namespace: namespace,
			LastUsed:  time.Now(), // default to time.Now()
		})
	} else { // update route
		// TODO should handle the domain removed from the service or change in the label
		route.Service = service
		route.Port = port
		route.Label = label
		route.Namespace = namespace
		updateRoute(identifier, route)
	}
	for _, domain := range domains {
		updateMapping(domain, identifier)
	}
	updateMapping(label, identifier)
}

<<<<<<< HEAD
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
=======
func DeleteObjectInStore(identifier string) {
	deleteRoute(identifier)
	deleteMappingByValue(identifier)
}

func GetRouteByDomain(key string) (Route, error) {
	return getRouteByMapping(key)
}

func GetRouteByLabel(key string) (Route, error) {
	return getRouteByMapping(key)
}

func getRouteByMapping(key string) (Route, error) {
	if identifier, err := getMapping(key); err != nil {
		return Route{}, err
	} else {
		return getRoute(identifier)
>>>>>>> 🚧 Autoscale the deployment after N seconds
	}
}

func UpdateLastUse(key string) {
	if identifier, err := getMapping(key); err == nil {
		updateLastUse(identifier)
	}
}
