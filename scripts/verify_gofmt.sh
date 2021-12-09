#!/usr/bin/env bash
# Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

goFiles=$(find . -name \*.go -not -path "*/vendor/*" -print)
invalidFiles=$(gofmt -l $goFiles)

if [ "$invalidFiles" ]; then
  echo -e "These files did not pass the 'go fmt' check, please run 'go fmt' on them:"
  echo -e $invalidFiles
  exit 1
fi
