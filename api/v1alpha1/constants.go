// Copyright 2024 StreamNative
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

package v1alpha1

const (
	// ConditionReady indicates status condition ready
	ConditionReady string = "Ready"
	// ConditionTopicPolicyReady indicates the topic policy ready
	ConditionTopicPolicyReady string = "PolicyReady"
	// FinalizerName is the finalizer string that add to object
	FinalizerName string = "cloud.streamnative.io/finalizer"

	// AuthPluginToken indicates the authentication pulgin type token
	AuthPluginToken string = "org.apache.pulsar.client.impl.auth.AuthenticationToken" // #nosec G101
	// AuthPluginOAuth2 indicates the authentication pulgin type oauth2
	AuthPluginOAuth2 string = "org.apache.pulsar.client.impl.auth.oauth2.AuthenticationOAuth2"
)
