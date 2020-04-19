package servicesengine

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/config"
	"kube-proxless/internal/store"
)

type ServicesEngine struct {
	store   store.StoreInterface
	cluster cluster.ClusterInterface
}

func NewServicesEngine(store store.StoreInterface, cluster cluster.ClusterInterface) *ServicesEngine {
	return &ServicesEngine{
		store:   store,
		cluster: cluster,
	}
}

func (e *ServicesEngine) Run() {
	log.Info().Msgf("Starting Services Enginer")

	e.cluster.RunServicesEngine(
		config.Namespace,
		e.labelDeployment,
		e.unlabelDeployment,
		e.upsertStore,
		e.deleteRouteFromStore,
	)
}

func (e *ServicesEngine) labelDeployment(deployName, namespace string) error {
	return e.cluster.LabelDeployment(deployName, namespace)
}

func (e *ServicesEngine) unlabelDeployment(deployName, namespace string) error {
	return e.cluster.UnlabelDeployment(deployName, namespace)
}

func (e *ServicesEngine) upsertStore(id, name, port, deployName, namespace string, domains []string) error {
	return e.store.UpsertStore(id, name, port, deployName, namespace, domains)
}

func (e *ServicesEngine) deleteRouteFromStore(id string) error {
	return e.store.DeleteRoute(id)
}
