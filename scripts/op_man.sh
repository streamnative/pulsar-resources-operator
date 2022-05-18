#!/usr/bin/env bash
# Copyright 2022 Stream Native
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


set -euo pipefail

# PROJECT_ID and GROUP_NAME could be customized for a private build
# PROJECT_ID is the id of the google cloud project
# GROUP_NAME is the logic name of images, e.g., the developer's name, and the default value is streamnative
SEP="/"
IMAGE_PREFIX="gcr.io/${PROJECT_ID:=affable-ray-226821}/${GROUP_NAME:=streamnative}${SEP}pulsar-operators"
OPERATOR_IMAGE_PREFIX="${IMAGE_PREFIX}${SEP}operator"
BUNDLE_IMAGE_PREFIX="${IMAGE_PREFIX}${SEP}bundle"
REGISTRY_IMAGE_PREFIX="${IMAGE_PREFIX}${SEP}registry"
UNIFIED_MANIFEST_NAME="pulsar-operators"

BINDIR=$(dirname "$0")
POP_HOME=$(cd $BINDIR/..;pwd)
CONC_JOB_NUM=8

SCRIPT_BRANCH=${SCRIPT_BRANCH:-master}
SCRIPT_FILES="scripts/op_man.sh Makefile bookkeeper-operator/Makefile pulsar-operator/Makefile zookeeper-operator/Makefile"

function help() {
    echo "Tools to ease bundling management.

Usage:
  $(basename $0) [command]

Available Commands:
  generate-csv
  create-bundle
  index-bundle
  build-operator
"
}

function git_rev() {
  git rev-parse --short HEAD
}

function git_branch() {
  NAME="$(git rev-parse --abbrev-ref HEAD)"
  if [ "$NAME" = "HEAD" ]
  then
    echo ''
  else
    echo "$NAME"
  fi
}

function git_tag() {
  git describe --exact-match 2> /dev/null || echo ''
}

function git_name() {
  NAME="$(git_tag)"
  if [ -n "$NAME" ]
  then
    echo "$NAME"
    exit
  fi
  NAME="$(git_branch)"
  if [ -n "$NAME" ]
  then
    echo "$NAME"
    exit
  fi
  git_rev
}

function operator_image_name() {
  echo "$OPERATOR_IMAGE_PREFIX$SEP$1"
}

function bundle_image_name() {
  echo "$BUNDLE_IMAGE_PREFIX$SEP$1"
}

function registry_image_name() {
  echo "$REGISTRY_IMAGE_PREFIX$SEP$1"
}

function prepare_yq() {
  RELEASE_VERSION=v4.9.8

  if [ -z ${YQ+x} ]
  then
    if [ -x "$(command -v yq)" ]
    then
      YQ=$(command -v yq)
    else
      if [ ! -d "${POP_HOME}/bin" ]
      then
        mkdir "${POP_HOME}/bin"
      fi

      cd "${POP_HOME}/bin"
      if [ ! -x yq ]
      then
        wget -q "https://github.com/mikefarah/yq/releases/download/$RELEASE_VERSION/yq_$(uname)_amd64" -O yq
        chmod +x yq
      fi
      cd -
      YQ="${POP_HOME}/bin/yq"
    fi
  fi
  export YQ
}

function prepare_sdk() {
  RELEASE_VERSION='v1.8.0'

  if [ -z ${OPERATOR_SDK+x} ];
  then
    if [ -x "${POP_HOME}/bin/operator-sdk" ]
    then
      OPERATOR_SDK="${POP_HOME}/bin/operator-sdk"
    elif [ -x "$(command -v operator-sdk)" ];
    then
      OPERATOR_SDK="$(command -v operator-sdk)"
    else
      if [ ! -d "${POP_HOME}/bin" ]
      then
        mkdir "${POP_HOME}/bin"
      fi

      cd "${POP_HOME}/bin"
      if [ ! -x operator-sdk ]
      then
        if [ 'Darwin' = "$(uname)" ]
        then
          wget -q "https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk_darwin_amd64" -O operator-sdk
        else
          wget -q "https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk_linux_amd64" -O operator-sdk
        fi
        chmod +x operator-sdk
      fi
      cd -
      OPERATOR_SDK="${POP_HOME}/bin/operator-sdk"
    fi
  fi
  export OPERATOR_SDK
}

