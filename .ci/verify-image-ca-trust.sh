#!/usr/bin/env bash
# Copyright 2024 StreamNative
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

repo_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
platform=${PLATFORM:-linux/amd64}
case "$platform" in
  linux/amd64|linux/arm64) ;;
  *) echo "Unsupported platform: $platform" >&2; exit 1 ;;
esac
arch=${platform#linux/}

image=${1:-pulsar-resources-operator-ca-test:$arch}
if [ "$#" -eq 0 ]; then
  # Exercise the production runtime stage without downloading private Go modules.
  docker build --platform "$platform" --target runtime \
    -f "$repo_dir/Dockerfile" -t "$image" "$repo_dir"
fi

work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT
(
  cd "$repo_dir"
  GOWORK=off CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -mod=readonly -trimpath \
    -o "$work_dir/verify-ca-trust" ./hack/verify-ca-trust
)
chmod 755 "$work_dir/verify-ca-trust"

docker run --rm --platform "$platform" --read-only --network none \
  --user 65532:65532 --entrypoint /verify-ca-trust \
  -v "$work_dir/verify-ca-trust:/verify-ca-trust:ro" \
  -v "$repo_dir/hack/verify-ca-trust/testdata:/chains:ro" \
  "$image" --chains-dir /chains
