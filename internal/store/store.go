package store

import "time"

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
	}
}

func UpdateLastUse(key string) {
	if identifier, err := getMapping(key); err == nil {
		updateLastUse(identifier)
	}
}
