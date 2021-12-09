#!/usr/bin/env bash
# Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

#!/usr/bin/env bash
# exit immediately when a command fails

set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

BINDIR=$(dirname "$0")
export POP_HOME=`cd $BINDIR/..;pwd`

if [ ! -f ${POP_HOME}/bin/golangci-lint ]; then
    cd ${POP_HOME}
    wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s v1.43.0
    cd -
fi
${POP_HOME}/bin/golangci-lint --version
${POP_HOME}/bin/golangci-lint run -c ${POP_HOME}/golangci.yml $@
