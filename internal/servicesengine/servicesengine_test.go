package servicesengine

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	c = NewServicesEngine(&mockStore{}, &mockCluster{})
)

// TODO enrich testing - should test if the parameters are nil or not

func Test_labelDeployment(t *testing.T) {
	err := c.labelDeployment("", "")

	assert.NoError(t, err)
}

func Test_unlabelDeployment(t *testing.T) {
	err := c.unlabelDeployment("", "")

	assert.NoError(t, err)
}

func Test_deleteRouteFromStore(t *testing.T) {
	err := c.deleteRouteFromStore("")

	assert.NoError(t, err)
}

func Test_upsertStore(t *testing.T) {
	err := c.upsertStore("", "", "", "", "", []string{})

	assert.NoError(t, err)
}
