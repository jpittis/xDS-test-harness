package harness

import (
	"errors"
	"log"
	"net/http"
	"os/exec"
	"time"

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

	originalUptime, err := h.uptime()
	if err != nil {
		return err
	}

	// TODO: Replace QuitQuitQuit with
	// err = h.HotRestart()
	err = h.Admin.QuitQuitQuit()
	if err != nil {
		return err
	}

	for {
		// Ignore the error because Envoy is restarting and may not be listening for admin
		// requests.
		currentUptime, _ := h.uptime()

		if currentUptime < originalUptime {
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
	// TODO: Add a daemon to the envoy pod that listens for an HTTP call and then triggers
	// a sighup.
	return exec.Command("config/sighup.sh").Run()
}
