package test

import (
	"testing"
	"time"

	"github.com/jpittis/xDS-test-harness/pkg/admin"
	"github.com/jpittis/xDS-test-harness/pkg/harness"
	"github.com/stretchr/testify/require"
)

const (
	defaultTimeout = 10 * time.Second
	shimTestHost   = "127.0.0.1:4678"
	envoyTestHost  = "127.0.0.1:9901"
)

func TestWorking(t *testing.T) {
	h := harness.NewHandle(shimTestHost, envoyTestHost)

	t.Log("Starting server")
	_, err := h.Shim.StartServer()
	require.NoError(t, err)

	t.Log("Setting snapshot")
	_, err = h.Shim.SetSnapshot()
	require.NoError(t, err)

	t.Log("Testing snapshot")
	err = h.WaitConfigDump(func(configDump *admin.ConfigDump) bool {
		dynamicActiveClusters := configDump.ClustersConfigDump.DynamicActiveClusters
		return len(dynamicActiveClusters) > 0 &&
			dynamicActiveClusters[0].Cluster.Name == "some_service"
	}, defaultTimeout)
	require.NoError(t, err)

	t.Log("Stopping server")
	_, err = h.Shim.StopServer()
	require.NoError(t, err)

	t.Log("Restarting envoy")
	err = h.WaitRestart(defaultTimeout)
	require.NoError(t, err)
}
