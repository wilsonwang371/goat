#!/bin/bash

set -xe

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

if [ "$(which docker-buildx)" == "" ]; then
    BUILDX_CMD=docker buildx
else
    BUILDX_CMD=docker-buildx
fi

${BUILDX_CMD} create --use --name goat-builder

docker run --privileged --rm tonistiigi/binfmt --install all

${BUILDX_CMD} build \
    --cache-to=type=local,dest=./  \
    --platform=linux/amd64,linux/arm64 \
    --tag=wilsonny/goat-build:latest -f ${SCRIPT_DIR}/build.dockerfile ${SCRIPT_DIR}/..

if [ ! -f ${SCRIPT_DIR}/goat-arm64 ] || [ ! -f ${SCRIPT_DIR}/goat-amd64 ] ; then
    ${SCRIPT_DIR}/../run-build.sh compile
fi

${BUILDX_CMD} build \
    --cache-to=type=local,dest=./ \
    --platform=linux/amd64,linux/arm64 \
    --tag=wilsonny/goat-release:latest -f ${SCRIPT_DIR}/release.dockerfile ${SCRIPT_DIR}/..

${BUILDX_CMD} rm goat-builder
