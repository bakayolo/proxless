package kube

import (
	"kube-proxless/internal/config"
	"testing"
)

var (
	k = KubeClient{
		deployClient: &mockDeploymentClient{},
	}
)

func TestKubeClient_waitForDeploymentAvailable(t *testing.T) {
	config.ReadinessPollInterval = "1"
	config.ReadinessPollTimeout = "1"

	testCases := []struct {
		name      string // name define the output of the `getDeployment`
		errWanted bool
	}{
		{"err", true},
		{"timeout", true},
		{"mock", false},
	}

	for _, tc := range testCases {
		errGot := k.waitForDeploymentAvailable(tc.name, "")

		if tc.errWanted != (errGot != nil) {
			t.Errorf("waitForDeploymentAvailable(%s, _) = %v; error wanted = %t", tc.name, errGot, tc.errWanted)
		}
	}
}

func TestKubeClient_ScaleUpDeployment(t *testing.T) {
	config.ReadinessPollInterval = "1"
	config.ReadinessPollTimeout = "1"

	testCases := []struct {
		name      string // name define the output of the `getDeployment`
		errWanted bool
	}{
		{"err", true},
		{"timeout", true},
		{"mock", false},
	}

	for _, tc := range testCases {
		errGot := k.ScaleUpDeployment(tc.name, "")

		if tc.errWanted != (errGot != nil) {
			t.Errorf("ScaleUpDeployment(%s, _) = %v; error wanted = %t", tc.name, errGot, tc.errWanted)
		}
	}
}
