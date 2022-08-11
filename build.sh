#!/bin/bash

BUILD_IMG=golang:1.18
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [ -z "${IN_DOCKER}" ]
then
    # We are not in docker, so we need to run the build script in docker.
    docker run --rm --env "IN_DOCKER=1"  -v "${SCRIPT_DIR}:/goalgotrade:rw" ${BUILD_IMG} /goalgotrade/build.sh $1
else
    # We are in docker, so we can run the build script directly.
    go install mvdan.cc/gofumpt@latest
    go install golang.org/x/tools/cmd/goimports@latest

    pushd /goalgotrade
    make $1
    popd
fi
