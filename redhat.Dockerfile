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

# Build the manager binary
FROM golang:1.21.10-alpine3.20 as builder

ARG ACCESS_TOKEN="none"

RUN go env -w GOPRIVATE=github.com/streamnative \
    && apk add --no-cache ca-certificates git \
    && git config --global url."https://${ACCESS_TOKEN}:@github.com/".insteadOf "https://github.com/"

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go


# Use ubi image as the base image which is required by the red hat certification. 
# Base on the image size, the order is ubi > ubi-minimal > ubi-micro.
# https://access.redhat.com/documentation/en-us/red_hat_software_certification/8.45/html/red_hat_openshift_software_certification_policy_guide/assembly-requirements-for-container-images_openshift-sw-cert-policy-introduction#con-image-metadata-requirements_openshift-sw-cert-policy-container-images
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ARG VERSION

LABEL name="pulsar-resources-operator" \
      vendor="StreamNative, Inc." \
      maintainer="StreamNative, Inc." \
      version="${VERSION}" \
      release="${VERSION}" \
      summary="The Pulsar Resources Operator provides a declarative way to manage the pulsar resources on a Kubernetes cluster." \
	  description="The Pulsar Resources Operator provides full lifecycle management for the following Pulsar resources, including creation, update, and deletion."
WORKDIR /
COPY --from=builder /workspace/manager .
COPY LICENSE /licenses/LICENSE
USER 1001

ENTRYPOINT ["/manager"]
