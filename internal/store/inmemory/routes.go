package inmemory

import (
	"errors"
	"fmt"
	"kube-proxless/internal/model"
	"sync"
	"time"
)

type routesMapType struct {
	rmap map[string]*model.Route
	lock sync.RWMutex
}

var (
	routesMap = routesMapType{
		rmap: make(map[string]*model.Route),
		lock: sync.RWMutex{},
	}
)

func updateRoute(key string, route *model.Route) {
	routesMap.lock.Lock()
	defer routesMap.lock.Unlock()
	routesMap.rmap[key] = route
}

func updateLastUse(key string) {
	routesMap.lock.Lock()
	defer routesMap.lock.Unlock()
	temp := routesMap.rmap[key]
	temp.SetLastUsed(time.Now())
	routesMap.rmap[key] = temp
}

func deleteRoute(key string) {
	routesMap.lock.Lock()
	defer routesMap.lock.Unlock()
	delete(routesMap.rmap, key)
}

func getRoute(key string) (*model.Route, error) {
	routesMap.lock.RLock()
	defer routesMap.lock.RUnlock()
	if r, ok := routesMap.rmap[key]; ok {
		return r, nil
	}

	return nil, errors.New(fmt.Sprintf("Service %s not found in routes map", key))
}
