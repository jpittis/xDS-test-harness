package test

import (
	"log"
	"testing"
	"time"

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

	err := h.WithFreshEnvoy(func(h *harness.Handle) error {
		// This sends a hardcoded snapshot with a single "some_service" cluster.
		log.Println("Sending snapshot")
		_, err := h.Shim.SetSnapshot()
		if err != nil {
			return err
		}

		log.Println("Waiting for config dump")
		// return h.WaitConfigDump(func(configDump *admin.ConfigDump) bool {
		// 	dynamicActiveClusters := configDump.ClustersConfigDump.DynamicActiveClusters
		// 	return len(dynamicActiveClusters) > 0 &&
		// 		dynamicActiveClusters[0].Cluster.Name == "some_service"
		// }, defaultTimeout)
		return nil
	})

	require.NoError(t, err)
}
