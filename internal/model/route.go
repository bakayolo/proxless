package model

import (
	"errors"
	"fmt"
	"time"
)

// /!\ the fields here are not exportable for extra safety
// We don't want to process dirty routes
// Use the constructor and the getter/setter to change the route
// https://github.com/golang/go/issues/28348#issuecomment-442250333
type Route struct {
	service    string
	port       string
	deployment string
	namespace  string
	domains    []string
	lastUsed   time.Time // TODO Need to store that in Kubernetes. This is not scalable!
}

func NewRoute(svc, port, deploy, ns string, domains []string) (*Route, error) {
	if svc == "" || deploy == "" || ns == "" || !isDomainsValid(domains) {
		return nil, errors.New(
			fmt.Sprintf(
				"Error creating route - svc = %s, deploy = %s, ns = %s, domains = %s - should not be nil",
				svc, deploy, ns, domains),
		)
	}

	return &Route{
		service:    svc,
		port:       useDefaultPortIfEmpty(port),
		deployment: deploy,
		namespace:  ns,
		domains:    domains,
		lastUsed:   time.Now(),
	}, nil
}

func isDomainsValid(domains []string) bool {
	return domains != nil && len(domains) > 0
}

func useDefaultPortIfEmpty(port string) string {
	if port == "" {
		return "80"
	}

	return port
}

func (r *Route) SetLastUsed(t time.Time) {
	r.lastUsed = t
}

func (r *Route) GetDomains() []string {
	return r.domains
}

func (r *Route) GetDeployment() string {
	return r.deployment
}

func (r *Route) GetPort() string {
	return r.port
}

func (r *Route) GetNamespace() string {
	return r.namespace
}

func (r *Route) GetService() string {
	return r.service
}

func (r *Route) GetLastUsed() time.Time {
	return r.lastUsed
}
