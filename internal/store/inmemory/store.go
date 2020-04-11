package inmemory

import (
	"fmt"
	"kube-proxless/internal/model"
	"kube-proxless/internal/utils"
)

func UpdateStore(identifier, service, port, deploy, namespace string, domains []string) error {
	if route, err := getRoute(identifier); err != nil {
		updateDomains([]string{}, domains, identifier)
		updateMappingDeployment("", deploy, namespace, identifier)
	} else {
		updateDomains(route.GetDomains(), domains, identifier)
		updateMappingDeployment(route.GetDeployment(), deploy, namespace, identifier)
	}

	newRoute, err := model.NewRoute(service, port, deploy, namespace, domains)

	if err != nil {
		return err
	}

	updateRoute(identifier, newRoute)

	return err
}

func updateDomains(oldDomains, newDomains []string, identifier string) {
	// get domains that are not in the new list
	var domainsToBeRemoved []string
	for _, domain := range oldDomains {
		if !utils.Contains(newDomains, domain) {
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

func GetRouteByDomainKey(domain string) (*model.Route, error) {
	return getRouteByMapping(domain)
}

func GetRouteByDeploymentKey(name, namespace string) (*model.Route, error) {
	return getRouteByMapping(genDeploymentKey(name, namespace))
}

func getRouteByMapping(key string) (*model.Route, error) {
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
