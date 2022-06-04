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


set -e

BINDIR=`dirname "$0"`
CHARTS_HOME=`cd ${BINDIR}/..;pwd`
CHARTS_PKGS=${CHARTS_HOME}/.chart-packages
CHARTS_INDEX=${CHARTS_HOME}/.chart-index
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

function release::update_chart_index() {
    echo "Updating chart index..."
    ${CR} index -t ${GITHUB_TOKEN}
}

function release::publish_charts() {
    git config user.email "${GITEMAIL}"
    git config user.name "${GITUSER}"
    git pull
    git checkout gh-pages
    cp --force ${CHARTS_INDEX}/index.yaml index.yaml
    git add index.yaml
    git commit --message="Publish new charts v${RELEASE_BRANCH:7}" --signoff
    git remote -v
    git remote add sn https://${SNBOT_USER}:${GITHUB_TOKEN}@github.com/${OWNER}/${REPO} 
    git push sn gh-pages 
}

# install cr
hack::ensure_cr

release::ensure_dir ${CHARTS_PKGS}
release::ensure_dir ${CHARTS_INDEX}

release::package_chart

release::upload_packages
release::update_chart_index

release::publish_charts
