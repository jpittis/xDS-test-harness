package admin

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	v2alpha "github.com/envoyproxy/go-control-plane/envoy/admin/v2alpha"
	"github.com/gogo/protobuf/jsonpb"
)

type Client struct {
	Host   string
	Client *http.Client
}

func (c *Client) ConfigDump() (*v2alpha.ConfigDump, error) {
	data, err := c.getJSON("/config_dump")
	if err != nil {
		return nil, err
	}

	var configDump v2alpha.ConfigDump
	err = jsonpb.Unmarshal(bytes.NewBuffer(data), &configDump)
	return &configDump, err
}

func (c *Client) Clusters() (*v2alpha.Clusters, error) {
	data, err := c.getJSON("/clusters")
	if err != nil {
		return nil, err
	}

	var clusters v2alpha.Clusters
	err = jsonpb.Unmarshal(bytes.NewBuffer(data), &clusters)
	return &clusters, err
}

func (c *Client) ServerInfo() (*v2alpha.ServerInfo, error) {
	data, err := c.getJSON("/server_info")
	if err != nil {
		return nil, err
	}

	var serverInfo v2alpha.ServerInfo
	err = jsonpb.Unmarshal(bytes.NewBuffer(data), &serverInfo)
	return &serverInfo, err
}

func (c *Client) QuitQuitQuit() error {
	_, err := c.post("/quitquitquit", nil)
	return err
}

func (c *Client) getJSON(path string) ([]byte, error) {
	return c.get(path + "?format=json")
}

func (c *Client) get(path string) ([]byte, error) {
	resp, err := c.Client.Get(c.url(path))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Client) post(path string, body io.Reader) (int, error) {
	resp, err := c.Client.Post(c.url(path), "application/json", body)
	return resp.StatusCode, err
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("http://%s%s", c.Host, path)
}
