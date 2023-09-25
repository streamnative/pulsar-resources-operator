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
