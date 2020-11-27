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
	UpdateIsRunningInMemory(id string) error
	ScaleUpDeployment(name, namespace string, readinessTimeoutSeconds int) error
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
		c.pubsub.PublishLastUsed(id, now)
	}

	return c.memory.UpdateLastUsed(id, now)
}

func (c *controller) UpdateIsRunningInMemory(id string) error {
	if c.pubsub != nil {
		c.pubsub.PublishIsRunning(id, true)
	}

	return c.memory.UpdateIsRunning(id, true)
}

func (c *controller) ScaleUpDeployment(name, namespace string, readinessTimeoutSeconds int) error {
	return c.cluster.ScaleUpDeployment(name, namespace, readinessTimeoutSeconds)
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
	deploymentsToScaleDown := c.memory.GetRoutesToScaleDown()

	var errs []error

	for _, route := range deploymentsToScaleDown {
		err := c.cluster.ScaleDownDeployment(route.GetDeployment(), route.GetNamespace())

		if err != nil {
			errs = append(errs, err)
		} else {
			_ = c.memory.UpdateIsRunning(route.GetId(), false)

			if c.pubsub != nil {
				c.pubsub.PublishIsRunning(route.GetId(), false)
			}
		}
	}

	return errs
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
		func(
			id, name, port, deployName, namespace string,
			domains []string,
			ttlSeconds, readinessTimeoutSeconds *int) error {
			if c.pubsub != nil {
				c.pubsub.SubscribeLastUsed(id, c.memory.UpdateLastUsed)
				c.pubsub.SubscribeIsRunning(id, c.memory.UpdateIsRunning)
			}

			route, err :=
				model.NewRoute(id, name, port, deployName, namespace, domains, ttlSeconds, readinessTimeoutSeconds)

			if err != nil {
				return err
			}

			return c.memory.UpsertMemoryMap(route)
		},
		func(id string) error {
			if c.pubsub != nil {
				c.pubsub.Unsubscribe(id)
			}

			return c.memory.DeleteRoute(id)
		})
}
