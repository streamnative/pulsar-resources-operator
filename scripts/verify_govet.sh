#!/usr/bin/env bash
# Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

go vet ./...