function prepare_opm() {
  RELEASE_VERSION='1.13.7'
  if [ -z ${OPM+x} ];
  then
    if [ -x "${POP_HOME}/bin/opm" ]
    then
      OPM="${POP_HOME}/bin/opm"
    elif [ -x "$(command -v opm)" ];
    then
      OPM="$(command -v opm)"
    else
      if [ ! -d "${POP_HOME}/bin" ]
      then
        mkdir "${POP_HOME}/bin"
      fi

      cd "${POP_HOME}/bin"
      if [ ! -x opm ]
      then
        GOOS=$(go env GOOS)
        GOARCH=$(go env GOARCH)
        wget -q "https://github.com/operator-framework/operator-registry/releases/download/v$RELEASE_VERSION/$GOOS-$GOARCH-opm" -O opm
        chmod +x opm
      fi
      cd -
      OPM="${POP_HOME}/bin/opm"
    fi
  fi
  export OPM
}

function prepare_kubebuilder() {
  os=$(go env GOOS)
  arch=$(go env GOARCH)
  KUBEBUILDER_ASSETS="$POP_HOME/bin/kubebuilder/bin"
  K8S_VERSION=1.19.2
  if [ ! -d "$KUBEBUILDER_ASSETS" ]
  then
    mkdir -p "$POP_HOME/bin"
    wget "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${K8S_VERSION}-$(go env GOOS)-$(go env GOARCH).tar.gz" -O envtest-bins.tar.gz
    tar xzf envtest-bins.tar.gz
    mv "kubebuilder" "$POP_HOME/bin/kubebuilder"
    wget "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.1.0/kubebuilder_${os}_${arch}" -O kubebuilder
    chmod +x kubebuilder
    mv "kubebuilder" "$POP_HOME/bin/kubebuilder/bin"
  fi
  export KUBEBUILDER_ASSETS
  export PATH="$PATH:$KUBEBUILDER_ASSETS"
}

function install_olm() {
  RELEASE_VERSION=0.14.1
  kubectl apply -f "https://github.com/operator-framework/operator-lifecycle-manager/releases/download/$RELEASE_VERSION/crds.yaml"
  kubectl apply -f "https://github.com/operator-framework/operator-lifecycle-manager/releases/download/$RELEASE_VERSION/olm.yaml"
}

function head_csv_version() {
  prepare_opm
  prepare_yq

  PACKAGE=$(basename "$PWD")
  CHANNEL='alpha'

  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --package)
      PACKAGE=$1
      shift
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  if [ -z "$CHANNEL" ]
  then
    echo "--channel is required!"
    exit 1
  fi

  TMP_DIR="$(mktemp -d)"
  echo "Use tmp dir $TMP_DIR"
  cd "$TMP_DIR"
  "$OPM" index export --index="$OPERATOR_INDEX_IMAGE" -c docker --package "$PACKAGE"
  cat downloaded/package.yaml
  export HEAD_CSV_VERSION="$(${YQ} r downloaded/package.yaml 'channels.(name==alpha).currentCSV')"
  cd -
  rm -rf "$TMP_DIR"
  export HEAD_CSV_SEMVER=${HEAD_CSV_VERSION//$PACKAGE\.v/}
  echo "Got head version for $PACKAGE: $HEAD_CSV_VERSION, $HEAD_CSV_SEMVER"
}

