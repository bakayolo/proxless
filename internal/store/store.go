package store

import (
	"fmt"
	"time"
)

func UpdateStore(identifier, service, port, deploy, namespace string, domains []string) {
	if route, err := getRoute(identifier); err != nil { // new route
		updateDomains([]string{}, domains, identifier)
		updateMappingDeployment("", deploy, namespace, identifier)
		updateRoute(identifier, Route{
			Service:    service,
			Port:       port,
			Deployment: deploy,
			Namespace:  namespace,
			Domains:    domains,
			LastUsed:   time.Now(), // default to time.Now()
		})
	} else { // update route
		updateDomains(route.Domains, domains, identifier)
		updateMappingDeployment(route.Deployment, deploy, namespace, identifier)
		route.Service = service
		route.Port = port
		route.Deployment = deploy
		route.Namespace = namespace
		route.Domains = domains
		updateRoute(identifier, *route)
	}
}

func updateDomains(oldDomains, newDomains []string, identifier string) {
	// get domains that are not in the new list
	var domainsToBeRemoved []string
	for _, domain := range oldDomains {
		if !contains(newDomains, domain) {
			domainsToBeRemoved = append(domainsToBeRemoved, domain)
		}
	}

	// remove the domains that are not in the new list
	for _, domain := range domainsToBeRemoved {
		deleteMappingByKey(domain)
	}

	for _, domain := range newDomains {
		updateMappingDomain(domain, identifier)
	}
}

func updateMappingDomain(domain, identifier string) {
	updateMapping(domain, identifier)
}

func updateMappingDeployment(oldName, newName, namespace, identifier string) {
	if oldName != newName { // need to remove the old name from the map
		deleteMappingByKey(genDeploymentKey(oldName, namespace))
	}
	updateMapping(genDeploymentKey(newName, namespace), identifier)
}

func DeleteObjectInStore(identifier string) {
	deleteRoute(identifier)
	deleteMappingByValue(identifier)
}

func GetRouteByDomainKey(domain string) (*Route, error) {
	return getRouteByMapping(domain)
}

func GetRouteByDeploymentKey(name, namespace string) (*Route, error) {
	return getRouteByMapping(genDeploymentKey(name, namespace))
}

func getRouteByMapping(key string) (*Route, error) {
	if identifier, err := getMappingValueByKey(key); err != nil {
		return nil, err
	} else {
		return getRoute(identifier)
	}
}

func UpdateLastUse(domain string) {
	if identifier, err := getMappingValueByKey(domain); err == nil {
		updateLastUse(identifier)
	}
}

func genDeploymentKey(name, namespace string) string {
	return fmt.Sprintf("%s.%s", name, namespace)
}
