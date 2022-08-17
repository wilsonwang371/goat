#!/bin/bash

set -e

BUILD_IMG=golang:1.18
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [ -z "${IN_DOCKER}" ]
then
    # We are not in docker, so we need to run the build script in docker.
    docker run --rm --env "IN_DOCKER=1"  -v "${SCRIPT_DIR}:/goat:rw" ${BUILD_IMG} /goat/run-build.sh $1
else
    # We are in docker, so we can run the build script directly.
    go install mvdan.cc/gofumpt@latest
    go install golang.org/x/tools/cmd/goimports@latest
    apt update && apt install -qy npm
    npm install -g prettier@latest

    pushd /goat
    make $1
    popd
fi
