package test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const shimAddr = "http://127.0.0.1:4678"
const envoyAddr = "http://127.0.0.1:9901"

func TestWorking(t *testing.T) {
	client := NewShimClient(shimAddr, envoyAddr)

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
		for i := 0; i < 10; i++ {
			data, err := client.ConfigDump()
			if err == nil && bytes.Contains(data, []byte("some_service")) {
				success = true
				break

			}
			t.Logf("Polling %d", i)
			time.Sleep(time.Second)
		}
		require.True(t, success)

		t.Log("Stopping server")
		err = client.StopServer()
		require.NoError(t, err)

		t.Log("Quit quit quit")
		err = client.QuitQuitQuit()
		require.NoError(t, err)
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

func (c *ShimClient) QuitQuitQuit() error {
	return c.post(c.envoyAddr, "/quitquitquit", nil)
}

func (c *ShimClient) ConfigDump() ([]byte, error) {
	return c.get(c.envoyAddr, "/config_dump")
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

func (c *ShimClient) get(addr string, path string) ([]byte, error) {
	resp, err := c.client.Get(addr + path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
