package memory

import (
	"errors"
	"fmt"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/model"
	"kube-proxless/internal/utils"
	"sync"
	"time"
)

type Interface interface {
	UpsertMemoryMap(route *model.Route) error
	GetRouteByDomain(domain string) (*model.Route, error)
	GetRouteByDeployment(deploy, namespace string) (*model.Route, error)
	UpdateLastUsed(id string, t time.Time) error
	DeleteRoute(id string) error
}

type MemoryMap struct {
	m    map[string]*model.Route
	lock sync.RWMutex
}

func NewMemoryMap() *MemoryMap {
	return &MemoryMap{
		m:    make(map[string]*model.Route),
		lock: sync.RWMutex{},
	}
}

func (s *MemoryMap) UpsertMemoryMap(route *model.Route) error {
	// error if deployment or domains are already associated to another route
	err := checkDeployAndDomainsOwnership(
		s, route.GetId(), route.GetDeployment(), route.GetNamespace(), route.GetDomains())

	if err != nil {
		return err
	}

	if existingRoute, ok := s.m[route.GetId()]; ok {
		// /!\ this need to be on top - otherwise the data will have already been overriden in the route
		newKeys := cleanMemoryMap(
			s,
			existingRoute.GetDeployment(), existingRoute.GetNamespace(), existingRoute.GetDomains(),
			route.GetDeployment(), route.GetNamespace(), route.GetDomains())

		// associate the route to new deployment key / domains
		for _, k := range newKeys {
			s.m[k] = existingRoute
		}

		// TODO check the errors
		_ = existingRoute.SetService(route.GetService())
		_ = existingRoute.SetPort(route.GetPort())
		_ = existingRoute.SetDeployment(route.GetDeployment())
		_ = existingRoute.SetDomains(route.GetDomains())
		existingRoute.SetTTLSeconds(route.GetTTLSeconds())
		existingRoute.SetReadinessTimeoutSeconds(route.GetReadinessTimeoutSeconds())
		// existingRoute is a pointer and it's changing dynamically - no need to "persist" the change in the map

		keys := append(
			[]string{route.GetId(), genDeploymentKey(existingRoute.GetDeployment(), existingRoute.GetNamespace())},
			route.GetDomains()...)
		logger.Debugf("Updated route - newKeys: [%s] - keys: [%s] - obj: %v", newKeys, keys, existingRoute)
	} else {
		createRoute(s, route)
	}

	return nil
}

// return an error if deploy or domains are already associated to a different id
func checkDeployAndDomainsOwnership(s *MemoryMap, id, deploy, ns string, domains []string) error {
	r, err := s.GetRouteByDeployment(deploy, ns)

	if err == nil && r.GetId() != id {
		return errors.New(fmt.Sprintf("Deployment %s.%s is already owned by %s", deploy, ns, r.GetId()))
	}

	for _, d := range domains {
		r, err = s.GetRouteByDomain(d)

		if err == nil && r.GetId() != id {
			return errors.New(fmt.Sprintf("Domain %s is already owned by %s", d, r.GetId()))
		}
	}

	return nil
}

func createRoute(s *MemoryMap, route *model.Route) {
	s.lock.Lock()
	defer s.lock.Unlock()

	deploymentKey := genDeploymentKey(route.GetDeployment(), route.GetNamespace())
	s.m[route.GetId()] = route
	s.m[deploymentKey] = route
	for _, d := range route.GetDomains() {
		s.m[d] = route
	}

	keys := append([]string{route.GetId(), deploymentKey}, route.GetDomains()...)
	logger.Debugf("Created route - keys: [%s] - obj: %v", keys, route)
}

// Remove old domains and deployment from the map if they are not == new ones
// return the domains and deployment that are not a key in the map
func cleanMemoryMap(
	s *MemoryMap,
	oldDeploy, oldNs string, oldDomains []string,
	newDeploy, newNs string, newDomains []string) []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	var newKeys []string

	deployKeyNotInMap := cleanOldDeploymentFromMemoryMap(s, oldDeploy, oldNs, newDeploy, newNs)

	if deployKeyNotInMap != "" {
		newKeys = append(newKeys, deployKeyNotInMap)
	}

	domainsNotInMap := cleanOldDomainsFromMemoryMap(s, oldDomains, newDomains)

	if newDomains != nil {
		newKeys = append(newKeys, domainsNotInMap...)
	}

	if newKeys == nil {
		return []string{}
	}

	return newKeys
}

// return the new deployment key if it does not exist in the map
func cleanOldDeploymentFromMemoryMap(s *MemoryMap, oldDeploy, oldNs, newDeploy, newNs string) string {
	oldDeploymentKey := genDeploymentKey(oldDeploy, oldNs)
	newDeploymentKey := genDeploymentKey(newDeploy, newNs)

	if oldDeploymentKey != newDeploymentKey {
		delete(s.m, oldDeploymentKey)
		return newDeploymentKey
	}

	return ""
}

// TODO review complexity
// return the new domains that are not in the newDomains list
func cleanOldDomainsFromMemoryMap(s *MemoryMap, oldDomains, newDomains []string) []string {
	// get the difference between the 2 domains arrays
	diff := utils.DiffUnorderedArray(oldDomains, newDomains)

	var newKeys []string

	if diff != nil && len(diff) > 0 {
		// remove domain from the map if they are not in the list of new Domains
		for _, d := range diff {
			if !utils.Contains(newDomains, d) {
				delete(s.m, d)
			} else {
				newKeys = append(newKeys, d)
			}
		}
	}

	if newKeys == nil {
		return []string{}
	}

	return newKeys
}

func genDeploymentKey(deployment, namespace string) string {
	return fmt.Sprintf("%s.%s", deployment, namespace)
}

func (s *MemoryMap) GetRouteByDomain(domain string) (*model.Route, error) {
	return getRoute(s, domain)
}

func (s *MemoryMap) GetRouteByDeployment(deploy, namespace string) (*model.Route, error) {
	deploymentKey := genDeploymentKey(deploy, namespace)
	return getRoute(s, deploymentKey)
}

func getRoute(s *MemoryMap, key string) (*model.Route, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if route, ok := s.m[key]; ok {
		return route, nil
	}

	return nil, errors.New(fmt.Sprintf("Route %s not found in map", key))
}

func (s *MemoryMap) UpdateLastUsed(id string, t time.Time) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if route, ok := s.m[id]; ok {
		// No need to persist in the map, it's a pointer
		route.SetLastUsed(t)
		return nil
	}

	return errors.New(fmt.Sprintf("Route %s not found in map", id))
}

func (s *MemoryMap) DeleteRoute(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if route, ok := s.m[id]; ok {
		deploymentKey := genDeploymentKey(route.GetDeployment(), route.GetNamespace())
		delete(s.m, route.GetId())
		delete(s.m, deploymentKey)
		for _, d := range route.GetDomains() {
			delete(s.m, d)
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Route %s not found in map", id))
}
