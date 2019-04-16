package shim

import (
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	Host       string
	HTTPClient *http.Client
}

func (c *Client) StartServer() (int, error) {
	return c.post("/start_server", nil)
}

func (c *Client) StopServer() (int, error) {
	return c.post("/stop_server", nil)
}

func (c *Client) SetSnapshot() (int, error) {
	return c.post("/set_snapshot", nil)
}

func (c *Client) post(path string, body io.Reader) (int, error) {
	resp, err := c.HTTPClient.Post(c.url(path), "application/json", body)
	return resp.StatusCode, err
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("http://%s%s", c.Host, path)
}
