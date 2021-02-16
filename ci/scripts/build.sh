#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-healthcheck
  make build
popd