// Copyright 2022 StreamNative
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

package utils

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// MakePulsarConnection will generate a object of PulsarConnection, without authentication
func MakePulsarConnection(namespace, name, adminServiceURL string) *v1alpha1.PulsarConnection {
	return &v1alpha1.PulsarConnection{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarConnectionSpec{
			AdminServiceURL: adminServiceURL,
		},
	}
}

// MakePulsarTenant will generate a object of PulsarTenant
func MakePulsarTenant(namespace, name, tenantName, connectionName string, adminRoles, allowedClusters []string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTenant {
	return &v1alpha1.PulsarTenant{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTenantSpec{
			Name: tenantName,
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			AdminRoles:      adminRoles,
			AllowedClusters: allowedClusters,
			LifecyclePolicy: policy,
		},
	}
}

// MakePulsarNamespace will generate a object of PulsarNamespace
func MakePulsarNamespace(namespace, name, namespaceName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarNamespace {
	backlogSize := resource.MustParse("1Gi")
	bundle := int32(16)
	return &v1alpha1.PulsarNamespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarNamespaceSpec{
			Name: namespaceName,
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			BacklogQuotaLimitTime: &metav1.Duration{
				Duration: time.Hour * 24,
			},
			BacklogQuotaLimitSize: &backlogSize,
			Bundles:               &bundle,
			MessageTTL: &metav1.Duration{
				Duration: time.Hour * 1,
			},
		},
	}
}

// MakePulsarTopic will generate a object of PulsarTopic
func MakePulsarTopic(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name: topicName,
		},
	}
}

// MakePulsarPermission will generate a object of PulsarPermission
func MakePulsarPermission(namespace, name, resourceName, connectionName string, resourceType v1alpha1.PulsarResourceType,
	roles, actions []string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarPermission {
	return &v1alpha1.PulsarPermission{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarPermissionSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			ResourceName: resourceName,
			ResoureType:  resourceType,
			Roles:        roles,
			Actions:      actions,
		},
	}
}