function current_version() {
  prepare_yq

  PACKAGE=$(basename "$PWD")
  CHANNEL=alpha
  cd "deploy/olm-catalog/$PACKAGE"
  export CURRENT_VERSION=$($YQ r "$PACKAGE".package.yaml 'channels.(name=='$CHANNEL').currentCSV')
  echo "$CURRENT_VERSION"
  cd -
}

function generate_csv() {
  prepare_sdk
  prepare_yq

  FROM_VERSION=''
  VERSION='0.0.1'
  PACKAGE=$(basename "$PWD")
  CHANNEL='alpha'
  SNAPSHOT=''

  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --version)
      VERSION=$1
      shift
      ;;
    --from-version)
      FROM_VERSION=$1
      shift
      ;;
    --channel)
      CHANNEL=$1
      shift
      ;;
    --snapshot)
      SNAPSHOT='1'
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  make bundle-manifests

  if [ "$SNAPSHOT" ]
  then
    cp "deploy/olm-catalog/$PACKAGE/$PACKAGE.package.yaml" "$PACKAGE.package.yaml.bak"
  fi

  "$OPERATOR_SDK" generate csv \
    --update-crds \
    --csv-channel="$CHANNEL" \
    --csv-version="$VERSION" \
    --from-version="$FROM_VERSION"

  if [ "$SNAPSHOT" ]
  then
    mv "$PACKAGE.package.yaml.bak" "deploy/olm-catalog/$PACKAGE/$PACKAGE.package.yaml"
  fi

  cd "deploy/olm-catalog/$PACKAGE/$VERSION"

  FILENAME=$(ls *clusterserviceversion*)

  IMAGE_NAME="$(operator_image_name "$PACKAGE"):v$VERSION"
  "$YQ" e ".spec.install.spec.deployments[0].spec.template.spec.containers[0].image = \"$IMAGE_NAME\"" -i "$FILENAME"
  if [ -z "$FROM_VERSION" ] || [ "$SNAPSHOT" ]
  then
    "$YQ" d -i "$FILENAME" spec.replaces
  fi

  if [ "$SNAPSHOT" ]
  then
    "$YQ" e ".metadata.annotations.\"olm.skipRange\" = \"<$VERSION\"" -i "$FILENAME"
  fi

  cd -
}

function create_bundle() {
  prepare_sdk
  VERSION=''
  PACKAGE=$(basename "$PWD")
  CHANNELS=alpha
  PUSH=''

  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --version)
      VERSION=$1
      shift
      ;;
    --package)
      PACKAGE=$1
      shift
      ;;
    --channels)
      CHANNELS=$1
      shift
      ;;
    --default-channel)
      DEFAULT_CHANNEL=$1
      shift
      ;;
    --push)
      PUSH='1'
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  if [ -z "$VERSION" ]
  then
    echo "--version is required!"
    exit 1
  fi

  OPERATOR_BUNDLE_IMAGE="$(bundle_image_name "$PACKAGE")"
  OPERATOR_BUNDLE_TAG="v$VERSION"
  export OPERATOR_BUNDLE_IMAGE
  export OPERATOR_BUNDLE_TAG
  "$OPERATOR_SDK" bundle create \
    --directory "./deploy/olm-catalog/${PACKAGE}/${VERSION}" \
    --channels "$CHANNELS" \
    "$OPERATOR_BUNDLE_IMAGE:$OPERATOR_BUNDLE_TAG"

  docker tag "$OPERATOR_BUNDLE_IMAGE:$OPERATOR_BUNDLE_TAG" "$OPERATOR_BUNDLE_IMAGE:$(git_rev)"

  if [ "$PUSH" ]
  then
    docker push "$OPERATOR_BUNDLE_IMAGE:$OPERATOR_BUNDLE_TAG"
    docker push "$OPERATOR_BUNDLE_IMAGE:$(git_rev)"
  fi
}

