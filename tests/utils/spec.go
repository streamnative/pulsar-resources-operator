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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	rsutils "github.com/streamnative/pulsar-resources-operator/pkg/utils"
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
	var du rsutils.Duration = "1d"
	limitTime := &du
	ttl := &du
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
			BacklogQuotaLimitTime: limitTime,
			BacklogQuotaLimitSize: &backlogSize,
			Bundles:               &bundle,
			MessageTTL:            ttl,
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

// MakePulsarPackage will generate a object of PulsarPackage
func MakePulsarPackage(namespace, name, packageURL, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarPackage {
	return &v1alpha1.PulsarPackage{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarPackageSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			PackageURL:      packageURL,
			FileURL:         "https://github.com/freeznet/pulsar-functions-api-examples/raw/main/api-examples-2.10.4.3.jar",
			Description:     "api-examples.jar",
			LifecyclePolicy: policy,
		},
	}
}

// MakePulsarFunction will generate a object of PulsarFunction
func MakePulsarFunction(namespace, name, functionPackageUrl, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarFunction {
	return &v1alpha1.PulsarFunction{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarFunctionSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			LifecyclePolicy:              policy,
			Jar:                          &v1alpha1.PackageContentRef{URL: functionPackageUrl},
			Tenant:                       "public",
			Namespace:                    "default",
			Name:                         name,
			Inputs:                       []string{"input"},
			Output:                       "output",
			Parallelism:                  1,
			ProcessingGuarantees:         "ATLEAST_ONCE",
			ClassName:                    "org.apache.pulsar.functions.api.examples.ExclamationFunction",
			SubName:                      "test-sub",
			SubscriptionPosition:         "Latest",
			CleanupSubscription:          true,
			SkipToLatest:                 true,
			ForwardSourceMessageProperty: true,
			RetainKeyOrdering:            true,
			AutoAck:                      true,
			MaxMessageRetries:            pointer.Int(101),
			DeadLetterTopic:              "dl-topic",
			LogTopic:                     "func-log",
			TimeoutMs:                    pointer.Int64(6666),
			Secrets: map[string]v1alpha1.SecretKeyRef{
				"SECRET1": {
					"sectest", "hello",
				},
			},
			CustomRuntimeOptions: getMapToJSON(map[string]interface{}{
				"env": map[string]string{
					"HELLO": "WORLD",
				},
			}),
		},
	}
}

// MakePulsarSink will generate a object of PulsarSink
func MakePulsarSink(namespace, name, sinkPackageUrl, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarSink {
	return &v1alpha1.PulsarSink{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarSinkSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			LifecyclePolicy:            policy,
			Archive:                    &v1alpha1.PackageContentRef{URL: sinkPackageUrl},
			Tenant:                     "public",
			Namespace:                  "default",
			Name:                       name,
			Inputs:                     []string{"sink-input"},
			Parallelism:                1,
			ProcessingGuarantees:       "EFFECTIVELY_ONCE",
			CleanupSubscription:        false,
			SourceSubscriptionPosition: "Latest",
			AutoAck:                    true,
			ClassName:                  "org.apache.pulsar.io.datagenerator.DataGeneratorPrintSink",
			Resources: &v1alpha1.Resources{
				CPU:  "1",
				RAM:  2048,
				Disk: 102400,
			},
			Secrets: map[string]v1alpha1.SecretKeyRef{
				"SECRET1": {
					"sectest", "hello",
				},
			},
			CustomRuntimeOptions: getMapToJSON(map[string]interface{}{
				"env": map[string]string{
					"HELLO": "WORLD",
				},
			}),
		},
	}
}

func getPulsarSourceConfig() *v1.JSON {
	c := map[string]interface{}{
		"sleepBetweenMessages": 1000,
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil
	}
	return &v1.JSON{Raw: bytes}
}

func getMapToJSON(c map[string]interface{}) *v1.JSON {
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil
	}
	return &v1.JSON{Raw: bytes}
}

// MakePulsarSource will generate a object of PulsarSource
func MakePulsarSource(namespace, name, sourcePackageUrl, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarSource {
	return &v1alpha1.PulsarSource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarSourceSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			LifecyclePolicy:      policy,
			Archive:              &v1alpha1.PackageContentRef{URL: sourcePackageUrl},
			Tenant:               "public",
			Namespace:            "default",
			Name:                 name,
			TopicName:            "sink-input",
			Parallelism:          1,
			ProcessingGuarantees: "EFFECTIVELY_ONCE",
			ClassName:            "org.apache.pulsar.io.datagenerator.DataGeneratorSource",
			Configs:              getPulsarSourceConfig(),
			Resources: &v1alpha1.Resources{
				CPU:  "1",
				RAM:  512,
				Disk: 102400,
			},
			Secrets: map[string]v1alpha1.SecretKeyRef{
				"SECRET1": {
					"sectest", "hello",
				},
			},
			CustomRuntimeOptions: getMapToJSON(map[string]interface{}{
				"env": map[string]string{
					"HELLO": "WORLD",
				},
			}),
		},
	}
}
