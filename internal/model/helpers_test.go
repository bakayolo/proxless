package model

import "testing"

func Test_compareRoute(t *testing.T) {
	route := &Route{
		service:    "helloworld",
		port:       "8080",
		deployment: "helloworld",
		namespace:  "helloworld",
		domains:    []string{"helloworld.io", "helloworld.com"},
	}

	routeWithDiffSvc := *route
	routeWithDiffSvc.service = "diff"

	routeWithDiffPort := *route
	routeWithDiffPort.port = "diff"

	routeWithDiffDeploy := *route
	routeWithDiffDeploy.deployment = "diff"

	routeWithDiffNs := *route
	routeWithDiffNs.namespace = "diff"

	routeWithDiffDomains := *route
	routeWithDiffDomains.domains = []string{"diff.io"}

	routeWithSameUnorderedDomains := *route
	routeWithSameUnorderedDomains.domains = []string{"helloworld.com", "helloworld.io"}

	testCases := []struct {
		id     int
		r1, r2 *Route
		want   bool
	}{
		{
			0,
			route,
			route,
			true,
		},
		{
			1,
			route,
			&routeWithDiffSvc,
			false,
		},
		{
			2,
			route,
			&routeWithDiffPort,
			false,
		},
		{
			3,
			route,
			&routeWithDiffDeploy,
			false,
		},
		{
			4,
			route,
			&routeWithDiffNs,
			false,
		},
		{
			5,
			route,
			&routeWithDiffDomains,
			false,
		},
		{
			6,
			route,
			&routeWithSameUnorderedDomains,
			true,
		},
	}

	for _, tc := range testCases {
		got := compareRoute(tc.r1, tc.r2)

		if got != tc.want {
			t.Errorf("compareRoute(tc %d) = %t; want %t", tc.id, got, tc.want)
		}
	}
}
