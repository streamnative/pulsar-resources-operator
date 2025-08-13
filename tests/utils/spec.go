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

package utils

import (
	"encoding/json"
	"os"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"

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
			TopicAutoCreationConfig: &v1alpha1.TopicAutoCreationConfig{
				Allow:      true,
				Type:       "partitioned",
				Partitions: ptr.To[int32](10),
			},
		},
	}
}

// MakePulsarNamespaceWithRateLimiting will generate a PulsarNamespace with rate limiting configurations
func MakePulsarNamespaceWithRateLimiting(namespace, name, namespaceName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarNamespace {
	backlogSize := resource.MustParse("5Gi")
	bundle := int32(16)
	var du rsutils.Duration = "12h"
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
			LifecyclePolicy:       policy,

			// Rate limiting configurations
			DispatchRate: &v1alpha1.DispatchRate{
				DispatchThrottlingRateInMsg:  ptr.To[int32](1000),
				DispatchThrottlingRateInByte: ptr.To[int64](1048576), // 1MB
				RatePeriodInSecond:           ptr.To[int32](1),
			},
			SubscriptionDispatchRate: &v1alpha1.DispatchRate{
				DispatchThrottlingRateInMsg:  ptr.To[int32](500),
				DispatchThrottlingRateInByte: ptr.To[int64](524288), // 512KB
				RatePeriodInSecond:           ptr.To[int32](1),
			},
			ReplicatorDispatchRate: &v1alpha1.DispatchRate{
				DispatchThrottlingRateInMsg:  ptr.To[int32](800),
				DispatchThrottlingRateInByte: ptr.To[int64](838860), // ~800KB
				RatePeriodInSecond:           ptr.To[int32](1),
			},
			PublishRate: &v1alpha1.PublishRate{
				PublishThrottlingRateInMsg:  ptr.To[int32](2000),
				PublishThrottlingRateInByte: ptr.To[int64](2097152), // 2MB
			},
			SubscribeRate: &v1alpha1.SubscribeRate{
				SubscribeThrottlingRatePerConsumer: ptr.To[int32](10),
				RatePeriodInSecond:                 ptr.To[int32](30),
			},
		},
	}
}

// MakePulsarNamespaceWithStoragePolicies will generate a PulsarNamespace with storage and persistence configurations
func MakePulsarNamespaceWithStoragePolicies(namespace, name, namespaceName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarNamespace {
	backlogSize := resource.MustParse("20Gi")
	retentionSize := resource.MustParse("50Gi")
	bundle := int32(32)
	var retentionTime rsutils.Duration = "48h"
	var backlogQuotaTime rsutils.Duration = "24h" // Must be less than retentionTime to avoid conflicts
	var subscriptionExpTime rsutils.Duration = "7d"
	limitTime := &retentionTime
	backlogLimitTime := &backlogQuotaTime

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
			Bundles:         &bundle,
			LifecyclePolicy: policy,

			// Storage and persistence configurations
			RetentionTime:               limitTime,
			RetentionSize:               &retentionSize,
			BacklogQuotaLimitSize:       &backlogSize,
			BacklogQuotaLimitTime:       backlogLimitTime,
			BacklogQuotaRetentionPolicy: ptr.To("producer_request_hold"),
			BacklogQuotaType:            ptr.To("destination_storage"),
			SubscriptionExpirationTime:  &subscriptionExpTime,

			PersistencePolicies: &v1alpha1.PersistencePolicies{
				BookkeeperEnsemble:             ptr.To[int32](5),
				BookkeeperWriteQuorum:          ptr.To[int32](3),
				BookkeeperAckQuorum:            ptr.To[int32](2),
				ManagedLedgerMaxMarkDeleteRate: ptr.To("1.5"),
			},

			CompactionThreshold: ptr.To[int64](104857600), // 100MB

			InactiveTopicPolicies: &v1alpha1.InactiveTopicPolicies{
				InactiveTopicDeleteMode:      ptr.To("delete_when_no_subscriptions"),
				MaxInactiveDurationInSeconds: ptr.To[int32](3600), // 1 hour
				DeleteWhileInactive:          ptr.To(true),
			},

			Properties: map[string]string{
				"environment": "test",
				"team":        "qa",
			},
		},
	}
}

