#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

params=$(cat ${SCRIPT_DIR}/params.txt)

for i in ${params}; do
    read -p "Enter $i: " $i
    eval "export $i=\${$i}"
done

make test
