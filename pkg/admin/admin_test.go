package admin

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const envoyTestHost = "127.0.0.1:9901"

var testClient = &Client{
	Host: envoyTestHost,
	Client: &http.Client{
		Timeout: 1 * time.Second,
	},
}

func TestConfigDump(t *testing.T) {
	configDump, err := testClient.ConfigDump()
	require.NoError(t, err)
	require.NotNil(t, configDump)
}

func TestServerInfo(t *testing.T) {
	serverInfo, err := testClient.ServerInfo()
	require.NoError(t, err)
	require.NotNil(t, serverInfo)
}

func TestClusters(t *testing.T) {
	clusters, err := testClient.Clusters()
	require.NoError(t, err)
	require.NotNil(t, clusters)
}

func TestQuitQuitQuit(t *testing.T) {
	err := testClient.QuitQuitQuit()
	require.NoError(t, err)
}
