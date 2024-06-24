#!/usr/bin/env bash
# Copyright 2022 StreamNative
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


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
    wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s v1.55.2
    cd -
fi
${POP_HOME}/bin/golangci-lint --version
${POP_HOME}/bin/golangci-lint run -c ${POP_HOME}/.golangci.yml $@
