package store

import (
	"errors"
	"github.com/rs/zerolog/log"
	"sync"
)

// Mapping between `domain`/`label` and UID

type mappingMapType struct {
	mmap map[string]string
	lock sync.RWMutex
}

var (
	mappingMap = mappingMapType{
		mmap: make(map[string]string),
		lock: sync.RWMutex{},
	}
)

func updateMapping(key, uid string) {
	mappingMap.lock.Lock()
	defer mappingMap.lock.Unlock()
	mappingMap.mmap[key] = uid
}

// TODO not performant
func deleteMappingByValue(value string) {
	mappingMap.lock.Lock()
	defer mappingMap.lock.Unlock()
	for k, v := range mappingMap.mmap {
		if v == value {
			delete(mappingMap.mmap, k)
		}
	}
}

func getMapping(key string) (string, error) {
	mappingMap.lock.RLock()
	defer mappingMap.lock.RUnlock()
	if v, ok := mappingMap.mmap[key]; ok {
		return v, nil
	}

	log.Error().Msgf("Mapping %s not found in map", key)
	return "", errors.New("Mapping not found in map")
}
