package model

import (
	"github.com/google/uuid"
	"kube-proxless/internal/utils"
	"testing"
	"time"
)

func TestNewRoute(t *testing.T) {
	route := &Route{
		id:         uuid.New().String(),
		service:    "helloworld",
		port:       "8080",
		deployment: "helloworld",
		namespace:  "helloworld",
		domains:    []string{"helloworld.io"},
	}

	routeWithDefaultPort := *route
	routeWithDefaultPort.port = "80"

	testCases := []struct {
		id        string
		svc       string
		port      string
		deploy    string
		ns        string
		domains   []string
		want      *Route
		errWanted bool
	}{
		{ // route ok
			route.id,
			route.service,
			route.port,
			route.deployment,
			route.namespace,
			route.domains,
			route,
			false,
		},
		{ // service missing
			route.id,
			"",
			route.port,
			route.deployment,
			route.namespace,
			route.domains,
			route,
			true,
		},
		{ // deployment name missing
			route.id,
			route.service,
			route.port,
			"",
			route.namespace,
			route.domains,
			route,
			true,
		},
		{ // namespace missing
			route.id,
			route.service,
			route.port,
			route.deployment,
			"",
			route.domains,
			route,
			true,
		},
		{ // domains nil
			route.id,
			route.service,
			route.port,
			route.deployment,
			route.namespace,
			nil,
			route,
			true,
		},
		{ // domains empty
			route.id,
			route.service,
			route.port,
			route.deployment,
			route.namespace,
			[]string{},
			route,
			true,
		},
		{ // port empty -> must default to "80"
			route.id,
			route.service,
			"",
			route.deployment,
			route.namespace,
			route.domains,
			&routeWithDefaultPort,
			false,
		},
	}

	for _, tc := range testCases {
		got, errGot := NewRoute(tc.id, tc.svc, tc.port, tc.deploy, tc.ns, tc.domains)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("CreateRoute(tc %s) = %v, errWanted = %t", tc.id, errGot, tc.errWanted)
		}

		if errGot == nil && !got.isEqual(tc.want) {
			t.Errorf("CreateRoute(tc %s) - got and want does not match", tc.id)
		}
	}
}

func TestRoute_useDefaultPortIfEmpty(t *testing.T) {
	testCases := []struct {
		p    string
		want string
	}{
		{"8080", "8080"},
		{"", "80"},
	}

	for _, tc := range testCases {
		got := useDefaultPortIfEmpty(tc.p)

		if got != tc.want {
			t.Errorf("useDefaultPortIfEmpty(%s) = %s, want %s", tc.p, got, tc.want)
		}
	}
}

// TODO lot of duplicate code becuz of getter and setter - see if we can factorize

func TestRoute_SetLastUsed(t *testing.T) {
	route := Route{}
	now := time.Now()
	route.SetLastUsed(now)

	if route.lastUsed != now {
		t.Errorf("SetLastUsed(%s) != want %s", route.lastUsed, now)
	}
}

func TestRoute_GetDeployment(t *testing.T) {
	route := Route{}
	route.deployment = "example"
	got := route.GetDeployment()

	if got != route.deployment {
		t.Errorf("GetDeployment() = %s, want %s", got, route.deployment)
	}
}

func TestRoute_GetDomains(t *testing.T) {
	route := Route{}
	route.domains = []string{"example.io", "example.com"}
	got := route.GetDomains()

	if !utils.CompareUnorderedArray(route.domains, got) {
		t.Errorf("GetDomains() = %s, want %s", got, route.domains)
	}
}

func TestRoute_GetNamespace(t *testing.T) {
	route := Route{}
	route.namespace = "example"
	got := route.GetNamespace()

	if got != route.namespace {
		t.Errorf("GetNamespace() = %s, want %s", got, route.namespace)
	}
}

func TestRoute_GetService(t *testing.T) {
	route := Route{}
	route.service = "example"
	got := route.GetService()

	if got != route.service {
		t.Errorf("GetService() = %s, want %s", got, route.service)
	}
}

func TestRoute_GetPort(t *testing.T) {
	route := Route{}
	route.port = "8080"
	got := route.GetPort()

	if got != route.port {
		t.Errorf("getPort() = %s, want %s", got, route.port)
	}
}

func TestRoute_GetLastUsed(t *testing.T) {
	route := Route{}
	route.lastUsed = time.Now()
	got := route.GetLastUsed()

	if got != route.lastUsed {
		t.Errorf("GetLastUsed() = %s, want %s", got, route.lastUsed)
	}
}

func TestRoute_GetId(t *testing.T) {
	route := Route{}
	route.id = uuid.New().String()
	got := route.GetId()

	if got != route.id {
		t.Errorf("getId() = %s, want %s", got, route.id)
	}
}

func TestRoute_SetDeployment(t *testing.T) {
	testCases := []struct {
		param     string
		errWanted bool
	}{
		{"example", false},
		{"", true},
	}

	for _, tc := range testCases {
		route := Route{}
		errGot := route.SetDeployment(tc.param)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("SetDeployment(%s) = %v; errWanted = %t", route.deployment, errGot, tc.errWanted)
		}

		if errGot == nil && route.deployment != tc.param {
			t.Errorf("SetDeployment(%s) != want %s", route.deployment, tc.param)
		}
	}
}

func TestRoute_SetNamespace(t *testing.T) {
	testCases := []struct {
		param     string
		errWanted bool
	}{
		{"example", false},
		{"", true},
	}

	for _, tc := range testCases {
		route := Route{}
		errGot := route.SetNamespace(tc.param)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("SetNamespace(%s) = %v; errWanted = %t", route.namespace, errGot, tc.errWanted)
		}

		if errGot == nil && route.namespace != tc.param {
			t.Errorf("SetNamespace(%s) != want %s", route.namespace, tc.param)
		}
	}
}

func TestRoute_SetPort(t *testing.T) {
	testCases := []struct {
		param     string
		errWanted bool
	}{
		{"8080", false},
		{"", true},
	}

	for _, tc := range testCases {
		route := Route{}
		errGot := route.SetPort(tc.param)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("SetPort(%s) = %v; errWanted = %t", route.port, errGot, tc.errWanted)
		}

		if errGot == nil && route.port != tc.param {
			t.Errorf("SetPort(%s) != want %s", route.port, tc.param)
		}
	}
}

func TestRoute_SetService(t *testing.T) {
	testCases := []struct {
		param     string
		errWanted bool
	}{
		{"example", false},
		{"", true},
	}

	for _, tc := range testCases {
		route := Route{}
		errGot := route.SetService(tc.param)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("SetService(%s) = %v; errWanted = %t", route.service, errGot, tc.errWanted)
		}

		if errGot == nil && route.service != tc.param {
			t.Errorf("SetService(%s) != want %s", route.service, tc.param)
		}
	}
}

func TestRoute_SetDomains(t *testing.T) {
	testCases := []struct {
		param     []string
		errWanted bool
	}{
		{[]string{"example.io", "example.com"}, false},
		{nil, true},
		{[]string{}, true},
	}

	for _, tc := range testCases {
		route := Route{}
		errGot := route.SetDomains(tc.param)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("SetDomains(%s) = %v; errWanted = %t", route.domains, errGot, tc.errWanted)
		}

		if errGot == nil && !utils.CompareUnorderedArray(tc.param, route.domains) {
			t.Errorf("SetDomains(%s) != want %s", route.domains, tc.param)
		}
	}
}
