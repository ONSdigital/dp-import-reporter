#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-import-reporter
  make audit
popd