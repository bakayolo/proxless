package model

import (
	"errors"
	"fmt"
	"kube-proxless/internal/utils"
	"time"
)

// /!\ the fields here are not exportable for extra safety
// We don't want to process dirty routes
// Use the constructor and the getter/setter to change the route
// https://github.com/golang/go/issues/28348#issuecomment-442250333
type Route struct {
	id                      string
	service                 string
	port                    string
	deployment              string
	namespace               string
	domains                 []string
	lastUsed                time.Time
	ttlSeconds              *int
	readinessTimeoutSeconds *int
}

func NewRoute(
	id, svc, port, deploy, ns string, domains []string, ttlSeconds, readinessTimeoutSeconds *int) (*Route, error) {
	if id == "" || svc == "" || deploy == "" || ns == "" || utils.IsArrayEmpty(domains) {
		return nil, errors.New(
			fmt.Sprintf(
				"Error creating route - id = %s, svc = %s, deploy = %s, ns = %s, domains = %s - must not be empty",
				id, svc, deploy, ns, domains),
		)
	}

	return &Route{
		id:                      id,
		service:                 svc,
		port:                    useDefaultPortIfEmpty(port),
		deployment:              deploy,
		namespace:               ns,
		domains:                 domains,
		lastUsed:                time.Now(),
		ttlSeconds:              ttlSeconds,
		readinessTimeoutSeconds: readinessTimeoutSeconds,
	}, nil
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

func (r *Route) SetDomains(d []string) error {
	if d == nil || len(d) == 0 {
		return errors.New(fmt.Sprintf("SetDomains for route %v should not be empty", r))
	}
	r.domains = d
	return nil
}

func (r *Route) SetService(s string) error {
	if s == "" {
		return errors.New(fmt.Sprintf("SetSercice for route %v should not be empty", r))
	}
	r.service = s
	return nil
}

func (r *Route) SetPort(p string) error {
	if p == "" {
		return errors.New(fmt.Sprintf("SetPort for route %v should not be empty", r))
	}
	r.port = p
	return nil
}

func (r *Route) SetDeployment(d string) error {
	if d == "" {
		return errors.New(fmt.Sprintf("SetDeployment for route %v should not be empty", r))
	}
	r.deployment = d
	return nil
}

func (r *Route) SetNamespace(n string) error {
	if n == "" {
		return errors.New(fmt.Sprintf("SetNamespace for route %v should not be empty", r))
	}
	r.namespace = n
	return nil
}

func (r *Route) SetTTLSeconds(t *int) {
	r.ttlSeconds = t
}

func (r *Route) SetReadinessTimeoutSeconds(t *int) {
	r.readinessTimeoutSeconds = t
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

func (r *Route) GetId() string {
	return r.id
}

func (r *Route) GetTTLSeconds() *int {
	return r.ttlSeconds
}

func (r *Route) GetReadinessTimeoutSeconds() *int {
	return r.readinessTimeoutSeconds
}
