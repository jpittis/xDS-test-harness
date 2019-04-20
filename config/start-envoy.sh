#!/bin/bash
envoy -c /etc/envoy-proxy-config.yaml \
  --restart-epoch $RESTART_EPOCH \
  --service-cluster cluster1 --service-node node1

