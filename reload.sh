#!/usr/bin/env bash

# bash safe mode. look at `set --help` to see what these are doing
set -euxo pipefail

# make reload-setup

/usr/local/go/bin/go mod tidy

exec /usr/local/go/bin/go run main.go $@