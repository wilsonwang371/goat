#!/bin/bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

params=$(cat ${SCRIPT_DIR}/params.txt)

for i in ${params}; do
	# shellcheck disable=SC2229
	read -p "Enter $i: " "${i}"
	if [ -z "${!i}" ]; then
		echo "skipping $i"
		continue # skip if empty
	fi
	eval "export $i=\${$i}"
done

make test
