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

import (
	"github.com/streamnative/cloud-api-server/pkg/apis/cloud/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretSpec defines the desired state of StreamNative Cloud Secret
type SecretSpec struct {
	// APIServerRef is the reference to the StreamNativeCloudConnection
	// +required
	APIServerRef corev1.LocalObjectReference `json:"apiServerRef"`

	// InstanceName is the name of the instance this secret is for (e.g. pulsar-instance)
	// +optional
	InstanceName string `json:"instanceName"`

	// Location is the location of the secret.
	// +optional
	Location string `json:"location"`

	// the value should be base64 encoded
	// +optional
	Data map[string]string `json:"data,omitempty"`

	// SecretRef is the reference to the kubernetes secret
	// When SecretRef is set, it will be used to fetch the secret data.
	// Data will be ignored.
	// +optional
	SecretRef *KubernetesSecretReference `json:"secretRef,omitempty"`

	// PoolMemberName is the pool member to deploy the secret.
	// +optional
	PoolMemberName *string `json:"poolMemberName,omitempty"`

	// +optional
	// +listType=atomic
	Tolerations []Toleration `json:"tolerations,omitempty"`
}

// SecretStatus defines the observed state of StreamNative Cloud Secret
type SecretStatus struct {
	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,categories={streamnative,all}
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"

// Secret is the Schema for the StreamNative Cloud Secret API
type Secret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretSpec   `json:"spec,omitempty"`
	Status SecretStatus `json:"status,omitempty"`
}

// PoolMemberReference is a reference to a pool member with a given name.
type PoolMemberReference struct {
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	Name      string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

func (r PoolMemberReference) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}

// KubernetesSecretReference is a reference to a Kubernetes Secret with a given name.
type KubernetesSecretReference struct {
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	Name      string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

func (r KubernetesSecretReference) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}

// The workload this Toleration is attached to tolerates any taint that matches
// the triple <key,value,effect> using the matching operator <operator>.
type Toleration struct {
	// Key is the taint key that the toleration applies to. Empty means match all taint keys.
	// If the key is empty, operator must be Exists; this combination means to match all values and all keys.
	// +optional
	Key string `json:"key,omitempty" protobuf:"bytes,1,opt,name=key"`
	// Operator represents a key's relationship to the value.
	// Valid operators are Exists and Equal. Defaults to Equal.
	// Exists is equivalent to wildcard for value, so that a workload can
	// tolerate all taints of a particular category.
	// +optional
	Operator v1alpha1.TolerationOperator `json:"operator,omitempty" protobuf:"bytes,2,opt,name=operator,casttype=TolerationOperator"`
	// Value is the taint value the toleration matches to.
	// If the operator is Exists, the value should be empty, otherwise just a regular string.
	// +optional
	Value string `json:"value,omitempty" protobuf:"bytes,3,opt,name=value"`
	// Effect indicates the taint effect to match. Empty means match all taint effects.
	// When specified, allowed values are NoSchedule and PreferNoSchedule.
	// +optional
	Effect v1alpha1.TaintEffect `json:"effect,omitempty" protobuf:"bytes,4,opt,name=effect,casttype=TaintEffect"`
}

//+kubebuilder:object:root=true

// SecretList contains a list of Secret
type SecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Secret{}, &SecretList{})
}
