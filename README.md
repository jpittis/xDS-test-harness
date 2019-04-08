This is an experiment in testing the correctness of an arbitrary Envoy xDS server. The
idea is that the user provides a thin shim API in front of their xDS server which allows
the suite of tests in this repo to be used against it. I'm writing the tests based on the
production ready `go-control-plane` implementation but I intend to then use to tests to
validate the correctness of my in progress `rust-control-plane` implementation.

## Limitations

- I spin up an Envoy per test (and likely multiple Envoys for more complex tests in the
  future). This is really slow. We're talking ~7 seconds per simple test on my development
  machine. Maybe parallelization is the answer or an way to force envoy to reset
  configuration without restarting the process.

## Work in Progress

- There's a simple proof of concept test in `test/harness_test.go`. The APIs it uses need
  to be factored out to make writing future tests easy.


- The control plane shim is based on a simple JSON / HTTP protocol. Using gRPC is probably
  a better idea because I need to implement an identical shim server in Rust.
