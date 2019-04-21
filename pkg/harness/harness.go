package harness

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	v2alpha "github.com/envoyproxy/go-control-plane/envoy/admin/v2alpha"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/jpittis/xDS-test-harness/pkg/admin"
	"github.com/jpittis/xDS-test-harness/pkg/shim"
)

const (
	DefaultTimeout = time.Second
	PollInterval   = time.Second
	RestartTimeout = 10 * time.Second
)

var ErrTimeout = errors.New("timeout")

type Handle struct {
	Shim  *shim.Client
	Admin *admin.Client
}

func NewHandle(shimHost, adminHost string) *Handle {
	return &Handle{
		Shim: &shim.Client{
			Host: shimHost,
			HTTPClient: &http.Client{
				Timeout: DefaultTimeout,
			},
		},
		Admin: &admin.Client{
			Host: adminHost,
			HTTPClient: &http.Client{
				Timeout: DefaultTimeout,
			},
		},
	}
}

func (h *Handle) WaitConfigDump(f func(*admin.ConfigDump) bool, timeout time.Duration) error {
	start := time.Now()

	for {
		configDump, err := h.Admin.ConfigDump()
		if err != nil {
			return err
		}

		done := f(configDump)
		if done {
			return nil
		} else if start.Add(timeout).Before(time.Now()) {
			return ErrTimeout
		}

		time.Sleep(PollInterval)
	}
}

func (h *Handle) WaitRestart(timeout time.Duration) error {
	start := time.Now()

	originalEpoch, err := h.restart_epoch()
	if err != nil {
		return err
	}

	err = h.HotRestart()
	if err != nil {
		return err
	}

	for {
		currentEpoch, err := h.restart_epoch()
		if err != nil {
			return err
		}

		fmt.Println(currentEpoch, originalEpoch)

		if currentEpoch > originalEpoch {
			return nil
		} else if start.Add(timeout).Before(time.Now()) {
			return ErrTimeout
		}

		time.Sleep(PollInterval)
	}
}

func (h *Handle) uptime() (time.Duration, error) {
	serverInfo, err := h.Admin.ServerInfo()
	if err != nil {
		return time.Duration(0), err
	}
	return types.DurationFromProto(serverInfo.UptimeAllEpochs)
}

func (h *Handle) restart_epoch() (uint32, error) {
	data, err := exec.Command("curl", "-s", "localhost:9901/server_info").Output()
	if err != nil {
		return 0, err
	}
	var serverInfo v2alpha.ServerInfo
	err = jsonpb.Unmarshal(bytes.NewBuffer(data), &serverInfo)
	if err != nil {
		return 0, err
	}

	// TODO: For some reason this client is talking to the old server.
	// serverInfo, err := h.Admin.ServerInfo()
	// if err != nil {
	// 	return 0, err
	// }

	return serverInfo.CommandLineOptions.RestartEpoch, nil
}

func (h *Handle) WithFreshEnvoy(f func(h *Handle) error) error {
	log.Println("starting xDS server")
	_, err := h.Shim.StartServer()
	if err != nil {
		return err
	}

	err = f(h)
	if err != nil {
		return err
	}

	log.Println("stopping xDS server")
	_, err = h.Shim.StopServer()
	if err != nil {
		return err
	}

	log.Println("restarting envoy")
	err = h.WaitRestart(RestartTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handle) HotRestart() error {
	resp, err := http.Post("http://127.0.0.1:3678", "", nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unknown status: %d", resp.StatusCode)
	}
	return nil
}
