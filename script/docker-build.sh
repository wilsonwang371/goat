#!/bin/bash -x

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [ -z "${IN_DOCKER}" ]
then
    # We are not in docker, so we need to run the build script in docker.
    docker run --rm --env "IN_DOCKER=1"  -v "${SCRIPT_DIR}/..:/goalgotrade:rw" golang:1.16 /goalgotrade/docker-build.sh
else
    # We are in docker, so we can run the build script directly.
    pushd ${SCRIPT_DIR}
    make
    popd
fi
