package controller

import (
	"kube-proxless/internal/cluster"
	"kube-proxless/internal/config"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/memory"
	"kube-proxless/internal/model"
	"kube-proxless/internal/pubsub"
	"time"
)

type Interface interface {
	GetRouteByDomainFromMemory(domain string) (*model.Route, error)
	UpdateLastUsedInMemory(id string) error
	ScaleUpDeployment(name, namespace string) error
	RunDownScaler(checkInterval int)
	RunServicesEngine()
}

type controller struct {
	memory  memory.Interface
	cluster cluster.Interface
	pubsub  pubsub.Interface
}

func NewController(memory memory.Interface, cluster cluster.Interface, ps pubsub.Interface) *controller {
	return &controller{
		memory:  memory,
		cluster: cluster,
		pubsub:  ps,
	}
}

func (c *controller) GetRouteByDomainFromMemory(domain string) (*model.Route, error) {
	return c.memory.GetRouteByDomain(domain)
}

func (c *controller) UpdateLastUsedInMemory(id string) error {
	now := time.Now()
	if c.pubsub != nil {
		c.pubsub.Publish(id, now)
	}

	return c.memory.UpdateLastUsed(id, now)
}

func (c *controller) ScaleUpDeployment(name, namespace string) error {
	return c.cluster.ScaleUpDeployment(name, namespace, config.DeploymentReadinessTimeoutSeconds)
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
		config.NamespaceScope,
		func(deployName, namespace string) (bool, time.Duration, error) {
			route, err := c.memory.GetRouteByDeployment(deployName, namespace)

			if err != nil {
				logger.Errorf(err, "Could not get route %s.%s from memory", deployName, namespace)
				return false, 0, err
			}

			timeIdle := time.Now().Sub(route.GetLastUsed())
			// https://stackoverflow.com/a/41503910/5683655
			if int64(timeIdle/time.Second) >= int64(config.ServerlessTTLSeconds) {
				return true, timeIdle, nil
			}

			return false, timeIdle, nil
		})
}

func (c *controller) RunServicesEngine() {
	logger.Infof("Starting Services Engine...")

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf(nil, "Services Engine panic. Restarting...")
			c.RunServicesEngine()
		}
	}()

	c.cluster.RunServicesEngine(
		config.NamespaceScope,
		config.ProxlessService,
		config.ProxlessNamespace,
		func(id, name, port, deployName, namespace string, domains []string) error {
			if c.pubsub != nil {
				c.pubsub.Subscribe(id, c.memory.UpdateLastUsed)
			}

			return c.memory.UpsertMemoryMap(id, name, port, deployName, namespace, domains)
		},
		func(id string) error {
			if c.pubsub != nil {
				c.pubsub.Unsubscribe(id)
			}

			return c.memory.DeleteRoute(id)
		})
}