function index_bundles() {
  prepare_opm

  ARGS=(index add -c docker --bundles "$BUNDLES" --tag "$OPERATOR_INDEX_IMAGE")

  if [ "$FROM_INDEX_VERSION" ]
  then
    ARGS+=(--from-index "${REGISTRY_IMAGE_NAME}:${FROM_INDEX_VERSION}")
  fi
  # shellcheck disable=SC2086
  "$OPM" ${ARGS[*]}
}

function index_bundle() {
  VERSION=''
  PACKAGE=$(basename "$PWD")
  TARGET_INDEX_VERSION="latest"
  FROM_INDEX_VERSION=""
  MANIFEST_NAME="$PACKAGE"
  PUSH=''

  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --version)
      VERSION=$1
      shift
      ;;
    --package)
      PACKAGE=$1
      shift
      ;;
    --manifest)
      MANIFEST_NAME=$1
      shift
      ;;
    --from-index-version)
      FROM_INDEX_VERSION=$1
      shift
      ;;
    --target-index-version)
      TARGET_INDEX_VERSION=$1
      shift
      ;;
    --push)
      PUSH='1'
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  if [ -z "$VERSION" ]
  then
    echo "--version is required!"
    exit 1
  fi

  REGISTRY_IMAGE_NAME="$(registry_image_name "$MANIFEST_NAME")"
  OPERATOR_INDEX_IMAGE="$REGISTRY_IMAGE_NAME:$TARGET_INDEX_VERSION"
  OPERATOR_INDEX_IMAGE_VERSIONED="${REGISTRY_IMAGE_NAME}:$(snapshot_version "$TARGET_INDEX_VERSION")"
  BUNDLES="$(bundle_image_name "$PACKAGE"):v$VERSION"
  index_bundles
  docker tag "$OPERATOR_INDEX_IMAGE" "$OPERATOR_INDEX_IMAGE_VERSIONED"

  if [ "$PUSH" ]
  then
    docker push "$REGISTRY_IMAGE_NAME:$TARGET_INDEX_VERSION"
    docker push "$OPERATOR_INDEX_IMAGE_VERSIONED"
  fi

  export REGISTRY_IMAGE_NAME
  export OPERATOR_INDEX_IMAGE
  export OPERATOR_INDEX_IMAGE_VERSIONED
}

function snapshot_version() {
  echo "$1-$(date -u +%Y%m%d%H%M%S)-$(git rev-parse --short HEAD)"
}

function build_operator() {
  prepare_sdk
  prepare_yq

  VERSION=''
  PACKAGE=$(basename "$PWD")
  VERSION_SUFFIX=''
  UPDATE_CRDS=false
  PUSH=''

  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --version)
      VERSION=$1
      shift
      ;;
    --package)
      PACKAGE=$1
      shift
      ;;
    --git-rev)
      if [ -n "${VERSION_SUFFIX}" ]
      then
        VERSION_SUFFIX="${VERSION_SUFFIX}-"
      fi
      VERSION_SUFFIX="${VERSION_SUFFIX}$(git rev-parse --short HEAD)"
      ;;
    --ust)
      if [ -n "${VERSION_SUFFIX}" ]
      then
        VERSION_SUFFIX="${VERSION_SUFFIX}-"
      fi
      VERSION_SUFFIX="${VERSION_SUFFIX}$(date -u +%Y%m%d%H%M%S)"
      ;;
    --update-crds)
      UPDATE_CRDS=true
      ;;
    --push)
      PUSH='1'
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  OPERATOR_TAG=$VERSION
  if [ -n "${VERSION_SUFFIX}" ]
  then
    if [ -n "$VERSION" ]
    then
      VERSION="${VERSION}-${VERSION_SUFFIX}"
    else
      VERSION="$VERSION_SUFFIX"
    fi
  fi
  OPERATOR_TAG="v${VERSION}"

  if [ -z "$VERSION" ]
  then
    echo "--version is required!"
    exit 1
  fi

  export OPERATOR_VERSION="$VERSION"
  OPERATOR_IMAGE="$(operator_image_name "$PACKAGE")"
  export OPERATOR_IMAGE
  export OPERATOR_TAG
  IMG="${OPERATOR_IMAGE}:$OPERATOR_TAG" make docker-build
  docker tag "${OPERATOR_IMAGE}:$OPERATOR_TAG" "${OPERATOR_IMAGE}:$(git_rev)"

  if [ "$PUSH" ]
  then
    docker push "$OPERATOR_IMAGE:$OPERATOR_TAG"
    docker push "$OPERATOR_IMAGE:$(git_rev)"
  fi
}

