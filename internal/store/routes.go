package store

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Route struct {
	Service    string
	Port       string
	Deployment string
	Namespace  string
	Domains    []string
	LastUsed   time.Time // TODO Need to store that in Kubernetes. This is not scalable!
}

type routesMapType struct {
	rmap map[string]Route
	lock sync.RWMutex
}

var (
	routesMap = routesMapType{
		rmap: make(map[string]Route),
		lock: sync.RWMutex{},
	}
)

func updateRoute(key string, route Route) {
	routesMap.lock.Lock()
	defer routesMap.lock.Unlock()
	routesMap.rmap[key] = route
}

func updateLastUse(key string) {
	routesMap.lock.Lock()
	defer routesMap.lock.Unlock()
	temp := routesMap.rmap[key]
	temp.LastUsed = time.Now()
	routesMap.rmap[key] = temp
}

func deleteRoute(key string) {
	routesMap.lock.Lock()
	defer routesMap.lock.Unlock()
	delete(routesMap.rmap, key)
}

func getRoute(key string) (*Route, error) {
	routesMap.lock.RLock()
	defer routesMap.lock.RUnlock()
	if r, ok := routesMap.rmap[key]; ok {
		return &r, nil
	}

	return nil, errors.New(fmt.Sprintf("Service %s not found in routes map", key))
}
