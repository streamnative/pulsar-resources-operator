#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

echo $SCRIPT_ROOT

source "${SCRIPT_ROOT}/hack/kube_codegen.sh"

THIS_PKG="github.com/streamnative/pulsar-resources-operator"

#kube::codegen::gen_helpers \
#    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#    --extra-peer-dir k8s.io/apimachinery/pkg/apis/meta/v1 \
#    --extra-peer-dir k8s.io/apimachinery/pkg/conversion \
#    --extra-peer-dir k8s.io/apimachinery/pkg/runtime \
#    "${SCRIPT_ROOT}/pkg/streamnativecloud/apis"

# Generate register code
#
# USAGE: kube::codegen::gen_register [FLAGS] <input-dir>
#
# <input-dir>
#   The root directory under which to search for Go files which request code to
#   be generated.  This must be a local path, not a Go package.
#
#   See note at the top about package structure below that.
#
# FLAGS:
#
#   --boilerplate <string = path_to_kube_codegen_boilerplate>
#     An optional override for the header file to insert into generated files.
#
function kube::codegen::gen_register() {
    local in_dir=""
    local boilerplate="${KUBE_CODEGEN_ROOT}/hack/boilerplate.go.txt"
    local v="${KUBE_VERBOSE:-0}"

    while [ "$#" -gt 0 ]; do
        case "$1" in
            "--boilerplate")
                boilerplate="$2"
                shift 2
                ;;
            *)
                if [[ "$1" =~ ^-- ]]; then
                    echo "unknown argument: $1" >&2
                    return 1
                fi
                if [ -n "$in_dir" ]; then
                    echo "too many arguments: $1 (already have $in_dir)" >&2
                    return 1
                fi
                in_dir="$1"
                shift
                ;;
        esac
    done

    if [ -z "${in_dir}" ]; then
        echo "input-dir argument is required" >&2
        return 1
    fi

    (
        # To support running this from anywhere, first cd into this directory,
        # and then install with forced module mode on and fully qualified name.
        cd "${KUBE_CODEGEN_ROOT}"
        BINS=(
            register-gen
        )
        # shellcheck disable=2046 # printf word-splitting is intentional
        GO111MODULE=on go install $(printf "k8s.io/code-generator/cmd/%s " "${BINS[@]}")
    )
    # Go installs in $GOBIN if defined, and $GOPATH/bin otherwise
    gobin="${GOBIN:-$(go env GOPATH)/bin}"

    # Register
    #
    local input_pkgs=()
    while read -r dir; do
        pkg="$(cd "${dir}" && GO111MODULE=on go list -find .)"
        input_pkgs+=("${pkg}")
    done < <(
        ( kube::codegen::internal::grep -l --null \
            -e '^\s*//\s*+groupName' \
            -r "${in_dir}" \
            --include '*.go' \
            || true \
        ) | while read -r -d $'\0' F; do dirname "${F}"; done \
          | LC_ALL=C sort -u
    )

    if [ "${#input_pkgs[@]}" != 0 ]; then
        echo "Generating register code for ${#input_pkgs[@]} targets"

        kube::codegen::internal::findz \
            "${in_dir}" \
            -type f \
            -name zz_generated.register.go \
            | xargs -0 rm -f

        "${gobin}/register-gen" \
            -v "${v}" \
            --output-file zz_generated.api.register.go \
            --go-header-file "${boilerplate}" \
            "${input_pkgs[@]}"
    fi
}

kube::codegen::gen_register \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/streamnativecloud/apis"

kube::codegen::gen_client \
    --with-watch \
    --output-dir "${SCRIPT_ROOT}/pkg/streamnativecloud/client" \
    --output-pkg "${THIS_PKG}/pkg/streamnativecloud/client" \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    --listers-name listers_generated \
    --informers-name informers_generated \
    --clientset-name clientset_generated \
    --versioned-name clientset \
    "${SCRIPT_ROOT}/pkg/streamnativecloud/apis"

# Run goimports to fix import statements in generated files
echo "Running goimports to fix imports in generated files..."

# Ensure goimports is installed
if ! command -v goimports &> /dev/null; then
    echo "goimports not found, installing..."
    GO111MODULE=on go install golang.org/x/tools/cmd/goimports@latest
fi

# Get goimports path
GOIMPORTS="${GOBIN:-$(go env GOPATH)/bin}/goimports"

# Ensure goimports is executable
if [ ! -x "$GOIMPORTS" ]; then
    echo "Error: Cannot find executable goimports, please ensure it's properly installed" >&2
    exit 1
fi

# Define directories with generated files to process
GENERATED_DIRS=(
    "${SCRIPT_ROOT}/pkg/streamnativecloud/apis"
    "${SCRIPT_ROOT}/pkg/streamnativecloud/client"
)

# Use find to locate all .go files and process them with goimports
for dir in "${GENERATED_DIRS[@]}"; do
    echo "Processing directory: $dir"
    find "$dir" -type f -name "*.go" -print0 | xargs -0 "$GOIMPORTS" -w
done

echo "Import statements fixing completed!"