function fork_bundle() {
  PACKAGE=$(basename "$PWD")
  current_version
  CURRENT_SEMVER=${CURRENT_VERSION//$PACKAGE\.v/}
  VERSION=''
  SNAPSHOT=''

  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --version)
      VERSION=$1
      shift
      ;;
    --snapshot)
      VERSION=$(snapshot_version "$CURRENT_SEMVER")
      SNAPSHOT='1'
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  if [ -z "$VERSION" ]
  then
    echo "--version is required!"
    exit 1
  fi

  if [ $SNAPSHOT ]
  then
    generate_csv --version "$VERSION" --from-version "$CURRENT_SEMVER" --snapshot
  else
    generate_csv --version "$VERSION" --from-version "$CURRENT_SEMVER"
  fi

  export OPERATOR_VERSION=$VERSION
}

function update_replaces_from_head() {
  prepare_yq

  PACKAGE=$(basename "$PWD")

  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --index-tag)
      INDEX_TAG=$1
      shift
      ;;
    --version)
      VERSION=$1
      shift
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  ls "deploy/olm-catalog/$PACKAGE/$VERSION"

  REGISTRY_IMAGE_NAME=$(registry_image_name "$PACKAGE")
  OPERATOR_INDEX_IMAGE="$REGISTRY_IMAGE_NAME:$INDEX_TAG"
  head_csv_version

  cd "deploy/olm-catalog/$PACKAGE/$VERSION"

  FILENAME=$(ls *clusterserviceversion*)
  "$YQ" w -i "$FILENAME" spec.replaces "$HEAD_CSV_VERSION"
  cd -
}

function release_bundle() {
  INITIAL=''
  FROM_HEAD=''
  INDEX_TAG="latest"
  SKIP_INDEX=''
  CHANNELS=${CHANNELS:-alpha}

  PACKAGE=$(basename "$PWD")
  current_version
  VERSION=$CURRENT_VERSION
  MANIFEST_NAME="$PACKAGE"

  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --initial)
      INITIAL=1
      ;;
    --from-head)
      FROM_HEAD=1
      ;;
    --index-tag)
      INDEX_TAG=$1
      shift
      ;;
    --version)
      VERSION=$1
      shift
      ;;
    --unified)
      MANIFEST_NAME="$UNIFIED_MANIFEST_NAME"
      ;;
    --skip-index)
      SKIP_INDEX=1
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  SEMVER=${VERSION//$PACKAGE\.v/}

  if [ "$FROM_HEAD" ]
  then
    update_replaces_from_head --version "$SEMVER" --index-tag "$INDEX_TAG"
  fi

  TARGET_INDEX_VERSION="$INDEX_TAG"
  FROM_INDEX_VERSION="$INDEX_TAG"
  if [ "$INITIAL" ] || [ "$SKIP_INDEX" ]
  then
    FROM_INDEX_VERSION=""
  fi

  echo "To release $PACKAGE, version $VERSION, semver $SEMVER"

  echo "To build and push operator image..."
  build_operator --version "$SEMVER" --push

  echo "To create and push operator bundle image..."
  create_bundle --version "$SEMVER" --channels "$CHANNELS" --push

  if [ -z "$SKIP_INDEX" ]
  then
    echo "To create and push operator manifests image..."
    index_bundle --version "$SEMVER" \
      --from-index-version "$FROM_INDEX_VERSION" \
      --target-index-version "$TARGET_INDEX_VERSION" \
      --manifest "$MANIFEST_NAME" --push
  fi
}