// MakePulsarNamespaceWithSecurityConfig will generate a PulsarNamespace with security configurations
func MakePulsarNamespaceWithSecurityConfig(namespace, name, namespaceName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarNamespace {
	backlogSize := resource.MustParse("10Gi")
	retentionSize := resource.MustParse("100Gi")
	bundle := int32(16)
	var retentionTime rsutils.Duration = "720h"       // 30 days
	var subscriptionExpTime rsutils.Duration = "168h" // 7 days
	limitTime := &retentionTime

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
			Bundles:         &bundle,
			LifecyclePolicy: policy,

			// Security configurations
			EncryptionRequired:      ptr.To(true),
			ValidateProducerName:    ptr.To(true),
			IsAllowAutoUpdateSchema: ptr.To(false),
			AntiAffinityGroup:       ptr.To("high-availability"),

			// Schema policies for security
			SchemaValidationEnforced: ptr.To(true),

			// Rate limiting for security (prevent spam)
			PublishRate: &v1alpha1.PublishRate{
				PublishThrottlingRateInMsg:  ptr.To[int32](1000),
				PublishThrottlingRateInByte: ptr.To[int64](1048576), // 1MB
			},
			SubscribeRate: &v1alpha1.SubscribeRate{
				SubscribeThrottlingRatePerConsumer: ptr.To[int32](5),
				RatePeriodInSecond:                 ptr.To[int32](60),
			},

			// Storage policies for audit and compliance
			RetentionTime:               limitTime,
			RetentionSize:               &retentionSize,
			BacklogQuotaLimitSize:       &backlogSize,
			BacklogQuotaLimitTime:       &subscriptionExpTime,
			BacklogQuotaRetentionPolicy: ptr.To("producer_request_hold"),
			SubscriptionExpirationTime:  &subscriptionExpTime,

			PersistencePolicies: &v1alpha1.PersistencePolicies{
				BookkeeperEnsemble:    ptr.To[int32](5),
				BookkeeperWriteQuorum: ptr.To[int32](3),
				BookkeeperAckQuorum:   ptr.To[int32](3), // Higher ack quorum for security
			},

			CompactionThreshold: ptr.To[int64](52428800), // 50MB - more frequent compaction

			Properties: map[string]string{
				"security-level":      "high",
				"compliance":          "required",
				"data-classification": "confidential",
			},

			Deduplication: ptr.To(true),
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

// MakePulsarTopicWithCompactionThreshold will generate a object of PulsarTopic with compaction threshold
func MakePulsarTopicWithCompactionThreshold(namespace, name, topicName, connectionName string, compactionThreshold int64, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:                topicName,
			CompactionThreshold: &compactionThreshold,
			LifecyclePolicy:     policy,
		},
	}
}

// MakePulsarTopicWithPersistencePolicies will generate a PulsarTopic with persistence configurations
func MakePulsarTopicWithPersistencePolicies(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			PersistencePolicies: &v1alpha1.PersistencePolicies{
				BookkeeperEnsemble:             ptr.To[int32](3),
				BookkeeperWriteQuorum:          ptr.To[int32](2),
				BookkeeperAckQuorum:            ptr.To[int32](2),
				ManagedLedgerMaxMarkDeleteRate: ptr.To("2.0"),
			},
		},
	}
}

// MakePulsarTopicWithDelayedDelivery will generate a PulsarTopic with delayed delivery configuration
func MakePulsarTopicWithDelayedDelivery(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			DelayedDelivery: &v1alpha1.DelayedDeliveryData{
				Active:         ptr.To(true),
				TickTimeMillis: ptr.To[int64](1000), // 1 second
			},
		},
	}
}

// MakePulsarTopicWithDispatchRate will generate a PulsarTopic with dispatch rate configuration
func MakePulsarTopicWithDispatchRate(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			DispatchRate: &v1alpha1.DispatchRate{
				DispatchThrottlingRateInMsg:  ptr.To[int32](500),
				DispatchThrottlingRateInByte: ptr.To[int64](524288), // 512KB
				RatePeriodInSecond:           ptr.To[int32](1),
			},
		},
	}
}

// MakePulsarTopicWithPublishRate will generate a PulsarTopic with publish rate configuration
func MakePulsarTopicWithPublishRate(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			PublishRate: &v1alpha1.PublishRate{
				PublishThrottlingRateInMsg:  ptr.To[int32](1000),
				PublishThrottlingRateInByte: ptr.To[int64](1048576), // 1MB
			},
		},
	}
}

// MakePulsarTopicWithInactiveTopicPolicies will generate a PulsarTopic with inactive topic policies
func MakePulsarTopicWithInactiveTopicPolicies(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			InactiveTopicPolicies: &v1alpha1.InactiveTopicPolicies{
				InactiveTopicDeleteMode:      ptr.To("delete_when_no_subscriptions"),
				MaxInactiveDurationInSeconds: ptr.To[int32](1800), // 30 minutes
				DeleteWhileInactive:          ptr.To(true),
			},
		},
	}
}

