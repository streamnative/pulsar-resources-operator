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

// Taint - The workload cluster this Taint is attached to has the "effect" on
// any workload that does not tolerate the Taint.
type Taint struct {
	// Required. The taint key to be applied to a workload cluster.
	Key string `json:"key" protobuf:"bytes,1,opt,name=key"`
	// Optional. The taint value corresponding to the taint key.
	// +optional
	Value string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	// Required. The effect of the taint on workloads
	// that do not tolerate the taint.
	// Valid effects are NoSchedule, PreferNoSchedule and NoExecute.
	Effect TaintEffect `json:"effect" protobuf:"bytes,3,opt,name=effect,casttype=TaintEffect"`
	// TimeAdded represents the time at which the taint was added.
	// +optional
	TimeAdded *metav1.Time `json:"timeAdded,omitempty" protobuf:"bytes,4,opt,name=timeAdded"`
}

type TaintEffect string

const (
	// TaintEffectNoSchedule has the effect of not scheduling new workloads onto the workload cluster
	// unless they tolerate the taint.
	// Enforced by the scheduler.
	TaintEffectNoSchedule TaintEffect = "NoSchedule"

	// TaintEffectPreferNoSchedule is Like TaintEffectNoSchedule, but the scheduler tries not to schedule
	// new workloads onto the workload cluster, rather than prohibiting it entirely.
	// Enforced by the scheduler.
	TaintEffectPreferNoSchedule TaintEffect = "PreferNoSchedule"

	// TaintEffectNoCleanup has the effect of skipping any cleanup of workload objects.
	// Set this effect to allow for API object finalization, which normally requires
	// the workload cluster to be ready, to proceed without normal cleanups.
	TaintEffectNoCleanup TaintEffect = "NoCleanup"

	// TaintEffectNoConnect has the effect of skipping any attempts to connect to the workload cluster.
	TaintEffectNoConnect TaintEffect = "NoConnect"
)

// Toleration The workload this Toleration is attached to tolerates any taint that matches
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
	Operator TolerationOperator `json:"operator,omitempty" protobuf:"bytes,2,opt,name=operator,casttype=TolerationOperator"`
	// Value is the taint value the toleration matches to.
	// If the operator is Exists, the value should be empty, otherwise just a regular string.
	// +optional
	Value string `json:"value,omitempty" protobuf:"bytes,3,opt,name=value"`
	// Effect indicates the taint effect to match. Empty means match all taint effects.
	// When specified, allowed values are NoSchedule and PreferNoSchedule.
	// +optional
	Effect TaintEffect `json:"effect,omitempty" protobuf:"bytes,4,opt,name=effect,casttype=TaintEffect"`
}

// TolerationOperator A toleration operator is the set of operators that can be used in a toleration.
type TolerationOperator string

const (
	TolerationOpExists TolerationOperator = "Exists"
	TolerationOpEqual  TolerationOperator = "Equal"
)

func (t Toleration) MatchAllKeys() bool {
	return t.Key == "" && t.Operator == TolerationOpExists
}

func (t Toleration) MatchAllEffects() bool {
	return t.Effect == ""
}

func (t Toleration) MatchAllValues() bool {
	return t.Operator == TolerationOpExists
}

func (t Toleration) Tolerates(taint Taint) bool {
	if !t.MatchAllKeys() && t.Key != taint.Key {
		return false
	}
	if !t.MatchAllEffects() && t.Effect != taint.Effect {
		return false
	}
	if !t.MatchAllValues() {
		if (t.Operator == "" || t.Operator == TolerationOpEqual) && t.Value != taint.Value {
			return false
		}
	}
	return true
}
