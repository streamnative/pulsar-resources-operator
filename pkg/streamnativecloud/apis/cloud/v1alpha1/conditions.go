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

type ConditionType string

type ConditionReason string

const (
	ConditionTypeReady ConditionType = "Ready"

	ConditionReasonReady    ConditionReason = "Ready"
	ConditionReasonNotReady ConditionReason = "NotReady"
)

type Condition struct {
	Type               ConditionType          `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ConditionType"`
	Status             metav1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/apimachinery/pkg/apis/meta/v1.ConditionStatus"`
	Reason             ConditionReason        `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason,casttype=ConditionReason"`
	Message            string                 `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
	// +optional
	// +kubebuilder:validation:Minimum=0
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,6,opt,name=observedGeneration"`
}
