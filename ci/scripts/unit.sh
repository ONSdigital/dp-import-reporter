#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-import-reporter
  make test
popd