// MakePulsarTopicWithAllNewPolicies will generate a PulsarTopic with all new policy configurations
func MakePulsarTopicWithAllNewPolicies(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			PersistencePolicies: &v1alpha1.PersistencePolicies{
				BookkeeperEnsemble:             ptr.To[int32](5),
				BookkeeperWriteQuorum:          ptr.To[int32](3),
				BookkeeperAckQuorum:            ptr.To[int32](2),
				ManagedLedgerMaxMarkDeleteRate: ptr.To("1.5"),
			},
			DelayedDelivery: &v1alpha1.DelayedDeliveryData{
				Active:         ptr.To(true),
				TickTimeMillis: ptr.To[int64](2000), // 2 seconds
			},
			DispatchRate: &v1alpha1.DispatchRate{
				DispatchThrottlingRateInMsg:  ptr.To[int32](750),
				DispatchThrottlingRateInByte: ptr.To[int64](786432), // 768KB
				RatePeriodInSecond:           ptr.To[int32](1),
			},
			PublishRate: &v1alpha1.PublishRate{
				PublishThrottlingRateInMsg:  ptr.To[int32](1500),
				PublishThrottlingRateInByte: ptr.To[int64](1572864), // 1.5MB
			},
			InactiveTopicPolicies: &v1alpha1.InactiveTopicPolicies{
				InactiveTopicDeleteMode:      ptr.To("delete_when_subscriptions_caught_up"),
				MaxInactiveDurationInSeconds: ptr.To[int32](3600), // 1 hour
				DeleteWhileInactive:          ptr.To(false),
			},
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
			Secrets: map[string]v1alpha1.FunctionSecretKeyRef{
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
			Secrets: map[string]v1alpha1.FunctionSecretKeyRef{
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
			Secrets: map[string]v1alpha1.FunctionSecretKeyRef{
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

// MakeNSIsolationPolicy will generate a object of PulsarNSIsolationPolicy
func MakeNSIsolationPolicy(namespace, name, clusterName, connectionName string, namespaces, primary, secondary []string, params map[string]string) *v1alpha1.PulsarNSIsolationPolicy {
	return &v1alpha1.PulsarNSIsolationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarNSIsolationPolicySpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:                     name,
			Cluster:                  clusterName,
			Namespaces:               namespaces,
			Primary:                  primary,
			Secondary:                secondary,
			AutoFailoverPolicyType:   v1alpha1.MinAvailable,
			AutoFailoverPolicyParams: params,
		},
	}
}

func GetEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

// MakePulsarTopicWithSubscribeRate will generate a PulsarTopic with subscribe rate configuration
func MakePulsarTopicWithSubscribeRate(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			SubscribeRate: &v1alpha1.SubscribeRate{
				SubscribeThrottlingRatePerConsumer: ptr.To[int32](10),
				RatePeriodInSecond:                 ptr.To[int32](30),
			},
		},
	}
}

// MakePulsarTopicWithOffloadPolicies will generate a PulsarTopic with offload policies configuration
func MakePulsarTopicWithOffloadPolicies(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {

	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			OffloadPolicies: &v1alpha1.OffloadPolicies{
				ManagedLedgerOffloadDriver:           "aws-s3",
				ManagedLedgerOffloadMaxThreads:       5,
				ManagedLedgerOffloadThresholdInBytes: 1073741824, // 1GB
			},
		},
	}
}

// MakePulsarTopicWithAutoSubscriptionCreation will generate a PulsarTopic with auto subscription creation configuration
func MakePulsarTopicWithAutoSubscriptionCreation(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			AutoSubscriptionCreation: &v1alpha1.AutoSubscriptionCreationOverride{
				AllowAutoSubscriptionCreation: true,
			},
		},
	}
}

// MakePulsarTopicWithSchemaCompatibilityStrategy will generate a PulsarTopic with schema compatibility strategy configuration
func MakePulsarTopicWithSchemaCompatibilityStrategy(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	strategy := v1alpha1.SchemaCompatibilityStrategy("BACKWARD")
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:                        topicName,
			LifecyclePolicy:             policy,
			SchemaCompatibilityStrategy: &strategy,
		},
	}
}

// MakePulsarTopicWithInfiniteRetention creates a PulsarTopic with infinite retention policies (both time and size)
func MakePulsarTopicWithInfiniteRetention(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	infiniteTime := rsutils.Duration("-1")
	infiniteSize := resource.MustParse("-1")
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			RetentionTime:   &infiniteTime,
			RetentionSize:   &infiniteSize,
		},
	}
}

// MakePulsarTopicWithInfiniteRetentionTime creates a PulsarTopic with infinite retention time only
func MakePulsarTopicWithInfiniteRetentionTime(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	infiniteTime := rsutils.Duration("-1")
	finiteSize := resource.MustParse("10Gi")
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			RetentionTime:   &infiniteTime,
			RetentionSize:   &finiteSize,
		},
	}
}

// MakePulsarTopicWithInfiniteRetentionSize creates a PulsarTopic with infinite retention size only
func MakePulsarTopicWithInfiniteRetentionSize(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	finiteTime := rsutils.Duration("7d")
	infiniteSize := resource.MustParse("-1")
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			RetentionTime:   &finiteTime,
			RetentionSize:   &infiniteSize,
		},
	}
}

// MakePulsarTopicWithFiniteRetention creates a PulsarTopic with finite retention policies for comparison
func MakePulsarTopicWithFiniteRetention(namespace, name, topicName, connectionName string, policy v1alpha1.PulsarResourceLifeCyclePolicy) *v1alpha1.PulsarTopic {
	finiteTime := rsutils.Duration("24h")
	finiteSize := resource.MustParse("5Gi")
	return &v1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha1.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name:            topicName,
			LifecyclePolicy: policy,
			RetentionTime:   &finiteTime,
			RetentionSize:   &finiteSize,
		},
	}
}
