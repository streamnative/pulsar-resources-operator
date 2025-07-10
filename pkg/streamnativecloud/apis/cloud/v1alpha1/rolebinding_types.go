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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoleBinding
// +k8s:openapi-gen=true
// +resource:path=rolebindings,strategy=RoleBindingStrategy
// +kubebuilder:categories=all
type RoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   RoleBindingSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status RoleBindingStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// RoleBindingSpec defines the desired state of RoleBinding
type RoleBindingSpec struct {
	// +listType=atomic
	Subjects []Subject `json:"subjects" protobuf:"bytes,1,rep,name=subjects"`
	RoleRef  RoleRef   `json:"roleRef" protobuf:"bytes,2,opt,name=roleRef"`

	// +optional
	ConditionGroup *ConditionGroup `json:"conditionGroup,omitempty" protobuf:"bytes,3,opt,name=conditionGroup"`
	// +optional
	CEL *string `json:"cel,omitempty" protobuf:"bytes,4,opt,name=cel"`

	// +optional
	// +listType=atomic
	// ResourceNames indicate the StreamNative Resource Name
	ResourceNames []ResourceName `json:"resourceNames,omitempty" protobuf:"bytes,5,opt,name=resourceNames"`
}

type ResourceName struct {
	Organization   string `json:"organization,omitempty" protobuf:"bytes,1,opt,name=organization"`
	Instance       string `json:"instance,omitempty" protobuf:"bytes,2,opt,name=instance"`
	Cluster        string `json:"cluster,omitempty" protobuf:"bytes,3,opt,name=cluster"`
	Tenant         string `json:"tenant,omitempty" protobuf:"bytes,4,opt,name=tenant"`
	Namespace      string `json:"namespace,omitempty" protobuf:"bytes,5,opt,name=namespace"`
	TopicDomain    string `json:"topicDomain,omitempty" protobuf:"bytes,6,opt,name=topicDomain"`
	TopicName      string `json:"topicName,omitempty" protobuf:"bytes,7,opt,name=topicName"`
	Subscription   string `json:"subscription,omitempty" protobuf:"bytes,8,opt,name=subscription"`
	ServiceAccount string `json:"serviceAccount,omitempty" protobuf:"bytes,9,opt,name=serviceAccount"`
	ApiKey         string `json:"apiKey,omitempty" protobuf:"bytes,10,opt,name=apiKey"`
	Secret         string `json:"secret,omitempty" protobuf:"bytes,11,opt,name=secret"`
} // RoleBindingStatus defines the observed state of RoleBinding
type RoleBindingStatus struct {
	// Conditions is an array of current observed conditions.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// FailedClusters is an array of clusters which failed to apply the RoleBinding resources.
	// +optional
	// +listType=atomic
	FailedClusters []FailedCluster `json:"failedClusters,omitempty" protobuf:"bytes,2,rep,name=failedClusters"`

	// SyncedClusters is a map of clusters which applied the RoleBinding resources and the RoleBinding's generation.
	// If the generation in this map is same with the current RoleBinding's generation, it will skip apply the
	// RoleBinding to the cluster
	// +optional
	// +listType=atomic
	SyncedClusters map[string]int64 `json:"syncedClusters,omitempty" protobuf:"bytes,3,rep,name=syncedClusters"`
}

// Deprecated sections

type ConditionGroupRelation int
type RoleBindingConditionType int
type RoleBindingConditionOperator int

const (
	RelationOr         ConditionGroupRelation       = 0
	RelationAnd        ConditionGroupRelation       = 1
	TypeSrn            RoleBindingConditionType     = 0
	OperatorKeyMatch   RoleBindingConditionOperator = 0
	OperatorRegexMatch RoleBindingConditionOperator = 1
)

// ConditionGroup
// Deprecated
type ConditionGroup struct {
	Relation        ConditionGroupRelation `json:"relation,omitempty" protobuf:"bytes,1,opt,name=relation"`
	Conditions      []RoleBindingCondition `json:"conditions,required" protobuf:"bytes,2,req,name=conditions"`
	ConditionGroups []ConditionGroup       `json:"conditionGroups,omitempty" protobuf:"bytes,3,opt,name=conditionGroups"`
}

// Srn
// Deprecated
type Srn struct {
	Schema       string `json:"schema,omitempty" protobuf:"bytes,1,opt,name=schema"`
	Version      string `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`
	Organization string `json:"organization,omitempty" protobuf:"bytes,3,opt,name=organization"`
	Instance     string `json:"instance,omitempty" protobuf:"bytes,4,opt,name=instance"`
	Cluster      string `json:"cluster,omitempty" protobuf:"bytes,5,opt,name=cluster"`
	Tenant       string `json:"tenant,omitempty" protobuf:"bytes,6,opt,name=tenant"`
	Namespace    string `json:"namespace,omitempty" protobuf:"bytes,7,opt,name=namespace"`
	TopicDomain  string `json:"topicDomain,omitempty" protobuf:"bytes,8,opt,name=topicDomain"`
	TopicName    string `json:"topicName,omitempty" protobuf:"bytes,9,opt,name=topicName"`
	Subscription string `json:"subscription,omitempty" protobuf:"bytes,10,opt,name=subscription"`
}

// RoleBindingCondition
// Deprecated
type RoleBindingCondition struct {
	Type     RoleBindingConditionType     `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`
	Operator RoleBindingConditionOperator `json:"operator,omitempty" protobuf:"bytes,2,opt,name=operator"`
	Srn      Srn                          `json:"srn,omitempty" protobuf:"bytes,3,opt,name=srn"`
}