function release_snapshot_bundle() {
  INDEX_TAG=$(git_name)
  INDEX_TAG=${INDEX_TAG//[\/]/-}
  ARGS=$@

  fork_bundle --snapshot
  ARGS+=(--version "$OPERATOR_VERSION" --index-tag "$INDEX_TAG")
  # shellcheck disable=SC2086
  release_bundle ${ARGS[*]}
}

function runE2ETest() {
  prepare_sdk
  "${OPERATOR_SDK}" test local ./test/e2e --go-test-flags "-v -parallel=2"
}

function update-skip-range() {
  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --op)
      OP=$1
      shift
      ;;
    --version)
      VERSION=$1
      shift
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  prepare_yq
  "$YQ" e ".metadata.annotations.\"olm.skipRange\" = \"<$VERSION\""  -i "$OP/bundle/manifests/$OP.clusterserviceversion.yaml"
}

function release() {
  prepare_yq

  SKIP_OPERATOR_BUILD=${SKIP_OPERATOR_BUILD:-''}
  SKIP_BUNDLE_BUILD=${SKIP_BUNDLE_BUILD:-''}
  SKIP_REGISTRY_BUILD=${SKIP_REGISTRY_BUILD:-''}
  while [ $# -gt 0 ]
  do
    KEY=$1
    shift
    case $KEY in
    --skip-operator-build)
      SKIP_OPERATOR_BUILD=1
      ;;
    --skip-bundle-build)
      SKIP_BUNDLE_BUILD=1
      ;;
    --skip-registry-build)
      SKIP_REGISTRY_BUILD=1
      ;;
    *)
      echo "Invalid argument $KEY"
      exit 1
      ;;
    esac
  done

  VERSIONS=()
  CHANNEL_NAMES=()
  VERSION_NUM=$("$YQ" e '.config.channels | length'  registry-config.yaml)

  i=0
  while [ $i -lt $VERSION_NUM ]
  do
    echo $i
    name=$("$YQ" e ".config.channels[$i].name" registry-config.yaml)
    version=$("$YQ" e ".config.channels[$i].version" registry-config.yaml)
    echo "$name, $version"
    j=0
    is_new=1
    while [ $j -lt ${#VERSIONS[@]} ]
    do
      if [ ${VERSIONS[$j]} == $version ]
      then
        is_new=''
        CHANNEL_NAMES[$j]="${CHANNEL_NAMES[$j]},$name"
        break
      fi
      j=$(echo $j + 1|bc)
    done
    if [ "$is_new" ]
    then
      VERSIONS+=("$version")
      CHANNEL_NAMES+=("$name")
    fi
    i=$(echo $i + 1|bc)
  done
  echo ${VERSIONS[*]}
  echo ${CHANNEL_NAMES[*]}

  DEFAULT_CHANNEL=$("$YQ" e '.config.defaultChannel' registry-config.yaml)
  ADDITIVE=$("$YQ" e '.config.additive' registry-config.yaml)
  USE_REPLACE=$("$YQ" e '.config.useReplace' registry-config.yaml)
  SNAPSHOT=$("$YQ" e '.config.snapshot' registry-config.yaml)
  REGISTRY_IMAGE=$("$YQ" e '.config.registryImage' registry-config.yaml)
  REGISTRY_IMAGE_TAG=$("$YQ" e '.config.registryImageTag' registry-config.yaml)

  export DEFAULT_CHANNEL
  export ADDITIVE
  export USE_REPLACE
  if [ "$REGISTRY_IMAGE" ]
  then
    export REGISTRY_IMAGE
  fi
  if [ "$REGISTRY_IMAGE_TAG" ]
  then
    export REGISTRY_IMAGE_TAG
  fi

  if [ "$SKIP_REGISTRY_BUILD" == '' ]
  then
    ADDITIVE=true
  fi

  IMAGE_NUM=$("$YQ" e '.config.imageBases | length' registry-config.yaml)
  ii=0

  while [ $ii -lt $IMAGE_NUM ]
  do
    IMAGE_BASE=$("$YQ" e ".config.imageBases[$ii].image" registry-config.yaml)
    export IMAGE_BASE
    echo $IMAGE_BASE
    URI_SEP=$("$YQ" e ".config.imageBases[$ii].sep" registry-config.yaml)
    if [ "$URI_SEP" != "" -a "$URI_SEP" != 'null' ]
    then
      echo "Using uri separator [$URI_SEP]"
      export URI_SEP
    fi
    KUBE_RBAC_PROXY_IMG=$("$YQ" e ".config.imageBases[$ii].kubeRbacProxyImage" registry-config.yaml)
    if [ "$KUBE_RBAC_PROXY_IMG" != "" -a "$KUBE_RBAC_PROXY_IMG" != 'null' ]
    then
      echo "Using kube_rbac_proxy image [$KUBE_RBAC_PROXY_IMG]"
      export KUBE_RBAC_PROXY_IMG
    fi
    ii=$(echo $ii + 1|bc)

    if [ "$SKIP_REGISTRY_BUILD" == '' ]
    then
      # make an empty registry first
      TMP_REGISTRY_IMAGE_TAG=$(snapshot_version tmp)
      echo "Using tmp tag [$TMP_REGISTRY_IMAGE_TAG] for registry image for image base [$IMAGE_BASE]"
      REGISTRY_IMAGE_TAG="$TMP_REGISTRY_IMAGE_TAG" make empty-registry
      export FROM_REGISTRY_IMAGE_TAG="$TMP_REGISTRY_IMAGE_TAG"
      REGISTRY_IMAGE_TAG="$TMP_REGISTRY_IMAGE_TAG" make registry-push
    fi

    i=0
    while [ $i -lt ${#VERSIONS[*]} ]
    do
      VERSION=${VERSIONS[$i]}

      echo "To build version $VERSION"
      if [ "$SNAPSHOT" != 'true' ]
      then
        git stash
        for dir in `ls -d *-operator`
        do
          pushd "$dir"
          git stash
          git checkout "v$VERSION"
          popd
        done
        git checkout "$SCRIPT_BRANCH" -- $SCRIPT_FILES
      fi

      TAG="v$VERSION"
      CHANNELS=${CHANNEL_NAMES[$i]}
      BUNDLE_IMAGE_TAG="${CHANNELS//[,]/-}-$TAG"
      echo $TAG, $CHANNELS

      export VERSION
      export TAG
      export CHANNELS
      export BUNDLE_IMAGE_TAG

      if [ "$SKIP_OPERATOR_BUILD" == '' ]
      then
        make docker-build -j $CONC_JOB_NUM
        make docker-push -j $CONC_JOB_NUM
      fi
      if [ "$SKIP_BUNDLE_BUILD" == '' ]
      then
        make bundle -j $CONC_JOB_NUM
        make bundle-build -j $CONC_JOB_NUM
        make bundle-push -j $CONC_JOB_NUM
      fi

      if [ "$SKIP_REGISTRY_BUILD" == '' ]
      then
        REGISTRY_IMAGE_TAG="$TMP_REGISTRY_IMAGE_TAG" make registry
        REGISTRY_IMAGE_TAG="$TMP_REGISTRY_IMAGE_TAG" make registry-push
      fi

      i=$(echo $i + 1|bc)
    done

    if [ "$SKIP_REGISTRY_BUILD" == '' ]
    then
      make tag-registry
      make registry-push
    fi
  done
}

function release-snapshot() {
  prepare_yq
  cp templates/registry/registry-config.yaml registry-config.yaml
  VERSION=$(cat VERSION)
  VERSION="$(snapshot_version $VERSION)"
  VERSION=${VERSION//[\/]/-}
  REGISTRY_IMAGE_TAG=$(git_name)
  REGISTRY_IMAGE_TAG=${REGISTRY_IMAGE_TAG//[\/]/-}

  "$YQ" e ".config.channels[0].version = \"$VERSION\"" -i registry-config.yaml
  "$YQ" e '.config.additive = false' -i registry-config.yaml
  "$YQ" e '.config.snapshot = true' -i registry-config.yaml
  "$YQ" e ".config.registryImageTag = \"$REGISTRY_IMAGE_TAG\"" -i registry-config.yaml

  release $@
}

function build_release() {
  BUNDLE_NAMES=()
  REGISTRY_TAG='latest'
  CHANNELS=stable,beta
  for OP in *-operator
  do
    echo "To build $OP"
    cd "$OP"
    if [ ! -d deploy/olm-catalog ]
    then
      echo "Skip $OP as it has no bundles"
      cd ..
      continue
    fi

    # shellcheck disable=SC2086
    release_bundle --skip-index
    BUNDLE_NAMES+=("$OPERATOR_BUNDLE_IMAGE:$OPERATOR_BUNDLE_TAG")

    cd ..
  done

  if [ ${#BUNDLE_NAMES[*]} -gt 0 ]
  then
    BUNDLES="$(IFS=, ; echo "${BUNDLE_NAMES[*]}")"
    echo "$BUNDLES"
    REGISTRY_IMAGE_NAME=$(registry_image_name "$UNIFIED_MANIFEST_NAME")
    OPERATOR_INDEX_IMAGE="${REGISTRY_IMAGE_NAME}:${REGISTRY_TAG}"
    OPERATOR_INDEX_IMAGE_VERSIONED="${REGISTRY_IMAGE_NAME}:${REGISTRY_VERSIONED_TAG}"
    index_bundles
    docker tag "$OPERATOR_INDEX_IMAGE" "$OPERATOR_INDEX_IMAGE_VERSIONED"
    docker push "$OPERATOR_INDEX_IMAGE"
    docker push "$OPERATOR_INDEX_IMAGE_VERSIONED"

    export REGISTRY_IMAGE_NAME
    export OPERATOR_INDEX_IMAGE
    export OPERATOR_INDEX_IMAGE_VERSIONED
  fi
}

if [ $# -eq 0 ]
then
  help
  exit 1
fi

COMMAND=$1
shift

case "$COMMAND" in
git-name)
  git_name
  ;;
yq)
  prepare_yq
  ;;
sdk)
  prepare_sdk
  ;;
opm)
  prepare_opm
  ;;
kubebuilder)
  prepare_kubebuilder
  ;;
install-olm)
  install_olm
  ;;
head-csv-version)
  head_csv_version $@
  ;;
current-version)
  current_version $@
  ;;
generate-csv)
  generate_csv $@
  ;;
create-bundle)
  create_bundle $@
  ;;
fork-bundle)
  fork_bundle $@
  ;;
index-bundle)
  index_bundle $@
  ;;
build-operator)
  build_operator $@
  ;;
update-replaces-from-head)
  update_replaces_from_head $@
  ;;
release-bundle)
  release_bundle $@
  ;;
release-snapshot-bundle)
  release_snapshot_bundle $@
  ;;
build-release)
  build_release $@
  ;;
build-snapshot)
  release-snapshot $@
  ;;
e2e-test)
  runE2ETest $@
  ;;
update-skip-range)
  update-skip-range $@
  ;;
release)
  release $@
  ;;
release-snapshot)
  release-snapshot $@
  ;;
*)
  echo "Invalid command: $COMMAND"
  help
  exit 1
esac
