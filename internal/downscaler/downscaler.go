package downscaler

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/config"
	"kube-proxless/internal/store"
	"time"
)

type DownScaler struct {
	store   store.StoreInterface
	cluster cluster.ClusterInterface
}

func NewDownScaler(store store.StoreInterface, cluster cluster.ClusterInterface) *DownScaler {
	return &DownScaler{
		store:   store,
		cluster: cluster,
	}
}

func (ds *DownScaler) Run() {
	log.Info().Msgf("Starting DownScaler")

	// TODO see if we wanna use `ServerlessTTL` for the interval check
	checkInterval := time.Duration(config.ServerlessTTL) * time.Second

	ds.cluster.RunDownScaler(config.Namespace, checkInterval, ds.mustScaleDown)
}

func (ds *DownScaler) mustScaleDown(deployName, namespace string) bool {
	route, err := ds.store.GetRouteByDeployment(deployName, namespace)

	if err != nil {
		log.Error().Err(err).Msgf("Could not get route %s.%s from store", deployName, namespace)
		return false
	}

	timeIdle := time.Now().Sub(route.GetLastUsed()).Seconds()
	if timeIdle >= float64(config.ServerlessTTL) {
		return true
	}

	return false
}
