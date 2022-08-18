#!/bin/bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

find "${SCRIPT_DIR}/.." -type f -name "*.sh" | xargs shfmt -w
