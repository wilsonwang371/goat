#!/bin/bash

set -e

BUILD_IMG=wilsonny/goat-build:latest
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

if [ -z "${IN_DOCKER}" ]; then
	# We are not in docker, so we need to run the build script in docker.
	docker run --rm --env "IN_DOCKER=1" -v "${SCRIPT_DIR}:/goat:rw" ${BUILD_IMG} /goat/run-build.sh $1
	if [ $1 = "compile" ]; then
		pushd docker
		docker-compose build release-img
		popd
	fi
else
	# We are in docker, so we can run the build script directly.
	pushd /goat
	make "$1"
	popd
fi
