This is an experiment in testing the correctness of an arbitrary Envoy xDS server. The
idea is that the user provides a thin shim API in front of their xDS server which allows
the suite of tests in this repo to be used against it. I'm writing the tests based on the
production ready `go-control-plane` but intend to then use the tests to validate the
correctness of my in progress `rust-control-plane`.

## Example

Right now I just have a simple proof of concept test.

```go
const (
	defaultTimeout = 10 * time.Second
	shimTestHost   = "127.0.0.1:4678"
	envoyTestHost  = "127.0.0.1:9901"
)

func TestWorking(t *testing.T) {
	h := harness.NewHandle(shimTestHost, envoyTestHost)

	err := h.WithFreshEnvoy(func(h *harness.Handle) error {
		// This sends a hardcoded snapshot with a single "some_service" cluster.
		_, err := h.Shim.SetSnapshot()
		if err != nil {
			return err
		}

		return h.WaitConfigDump(func(configDump *admin.ConfigDump) bool {
			dynamicActiveClusters := configDump.ClustersConfigDump.DynamicActiveClusters
			return len(dynamicActiveClusters) > 0 &&
				dynamicActiveClusters[0].Cluster.Name == "some_service"
		}, defaultTimeout)
	})

	require.NoError(t, err)
}
```

## Development

```
$ docker-compose up -d
$ cd test
$ go test
```
