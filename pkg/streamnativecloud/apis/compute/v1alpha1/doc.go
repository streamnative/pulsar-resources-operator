// Copyright 2025 StreamNative
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

//go:generate go run k8s.io/code-generator/cmd/deepcopy-gen -O zz_generated.deepcopy -i . -h ../../../../../hack/boilerplate.go.txt
//go:generate go run k8s.io/code-generator/cmd/defaulter-gen -O zz_generated.defaults -i . -h ../../../../../hack/boilerplate.go.txt
//go:generate go run k8s.io/code-generator/cmd/conversion-gen -O zz_generated.conversion -i . -h ../../../../../hack/boilerplate.go.txt

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/compute
// +k8s:defaulter-gen=TypeMeta
// +k8s:protobuf-gen=package
// +groupName=compute.streamnative.io
//
//nolint:stylecheck
package v1alpha1 // import "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/compute/v1alpha1"
