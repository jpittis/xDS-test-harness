package test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jpittis/xDS-test-harness/pkg/admin"
	"github.com/jpittis/xDS-test-harness/pkg/shim"
	"github.com/stretchr/testify/require"
)

const shimTestHost = "127.0.0.1:4678"
const envoyTestHost = "127.0.0.1:9901"

func TestWorking(t *testing.T) {
	shimClient := &shim.Client{
		Host: shimTestHost,
		HTTPClient: &http.Client{
			Timeout: time.Second,
		},
	}

	adminClient := &admin.Client{
		Host: envoyTestHost,
		HTTPClient: &http.Client{
			Timeout: time.Second,
		},
	}

	for i := 0; i < 2; i++ {
		t.Logf("Attempt %d", i)

		t.Log("Double checking server is stopped")
		_, err := shimClient.StopServer()
		require.NoError(t, err)

		t.Log("Starting server")
		_, err = shimClient.StartServer()
		require.NoError(t, err)

		t.Log("Setting snapshot")
		_, err = shimClient.SetSnapshot()
		require.NoError(t, err)

		t.Log("Testing snapshot")
		var success bool
		for i := 0; i < 15; i++ {
			configDump, err := adminClient.ConfigDump()
			require.NoError(t, err)
			fmt.Println(configDump.ClustersConfigDump.DynamicActiveClusters)
			if len(configDump.ClustersConfigDump.DynamicActiveClusters) > 0 &&
				configDump.ClustersConfigDump.DynamicActiveClusters[0].Cluster.Name == "some_service" {
				success = true
				break

			}

			t.Logf("Polling config dump %d", i)
			time.Sleep(time.Second)
		}
		require.True(t, success)

		t.Log("Getting original uptime")
		originalServerInfo, err := adminClient.ServerInfo()
		require.NoError(t, err)

		t.Log("Stopping server")
		_, err = shimClient.StopServer()
		require.NoError(t, err)

		t.Log("Quit quit quit")
		err = adminClient.QuitQuitQuit()
		require.NoError(t, err)

		success = false
		for i := 0; i < 15; i++ {
			currentServerInfo, err := adminClient.ServerInfo()
			require.NoError(t, err)
			currentUptime, err := types.DurationFromProto(currentServerInfo.UptimeAllEpochs)
			require.NoError(t, err)
			originalUptime, err := types.DurationFromProto(originalServerInfo.UptimeAllEpochs)
			require.NoError(t, err)

			if err == nil && currentUptime < originalUptime {
				success = true
				break

			}
			t.Logf("Polling uptime %d: %s < %s", i, currentUptime, originalUptime)
			time.Sleep(time.Second)
		}
		require.True(t, success)
	}
}
