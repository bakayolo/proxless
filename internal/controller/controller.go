package controller

import (
	"github.com/rs/zerolog/log"
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/config"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/model"
	"kube-proxless/internal/store"
	"time"
)

type Interface interface {
	GetRouteByDomainFromStore(domain string) (*model.Route, error)
	UpdateLastUseInStore(domain string) error
	ScaleUpDeployment(name, namespace string) error
	RunDownScaler(checkInterval int)
	RunServicesEngine()
}

type controller struct {
	store   store.Interface
	cluster cluster.Interface
}

func NewController(store store.Interface, cluster cluster.Interface) *controller {
	return &controller{
		store:   store,
		cluster: cluster,
	}
}

func (c *controller) GetRouteByDomainFromStore(domain string) (*model.Route, error) {
	return c.store.GetRouteByDomain(domain)
}

func (c *controller) UpdateLastUseInStore(domain string) error {
	return c.store.UpdateLastUse(domain)
}

func (c *controller) ScaleUpDeployment(name, namespace string) error {
	return c.cluster.ScaleUpDeployment(name, namespace, config.DeploymentReadinessTimeout)
}

func (c *controller) RunDownScaler(checkInterval int) {
	logger.Infof("Starting DownScaler...")

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf(nil, "DownScaler panic. Restarting...")
			c.RunDownScaler(checkInterval)
		}
	}()

	for {
		errs := scaleDownDeployments(c)

		for _, err := range errs {
			logger.Errorf(err, "Error during scale down")
		}

		time.Sleep(time.Duration(checkInterval) * time.Second)
	}
}

func scaleDownDeployments(c *controller) []error {
	return c.cluster.ScaleDownDeployments(
		config.Namespace,
		func(deployName, namespace string) (bool, error) {
			route, err := c.store.GetRouteByDeployment(deployName, namespace)

			if err != nil {
				logger.Errorf(err, "Could not get route %s.%s from store", deployName, namespace)
				return false, err
			}

			timeIdle := time.Now().Sub(route.GetLastUsed()).Seconds()
			if timeIdle >= float64(config.ServerlessTTL) {
				return true, nil
			}

			return false, nil
		})
}

func (c *controller) RunServicesEngine() {
	log.Info().Msgf("Starting Services Engine...")

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf(nil, "Services Engine panic. Restarting...")
			c.RunServicesEngine()
		}
	}()

	c.cluster.RunServicesEngine(
		config.Namespace,
		func(id, name, port, deployName, namespace string, domains []string) error {
			return c.store.UpsertStore(id, name, port, deployName, namespace, domains)
		},
		func(id string) error {
			return c.store.DeleteRoute(id)
		})
}
