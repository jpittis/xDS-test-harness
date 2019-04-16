package admin

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	v2alpha "github.com/envoyproxy/go-control-plane/envoy/admin/v2alpha"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
)

type Client struct {
	Host       string
	HTTPClient *http.Client
}

func (c *Client) ConfigDump() (*ConfigDump, error) {
	data, err := c.getJSON("/config_dump")
	if err != nil {
		return nil, err
	}

	var configDump v2alpha.ConfigDump
	err = jsonpb.Unmarshal(bytes.NewBuffer(data), &configDump)

	parsed, err := parseAnyConfigs(&configDump)
	return parsed, err
}

// We need to define a custom wrapper ConfigDump that holds the parsed Configs protobuf
// any values into their respective protobuf messages.
type ConfigDump struct {
	BootstrapConfigDump *v2alpha.BootstrapConfigDump
	ClustersConfigDump  *v2alpha.ClustersConfigDump
	ListenersConfigDump *v2alpha.ListenersConfigDump
	RoutesConfigDump    *v2alpha.RoutesConfigDump
}

func parseAnyConfigs(configDump *v2alpha.ConfigDump) (*ConfigDump, error) {
	result := &ConfigDump{}

	result.BootstrapConfigDump = &v2alpha.BootstrapConfigDump{}
	err := types.UnmarshalAny(&configDump.Configs[0], result.BootstrapConfigDump)
	if err != nil {
		return result, err
	}

	result.ClustersConfigDump = &v2alpha.ClustersConfigDump{}
	err = types.UnmarshalAny(&configDump.Configs[1], result.ClustersConfigDump)
	if err != nil {
		return result, err
	}

	result.ListenersConfigDump = &v2alpha.ListenersConfigDump{}
	err = types.UnmarshalAny(&configDump.Configs[2], result.ListenersConfigDump)
	if err != nil {
		return result, err
	}

	// Looks like the routes config dump isn't always present.
	if len(configDump.Configs) >= 4 {
		result.RoutesConfigDump = &v2alpha.RoutesConfigDump{}
		err = types.UnmarshalAny(&configDump.Configs[3], result.RoutesConfigDump)
		if err != nil {
			return result, err
		}
	}

	return result, err
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
	resp, err := c.HTTPClient.Get(c.url(path))
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
	resp, err := c.HTTPClient.Post(c.url(path), "application/json", body)
	return resp.StatusCode, err
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("http://%s%s", c.Host, path)
}
