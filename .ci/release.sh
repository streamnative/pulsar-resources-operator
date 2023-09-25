#!/usr/bin/env bash
# Copyright 2023 StreamNative
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


set -ex

BINDIR=`dirname "$0"`
CHARTS_HOME=`cd ${BINDIR}/..;pwd`
CHARTS_PKGS=${CHARTS_HOME}/.chart-packages
CHARTS_REPO_PATH=`cd ${CHARTS_HOME}/../charts;pwd`
CHARTS_INDEX=${CHARTS_REPO_PATH}/.chart-index
# Deprecated. you can find these args in cr.yaml
# CHARTS_REPO=${CHARTS_REPO:-"https://charts.streamnative.io"}
OWNER=${OWNER:-streamnative}
REPO=${REPO:-pulsar-resources-operator}
GITHUB_TOKEN=${GITHUB_TOKEN:-"UNSET"}
GITUSER=${GITUSER:-"UNSET"}
GITEMAIL=${GITEMAIL:-"UNSET"}
CHART="$1"
RELEASE_BRANCH="$2"

# hack/common.sh need this variable to be set
PULSAR_CHART_HOME=${CHARTS_HOME}
source ${CHARTS_HOME}/hack/common.sh

# allow overwriting cr binary
CR="${CR_BIN} --config ${CHARTS_HOME}/cr.yaml"

function release::ensure_dir() {
    local dir=$1
    if [[ -d ${dir} ]]; then
        rm -rf ${dir}
    fi
    mkdir -p ${dir}
}


function release::package_chart() {
    echo "Packaging chart '${CHART}'..."
    ${CR} package ${CHARTS_HOME}/charts/${CHART}
}

function release::upload_packages() {
    echo "Uploading charts..."
    ${CR} upload -t ${GITHUB_TOKEN} --commit ${RELEASE_BRANCH}
}





# install cr
hack::ensure_cr

release::ensure_dir ${CHARTS_PKGS}
release::ensure_dir ${CHARTS_INDEX}

release::package_chart

release::upload_packages
