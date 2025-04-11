#!/usr/bin/env bash

# bash safe mode. look at `set --help` to see what these are doing
set -euxo pipefail

# Only run make reload-setup if VIAM_RELOAD is "1"
if [ "${VIAM_RELOAD:-0}" = "reinstall" ]; then
    make reload-setup
    /usr/local/go/bin/go mod tidy
fi

exec /usr/local/go/bin/go run main.go $@