package model

import (
	"kube-proxless/internal/utils"
	"testing"
	"time"
)

func TestNewRoute(t *testing.T) {
	route := &Route{
		service:    "helloworld",
		port:       "8080",
		deployment: "helloworld",
		namespace:  "helloworld",
		domains:    []string{"helloworld.io"},
	}

	routeWithDefaultPort := *route
	routeWithDefaultPort.port = "80"

	testCases := []struct {
		id        int
		svc       string
		port      string
		deploy    string
		ns        string
		domains   []string
		want      *Route
		errWanted bool
	}{
		{ // route ok
			0,
			route.service,
			route.port,
			route.deployment,
			route.namespace,
			route.domains,
			route,
			false,
		},
		{ // service missing
			1,
			"",
			route.port,
			route.deployment,
			route.namespace,
			route.domains,
			route,
			true,
		},
		{ // deployment name missing
			2,
			route.service,
			route.port,
			"",
			route.namespace,
			route.domains,
			route,
			true,
		},
		{ // namespace missing
			3,
			route.service,
			route.port,
			route.deployment,
			"",
			route.domains,
			route,
			true,
		},
		{ // domains nil
			4,
			route.service,
			route.port,
			route.deployment,
			route.namespace,
			nil,
			route,
			true,
		},
		{ // domains empty
			5,
			route.service,
			route.port,
			route.deployment,
			route.namespace,
			[]string{},
			route,
			true,
		},
		{ // port empty -> must default to "80"
			6,
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
		got, errGot := NewRoute(tc.svc, tc.port, tc.deploy, tc.ns, tc.domains)

		if tc.errWanted != (errGot != nil) {
			t.Errorf("CreateRoute(tc %d) - must get an error", tc.id)
		}

		if errGot == nil && !compareRoute(got, tc.want) {
			t.Errorf("CreateRoute(tc %d) - got and want does not match", tc.id)
		}
	}
}

func TestRoute_isDomainsValid(t *testing.T) {
	testCases := []struct {
		d    []string
		want bool
	}{
		{[]string{"example.io"}, true},
		{[]string{}, false},
		{nil, false},
	}

	for _, tc := range testCases {
		got := isDomainsValid(tc.d)

		if got != tc.want {
			t.Errorf("isDomainsValid(%s) = %t, want %t", tc.d, got, tc.want)
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
