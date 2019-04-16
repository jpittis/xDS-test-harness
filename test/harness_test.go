package test

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jpittis/xDS-test-harness/pkg/admin"
	"github.com/stretchr/testify/require"
)

const shimAddr = "http://127.0.0.1:4678"
const envoyTestHost = "127.0.0.1:9901"

func TestWorking(t *testing.T) {
	client := NewShimClient(shimAddr, envoyTestHost)

	adminClient := &admin.Client{
		Host: envoyTestHost,
		HTTPClient: &http.Client{
			Timeout: time.Second,
		},
	}

	for i := 0; i < 2; i++ {
		t.Logf("Attempt %d", i)

		t.Log("Double checking server is stopped")
		err := client.StopServer()
		require.NoError(t, err)

		t.Log("Starting server")
		err = client.StartServer()
		require.NoError(t, err)

		t.Log("Setting snapshot")
		err = client.SetSnapshot()
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
		err = client.StopServer()
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

func NewShimClient(shimAddr, envoyAddr string) *ShimClient {
	return &ShimClient{
		shimAddr:  shimAddr,
		envoyAddr: envoyAddr,
		client: &http.Client{
			Timeout: time.Second,
		},
	}
}

type ShimClient struct {
	shimAddr  string
	envoyAddr string
	client    *http.Client
}

func (c *ShimClient) StartServer() error {
	return c.post(c.shimAddr, "/start_server", nil)
}

func (c *ShimClient) StopServer() error {
	return c.post(c.shimAddr, "/stop_server", nil)
}

func (c *ShimClient) SetSnapshot() error {
	return c.post(c.shimAddr, "/set_snapshot", nil)
}

func (c *ShimClient) post(addr string, path string, body io.Reader) error {
	resp, err := c.client.Post(addr+path, "application/json", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	return nil
}
