// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.
package utils

import (
	"time"

	pulsarv1alpha1 "github.com/streamnative/resources-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MakePulsarConnection will generate a object of PulsarConnection, without authentication
func MakePulsarConnection(namespace, name, adminServiceURL, tokenSecretName string) *pulsarv1alpha1.PulsarConnection {
	return &pulsarv1alpha1.PulsarConnection{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: pulsarv1alpha1.PulsarConnectionSpec{
			AdminServiceURL: adminServiceURL,
		},
	}
}

// MakePulsarTenant will generate a object of PulsarTenant
func MakePulsarTenant(namespace, name, tenantName, connectionName string, adminRoles, allowedClusters []string, policy pulsarv1alpha1.PulsarResourceLifeCyclePolicy) *pulsarv1alpha1.PulsarTenant {
	return &pulsarv1alpha1.PulsarTenant{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: pulsarv1alpha1.PulsarTenantSpec{
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
func MakePulsarNamespace(namespace, name, namespaceName, connectionName string, policy pulsarv1alpha1.PulsarResourceLifeCyclePolicy) *pulsarv1alpha1.PulsarNamespace {
	backlogSize := resource.MustParse("1Gi")
	bundle := int32(16)
	return &pulsarv1alpha1.PulsarNamespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: pulsarv1alpha1.PulsarNamespaceSpec{
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
func MakePulsarTopic(namespace, name, topicName, connectionName string, policy pulsarv1alpha1.PulsarResourceLifeCyclePolicy) *pulsarv1alpha1.PulsarTopic {
	topic := &pulsarv1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: pulsarv1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name: topicName,
		},
	}
	return topic
}

// MakePulsarPermission will generate a object of PulsarPermission
func MakePulsarPermission(namespace, name, resourceName, connectionName string, resourceType pulsarv1alpha1.PulsarResoureType,
	roles, actions []string, policy pulsarv1alpha1.PulsarResourceLifeCyclePolicy) *pulsarv1alpha1.PulsarPermission {
	return &pulsarv1alpha1.PulsarPermission{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: pulsarv1alpha1.PulsarPermissionSpec{
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
