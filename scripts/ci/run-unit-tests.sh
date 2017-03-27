#!/bin/bash

set -v

dir="$(dirname "$0")"
. "${dir}/install-go-deps.sh"

cd ${GOPATH}/src/github.com/skydive-project/skydive
./coverage.sh --xml --no-functionals | go2xunit -output $WORKSPACE/tests.xml
