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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterRole
// +k8s:openapi-gen=true
// +resource:path=clusterroles,strategy=ClusterRoleStrategy
// +kubebuilder:categories=all
type ClusterRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ClusterRoleSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ClusterRoleStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ClusterRoleSpec defines the desired state of ClusterRole
type ClusterRoleSpec struct {
	// +listType=atomic
	// Permissions Designed for general permission format
	//     SERVICE.RESOURCE.VERB
	Permissions []string `json:"permissions,omitempty" protobuf:"bytes,1,opt,name=permissions"`
}

type FailedCluster struct {
	// Name is the Cluster's name
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// Namespace is the Cluster's namespace
	Namespace string `json:"namespace" protobuf:"bytes,2,opt,name=namespace"`

	// Reason is the failed reason
	Reason string `json:"reason" protobuf:"bytes,3,opt,name=reason"`
}

// ClusterRoleStatus defines the observed state of ClusterRole
type ClusterRoleStatus struct {
	// Conditions is an array of current observed conditions.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// FailedClusters is an array of clusters which failed to apply the ClusterRole resources.
	// +deprecated, shouldn't expose the failed clusters to the user
	// +optional
	// +listType=atomic
	FailedClusters []FailedCluster `json:"failedClusters,omitempty" protobuf:"bytes,2,rep,name=failedClusters"`
}
