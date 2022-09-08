#!/bin/bash

set -xe

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

docker buildx create --use --name goat-builder 

docker buildx build \
    --push \
    --platform=linux/amd64,linux/arm64 \
    --tag=wilsonny/goat-build:latest -f ${SCRIPT_DIR}/build.dockerfile ${SCRIPT_DIR}/..

if [ ! -f ${SCRIPT_DIR}/goat-arm64 ] || [ ! -f ${SCRIPT_DIR}/goat-amd64 ] ; then
    ${SCRIPT_DIR}/build-docker.sh compile
fi

docker buildx build \
    --push \
    --platform=linux/amd64,linux/arm64 \
    --tag=wilsonny/goat-release:latest -f ${SCRIPT_DIR}/release.dockerfile ${SCRIPT_DIR}/..

docker buildx rm goat-builder
