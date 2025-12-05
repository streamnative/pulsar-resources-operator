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

package operator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/streamnative/pulsar-resources-operator/pkg/connection"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	v1alphav1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	rutils "github.com/streamnative/pulsar-resources-operator/pkg/utils"
	"github.com/streamnative/pulsar-resources-operator/tests/utils"
)

type testJSON struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// validatePermissionStateAnnotation validates the PulsarPermission state annotation
func validatePermissionStateAnnotation(ctx context.Context, permissionName, namespaceName string, expected connection.PulsarPermissionState) {
	permission := &v1alphav1.PulsarPermission{}
	tns := types.NamespacedName{Namespace: namespaceName, Name: permissionName}
	Expect(k8sClient.Get(ctx, tns, permission)).Should(Succeed())

	annotations := permission.GetAnnotations()
	Expect(annotations).ShouldNot(BeNil(), "Permission should have annotations")

	stateJSON, exists := annotations[connection.PulsarPermissionStateAnnotation]
	Expect(exists).Should(BeTrue(), "Permission should have state annotation")
	Expect(stateJSON).ShouldNot(BeEmpty(), "State annotation should not be empty")

	var actualState connection.PulsarPermissionState
	err := json.Unmarshal([]byte(stateJSON), &actualState)
	Expect(err).ShouldNot(HaveOccurred(), "State annotation should be valid JSON")

	// Sort roles for consistent comparison
	expectedRoles := make([]string, len(expected.Roles))
	copy(expectedRoles, expected.Roles)
	slices.Sort(expectedRoles)

	actualRoles := make([]string, len(actualState.Roles))
	copy(actualRoles, actualState.Roles)
	slices.Sort(actualRoles)

	// Sort actions for consistent comparison
	expectedActions := make([]string, len(expected.Actions))
	copy(expectedActions, expected.Actions)
	slices.Sort(expectedActions)

	actualActions := make([]string, len(actualState.Actions))
	copy(actualActions, actualState.Actions)
	slices.Sort(actualActions)

	Expect(actualState.ResourceType).Should(Equal(expected.ResourceType),
		"ResourceType should match expected value")
	Expect(actualState.ResourceName).Should(Equal(expected.ResourceName),
		"ResourceName should match expected value")
	Expect(actualRoles).Should(Equal(expectedRoles),
		"Roles should match expected values")
	Expect(actualActions).Should(Equal(expectedActions),
		"Actions should match expected values")
}

// validateTopicPermissionsContain validates that the topic contains the specified roles
// but allows for additional roles to exist (for multi-permission testing)
func validateTopicPermissionsContain(topicName string, expectedRoles []string) {
	Eventually(func(g Gomega) {
		podName := fmt.Sprintf("%s-broker-0", brokerName)
		containerName := fmt.Sprintf("%s-broker", brokerName)
		stdout, _, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
			"./bin/pulsar-admin topics permissions "+topicName)
		g.Expect(err).Should(Succeed())
		g.Expect(stdout).Should(Not(BeEmpty()))

		// Verify all expected roles exist
		for _, role := range expectedRoles {
			g.Expect(stdout).Should(ContainSubstring(role), "Topic should contain role: "+role)
		}
	}, "20s", "100ms").Should(Succeed())
}

var _ = Describe("Resources", func() {

	var (
		ctx                 context.Context
		pconn               *v1alphav1.PulsarConnection
		pconnName           string = "test-connection"
		ptenant             *v1alphav1.PulsarTenant
		ptenantName         string = "test-tenant"
		tenantName          string = "cloud"
		lifecyclePolicy     v1alphav1.PulsarResourceLifeCyclePolicy
		pnamespace          *v1alphav1.PulsarNamespace
		pnamespaceName      string = "test-namespace"
		pulsarNamespaceName string = "cloud/stage"
		ptopic              *v1alphav1.PulsarTopic
		ptopicName          string = "test-topic"
		topicName           string = "persistent://cloud/stage/user"
		ptopic2             *v1alphav1.PulsarTopic
		ptopicName2         string = "test-topic2"
		topicName2          string = "persistent://cloud/stage/user2"
		ppermission         *v1alphav1.PulsarPermission
		ppermissionName     string = "test-permission"
		ppermission2        *v1alphav1.PulsarPermission
		ppermission2Name    string = "test-permission-2"
		ppermission3        *v1alphav1.PulsarPermission
		ppermission3Name    string = "test-permission-3"
		exampleSchemaDef           = "{\"type\":\"record\",\"name\":\"Example\",\"namespace\":\"test\"," +
			"\"fields\":[{\"name\":\"ID\",\"type\":\"int\"},{\"name\":\"Name\",\"type\":\"string\"}]}"
		partitionedTopic = &v1alphav1.PulsarTopic{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespaceName,
				Name:      "test-partitioned-topic",
			},
			Spec: v1alphav1.PulsarTopicSpec{
				Name:       "persistent://cloud/stage/partitioned-topic",
				Partitions: pointer.Int32(1),
				ConnectionRef: corev1.LocalObjectReference{
					Name: pconnName,
				},
			},
		}
		ppackage               *v1alphav1.PulsarPackage
		ppackageurl            string = "function://public/default/api-examples@v3.2.3.3"
		pfuncName              string = "test-func"
		pfuncFailureName       string = "func-test-failure"
		psinkName              string = "test-sink"
		psourceName            string = "test-source"
		pclusterName           string = "test-pulsar"
		pnsIsolationPolicyName string = "test-ns-isolation-policy"
		pfunc                  *v1alphav1.PulsarFunction
		pfuncfailure           *v1alphav1.PulsarFunction
		psinkpackageurl        string = "builtin://data-generator"
		psink                  *v1alphav1.PulsarSink
		psource                *v1alphav1.PulsarSource
		pnsisolationpolicy     *v1alphav1.PulsarNSIsolationPolicy
		psourcepackageurl      string = "builtin://data-generator"
	)

	BeforeEach(func() {
		Expect(feature.SetFeatureGates()).ShouldNot(HaveOccurred())
		ctx = context.TODO()
		// use ClusterIP svc when run operator in k8s
		adminServiceURL := utils.GetEnv("ADMIN_SERVICE_URL", fmt.Sprintf("http://%s-broker.%s.svc.cluster.local:8080", brokerName, namespaceName))
		// use NodePort svc when cluster is kind cluster and run operator locally, the nodePort need to be setup in kind
		// adminServiceURL := fmt.Sprintf("http://127.0.0.1:%d", nodePort)
		pconn = utils.MakePulsarConnection(namespaceName, pconnName, adminServiceURL)
		adminRoles := []string{"superman"}
		allowedClusters := []string{}
		lifecyclePolicy = v1alphav1.CleanUpAfterDeletion
		ptenant = utils.MakePulsarTenant(namespaceName, ptenantName, tenantName, pconnName, adminRoles, allowedClusters, lifecyclePolicy)
		pnamespace = utils.MakePulsarNamespace(namespaceName, pnamespaceName, pulsarNamespaceName, pconnName, lifecyclePolicy)
		ptopic = utils.MakePulsarTopic(namespaceName, ptopicName, topicName, pconnName, lifecyclePolicy)
		ptopic2 = utils.MakePulsarTopic(namespaceName, ptopicName2, topicName2, pconnName, lifecyclePolicy)
		roles := []string{"ironman"}
		actions := []string{"produce", "consume", "functions"}
		ppermission = utils.MakePulsarPermission(namespaceName, ppermissionName, topicName, pconnName, v1alphav1.PulsarResourceTypeTopic, roles, actions, v1alphav1.CleanUpAfterDeletion)

		// Additional permissions for multi-permission testing
		roles2 := []string{"batman"}
		actions2 := []string{"produce", "consume"}
		ppermission2 = utils.MakePulsarPermission(namespaceName, ppermission2Name, topicName, pconnName, v1alphav1.PulsarResourceTypeTopic, roles2, actions2, v1alphav1.CleanUpAfterDeletion)

		roles3 := []string{"superman"}
		actions3 := []string{"functions"}
		ppermission3 = utils.MakePulsarPermission(namespaceName, ppermission3Name, topicName, pconnName, v1alphav1.PulsarResourceTypeTopic, roles3, actions3, v1alphav1.CleanUpAfterDeletion)
		ppackage = utils.MakePulsarPackage(namespaceName, pfuncName, ppackageurl, pconnName, lifecyclePolicy)
		pfunc = utils.MakePulsarFunction(namespaceName, pfuncName, ppackageurl, pconnName, lifecyclePolicy)
		pfuncfailure = utils.MakePulsarFunction(namespaceName, pfuncFailureName, "function://not/exists/package@latest", pconnName, lifecyclePolicy)
		psink = utils.MakePulsarSink(namespaceName, psinkName, psinkpackageurl, pconnName, lifecyclePolicy)
		psource = utils.MakePulsarSource(namespaceName, psourceName, psourcepackageurl, pconnName, lifecyclePolicy)
		pnsisolationpolicy = utils.MakeNSIsolationPolicy(namespaceName, pnsIsolationPolicyName, pclusterName, pconnName,
			[]string{pnamespaceName},
			[]string{"test-pulsar-broker-0.*"},
			[]string{},
			map[string]string{
				"min_limit":       "1",
				"usage_threshold": "80",
			})
	})

	Describe("Basic resource operations", Ordered, func() {
		Context("Check pulsar broker", Label("Permissions"), func() {
			It("should create the pulsar broker successfully", func() {
				Eventually(func() bool {
					statefulset := &v1.StatefulSet{}
					k8sClient.Get(ctx, types.NamespacedName{
						Name:      brokerName + "-broker",
						Namespace: namespaceName,
					}, statefulset)
					return statefulset.Status.ReadyReplicas > 0 && statefulset.Status.ReadyReplicas == statefulset.Status.Replicas
				}, "600s", "100ms").Should(BeTrue())
			})
		})

		Context("PulsarConnection operation", Label("Permissions"), func() {
			It("should create the pulsarconnection successfully", func() {
				err := k8sClient.Create(ctx, pconn)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})
		})

		Context("PulsarTenant operation", Label("Permissions"), func() {
			It("should create the pulsartenant successfully", func() {
				err := k8sClient.Create(ctx, ptenant)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTenant{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ptenantName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "240s", "100ms").Should(BeTrue())
			})

			It("should update the pulsartenant successfully", func() {
				ptenant.Spec.AdminRoles = append(ptenant.Spec.AdminRoles, "ironman")
				err := k8sClient.Create(ctx, ptenant)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTenant{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ptenantName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "240s", "100ms").Should(BeTrue())
			})
		})

		Context("PulsarNamespace operation", Label("Permissions"), func() {
			It("should create the pulsarnamespace successfully", func() {
				err := k8sClient.Create(ctx, pnamespace)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pnamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct TopicAutoCreationConfig", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: pnamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify TopicAutoCreationConfig is set correctly
				Expect(ns.Spec.TopicAutoCreationConfig).ShouldNot(BeNil())
				Expect(ns.Spec.TopicAutoCreationConfig.Allow).Should(BeTrue())
				Expect(ns.Spec.TopicAutoCreationConfig.Type).Should(Equal("partitioned"))
				Expect(ns.Spec.TopicAutoCreationConfig.Partitions).ShouldNot(BeNil())
				Expect(*ns.Spec.TopicAutoCreationConfig.Partitions).Should(Equal(int32(10)))
			})

			It("should update TopicAutoCreationConfig successfully", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: pnamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Update TopicAutoCreationConfig
				ns.Spec.TopicAutoCreationConfig.Allow = false
				ns.Spec.TopicAutoCreationConfig.Type = "non-partitioned"
				ns.Spec.TopicAutoCreationConfig.Partitions = pointer.Int32(5)

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after TopicAutoCreationConfig update", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pnamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have updated TopicAutoCreationConfig values", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: pnamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify updated values
				Expect(ns.Spec.TopicAutoCreationConfig).ShouldNot(BeNil())
				Expect(ns.Spec.TopicAutoCreationConfig.Allow).Should(BeFalse())
				Expect(ns.Spec.TopicAutoCreationConfig.Type).Should(Equal("non-partitioned"))
				Expect(ns.Spec.TopicAutoCreationConfig.Partitions).ShouldNot(BeNil())
				Expect(*ns.Spec.TopicAutoCreationConfig.Partitions).Should(Equal(int32(5)))
			})
		})

		Context("PulsarTopic operation", Ordered, func() {
			It("should create the pulsartopic successfully", Label("Permissions"), func() {
				err := k8sClient.Create(ctx, ptopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
				err = k8sClient.Create(ctx, ptopic2)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
				err = k8sClient.Create(ctx, partitionedTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", Label("Permissions"), func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: partitionedTopic.Namespace, Name: partitionedTopic.Name}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have the schema set", func() {
				// set schema for two topics
				updateTopicSchema(ctx, ptopicName, exampleSchemaDef)
				updateTopicSchema(ctx, ptopicName2, exampleSchemaDef)

				Eventually(func(g Gomega) {
					podName := fmt.Sprintf("%s-broker-0", brokerName)
					containerName := fmt.Sprintf("%s-broker", brokerName)
					stdout, _, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
						"./bin/pulsarctl -s http://localhost:8080 --token=$PROXY_TOKEN  schemas get "+ptopic.Spec.Name)
					g.Expect(err).Should(Succeed())
					g.Expect(stdout).Should(Not(BeEmpty()))
					format.MaxLength = 0
					g.Expect(stdout).Should(ContainSubstring("JSON"))

					stdout, _, err = utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
						"./bin/pulsarctl -s http://localhost:8080 --token=$PROXY_TOKEN  schemas get "+ptopic2.Spec.Name)
					g.Expect(err).Should(Succeed())
					g.Expect(stdout).Should(Not(BeEmpty()))
					format.MaxLength = 0
					g.Expect(stdout).Should(ContainSubstring("JSON"))
				}, "20s", "100ms").Should(Succeed())
			})

			It("should always update pulsar resource when enable AlwaysUpdatePulsarResource", func() {
				podName := fmt.Sprintf("%s-broker-0", brokerName)
				containerName := fmt.Sprintf("%s-broker", brokerName)

				By("delete topic2 with pulsarctl")
				_, stderr, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
					"./bin/pulsarctl -s http://localhost:8080 --token=$PROXY_TOKEN "+
						"topics delete -f --non-partitioned "+ptopic2.Spec.Name)
				Expect(err).ShouldNot(HaveOccurred())
				format.MaxLength = 0
				Expect(stderr).Should(ContainSubstring("successfully"))

				By("delete topic1 schema in k8s")
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())
				topic.Spec.SchemaInfo = nil
				Expect(k8sClient.Update(ctx, topic)).Should(Succeed())

				// By("check topic1 schema is deleted in pulsar")
				// Eventually(func(g Gomega) {
				// 	_, stderr, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
				// 		"./bin/pulsarctl -s http://localhost:8080 --token=$PROXY_TOKEN  schemas get "+ptopic.Spec.Name)
				// 	g.Expect(err).ShouldNot(BeNil())
				// 	g.Expect(stderr).Should(Not(BeEmpty()))
				// 	format.MaxLength = 0
				// 	g.Expect(stderr).Should(ContainSubstring("404"))
				// }, "5s", "100ms").Should(Succeed())
			})

			It("should increase the partitions successfully", func() {
				curTopic := &v1alphav1.PulsarTopic{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Namespace: partitionedTopic.Namespace,
					Name:      partitionedTopic.Name,
				}, curTopic)).ShouldNot(HaveOccurred())

				curTopic.Spec.Partitions = pointer.Int32(2)
				err := k8sClient.Update(ctx, curTopic)
				Expect(err).ShouldNot(HaveOccurred())
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: curTopic.Namespace, Name: curTopic.Name}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}).Should(BeTrue())
			})

		})

		Context("PulsarTopic Compaction Threshold", Ordered, func() {
			var (
				compactionTopic     *v1alphav1.PulsarTopic
				compactionTopicName string = "test-compaction-topic"
				compactionThreshold int64  = 104857600 // 100MB in bytes
			)

			BeforeAll(func() {
				compactionTopic = utils.MakePulsarTopicWithCompactionThreshold(
					namespaceName,
					compactionTopicName,
					"persistent://public/default/compaction-test",
					pconnName,
					compactionThreshold,
					lifecyclePolicy,
				)
			})

			It("should create topic with compaction threshold successfully", func() {
				err := k8sClient.Create(ctx, compactionTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: compactionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should update compaction threshold successfully", func() {
				newThreshold := int64(209715200) // 200MB in bytes

				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: compactionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				topic.Spec.CompactionThreshold = &newThreshold
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: compactionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should remove compaction threshold when set to nil", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: compactionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				topic.Spec.CompactionThreshold = nil
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after removing threshold", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: compactionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should reset compaction state annotation after removing threshold", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: compactionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())

					annotations := t.GetAnnotations()
					if annotations == nil {
						return true
					}

					rawState, exists := annotations[connection.PulsarTopicPolicyStateAnnotation]
					if !exists {
						return true
					}

					var state map[string]interface{}
					Expect(json.Unmarshal([]byte(rawState), &state)).Should(Succeed())
					value, ok := state["compactionThreshold"]
					return !ok || value == nil
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				// Clean up the compaction test topic after all tests complete
				if compactionTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: compactionTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Persistence Policies", Ordered, func() {
			var (
				persistenceTopic     *v1alphav1.PulsarTopic
				persistenceTopicName string = "test-persistence-topic"
			)

			BeforeAll(func() {
				persistenceTopic = utils.MakePulsarTopicWithPersistencePolicies(
					namespaceName,
					persistenceTopicName,
					"persistent://public/default/persistence-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with persistence policies successfully", func() {
				err := k8sClient.Create(ctx, persistenceTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: persistenceTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct persistence policies configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: persistenceTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify PersistencePolicies
				Expect(topic.Spec.PersistencePolicies).ShouldNot(BeNil())
				Expect(*topic.Spec.PersistencePolicies.BookkeeperEnsemble).Should(Equal(int32(3)))
				Expect(*topic.Spec.PersistencePolicies.BookkeeperWriteQuorum).Should(Equal(int32(2)))
				Expect(*topic.Spec.PersistencePolicies.BookkeeperAckQuorum).Should(Equal(int32(2)))
				Expect(*topic.Spec.PersistencePolicies.ManagedLedgerMaxMarkDeleteRate).Should(Equal("2.0"))
			})

			It("should update persistence policies successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: persistenceTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update persistence policies
				topic.Spec.PersistencePolicies.BookkeeperEnsemble = pointer.Int32(5)
				topic.Spec.PersistencePolicies.BookkeeperWriteQuorum = pointer.Int32(3)
				topic.Spec.PersistencePolicies.BookkeeperAckQuorum = pointer.Int32(3)
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: persistenceTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if persistenceTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: persistenceTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Delayed Delivery", Ordered, func() {
			var (
				delayedDeliveryTopic     *v1alphav1.PulsarTopic
				delayedDeliveryTopicName string = "test-delayed-delivery-topic"
			)

			BeforeAll(func() {
				delayedDeliveryTopic = utils.MakePulsarTopicWithDelayedDelivery(
					namespaceName,
					delayedDeliveryTopicName,
					"persistent://public/default/delayed-delivery-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with delayed delivery successfully", func() {
				err := k8sClient.Create(ctx, delayedDeliveryTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: delayedDeliveryTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct delayed delivery configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: delayedDeliveryTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify DelayedDelivery
				Expect(topic.Spec.DelayedDelivery).ShouldNot(BeNil())
				Expect(*topic.Spec.DelayedDelivery.Active).Should(Equal(true))
				Expect(*topic.Spec.DelayedDelivery.TickTimeMillis).Should(Equal(int64(1000)))
			})

			It("should update delayed delivery configuration successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: delayedDeliveryTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update delayed delivery
				topic.Spec.DelayedDelivery.TickTimeMillis = pointer.Int64(2000)
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: delayedDeliveryTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if delayedDeliveryTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: delayedDeliveryTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Dispatch Rate", Ordered, func() {
			var (
				dispatchRateTopic     *v1alphav1.PulsarTopic
				dispatchRateTopicName string = "test-dispatch-rate-topic"
			)

			BeforeAll(func() {
				dispatchRateTopic = utils.MakePulsarTopicWithDispatchRate(
					namespaceName,
					dispatchRateTopicName,
					"persistent://public/default/dispatch-rate-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with dispatch rate successfully", func() {
				err := k8sClient.Create(ctx, dispatchRateTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: dispatchRateTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct dispatch rate configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: dispatchRateTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify DispatchRate
				Expect(topic.Spec.DispatchRate).ShouldNot(BeNil())
				Expect(*topic.Spec.DispatchRate.DispatchThrottlingRateInMsg).Should(Equal(int32(500)))
				Expect(*topic.Spec.DispatchRate.DispatchThrottlingRateInByte).Should(Equal(int64(524288)))
				Expect(*topic.Spec.DispatchRate.RatePeriodInSecond).Should(Equal(int32(1)))
			})

			It("should update dispatch rate successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: dispatchRateTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update dispatch rate
				topic.Spec.DispatchRate.DispatchThrottlingRateInMsg = pointer.Int32(1000)
				topic.Spec.DispatchRate.DispatchThrottlingRateInByte = pointer.Int64(1048576)
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: dispatchRateTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if dispatchRateTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: dispatchRateTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Publish Rate", Ordered, func() {
			var (
				publishRateTopic     *v1alphav1.PulsarTopic
				publishRateTopicName string = "test-publish-rate-topic"
			)

			BeforeAll(func() {
				publishRateTopic = utils.MakePulsarTopicWithPublishRate(
					namespaceName,
					publishRateTopicName,
					"persistent://public/default/publish-rate-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with publish rate successfully", func() {
				err := k8sClient.Create(ctx, publishRateTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: publishRateTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct publish rate configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: publishRateTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify PublishRate
				Expect(topic.Spec.PublishRate).ShouldNot(BeNil())
				Expect(*topic.Spec.PublishRate.PublishThrottlingRateInMsg).Should(Equal(int32(1000)))
				Expect(*topic.Spec.PublishRate.PublishThrottlingRateInByte).Should(Equal(int64(1048576)))
			})

			It("should update publish rate successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: publishRateTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update publish rate
				topic.Spec.PublishRate.PublishThrottlingRateInMsg = pointer.Int32(2000)
				topic.Spec.PublishRate.PublishThrottlingRateInByte = pointer.Int64(2097152)
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: publishRateTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if publishRateTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: publishRateTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Inactive Topic Policies", Ordered, func() {
			var (
				inactiveTopicPoliciesTopic     *v1alphav1.PulsarTopic
				inactiveTopicPoliciesTopicName string = "test-inactive-topic-policies-topic"
			)

			BeforeAll(func() {
				inactiveTopicPoliciesTopic = utils.MakePulsarTopicWithInactiveTopicPolicies(
					namespaceName,
					inactiveTopicPoliciesTopicName,
					"persistent://public/default/inactive-topic-policies-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with inactive topic policies successfully", func() {
				err := k8sClient.Create(ctx, inactiveTopicPoliciesTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: inactiveTopicPoliciesTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct inactive topic policies configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: inactiveTopicPoliciesTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify InactiveTopicPolicies
				Expect(topic.Spec.InactiveTopicPolicies).ShouldNot(BeNil())
				Expect(*topic.Spec.InactiveTopicPolicies.InactiveTopicDeleteMode).Should(Equal("delete_when_no_subscriptions"))
				Expect(*topic.Spec.InactiveTopicPolicies.MaxInactiveDurationInSeconds).Should(Equal(int32(1800)))
				Expect(*topic.Spec.InactiveTopicPolicies.DeleteWhileInactive).Should(Equal(true))
			})

			It("should update inactive topic policies successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: inactiveTopicPoliciesTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update inactive topic policies
				topic.Spec.InactiveTopicPolicies.MaxInactiveDurationInSeconds = pointer.Int32(3600)
				topic.Spec.InactiveTopicPolicies.DeleteWhileInactive = pointer.Bool(false)
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: inactiveTopicPoliciesTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if inactiveTopicPoliciesTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: inactiveTopicPoliciesTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Subscribe Rate", Ordered, func() {
			var (
				subscribeRateTopic     *v1alphav1.PulsarTopic
				subscribeRateTopicName string = "test-subscribe-rate-topic"
			)

			BeforeAll(func() {
				subscribeRateTopic = utils.MakePulsarTopicWithSubscribeRate(
					namespaceName,
					subscribeRateTopicName,
					"persistent://public/default/subscribe-rate-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with subscribe rate successfully", func() {
				err := k8sClient.Create(ctx, subscribeRateTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: subscribeRateTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct subscribe rate configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: subscribeRateTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify SubscribeRate
				Expect(topic.Spec.SubscribeRate).ShouldNot(BeNil())
				Expect(*topic.Spec.SubscribeRate.SubscribeThrottlingRatePerConsumer).Should(Equal(int32(10)))
				Expect(*topic.Spec.SubscribeRate.RatePeriodInSecond).Should(Equal(int32(30)))
			})

			It("should update subscribe rate successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: subscribeRateTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update subscribe rate
				topic.Spec.SubscribeRate.SubscribeThrottlingRatePerConsumer = pointer.Int32(20)
				topic.Spec.SubscribeRate.RatePeriodInSecond = pointer.Int32(60)
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: subscribeRateTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if subscribeRateTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: subscribeRateTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Offload Policies", Ordered, func() {
			var (
				offloadPoliciesTopic     *v1alphav1.PulsarTopic
				offloadPoliciesTopicName string = "test-offload-policies-topic"
			)

			BeforeAll(func() {
				offloadPoliciesTopic = utils.MakePulsarTopicWithOffloadPolicies(
					namespaceName,
					offloadPoliciesTopicName,
					"persistent://public/default/offload-policies-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with offload policies successfully", func() {
				err := k8sClient.Create(ctx, offloadPoliciesTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: offloadPoliciesTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct offload policies configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: offloadPoliciesTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify OffloadPolicies
				Expect(topic.Spec.OffloadPolicies).ShouldNot(BeNil())
				Expect(topic.Spec.OffloadPolicies.ManagedLedgerOffloadDriver).Should(Equal("aws-s3"))
				Expect(topic.Spec.OffloadPolicies.ManagedLedgerOffloadMaxThreads).Should(Equal(5))
				Expect(topic.Spec.OffloadPolicies.ManagedLedgerOffloadThresholdInBytes).Should(Equal(int64(1073741824)))
			})

			It("should update offload policies successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: offloadPoliciesTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update offload policies
				topic.Spec.OffloadPolicies.ManagedLedgerOffloadMaxThreads = 10
				topic.Spec.OffloadPolicies.ManagedLedgerOffloadThresholdInBytes = 2147483648 // 2GB
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: offloadPoliciesTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if offloadPoliciesTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: offloadPoliciesTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Auto Subscription Creation", Ordered, func() {
			var (
				autoSubscriptionTopic     *v1alphav1.PulsarTopic
				autoSubscriptionTopicName string = "test-auto-subscription-topic"
			)

			BeforeAll(func() {
				autoSubscriptionTopic = utils.MakePulsarTopicWithAutoSubscriptionCreation(
					namespaceName,
					autoSubscriptionTopicName,
					"persistent://public/default/auto-subscription-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with auto subscription creation successfully", func() {
				err := k8sClient.Create(ctx, autoSubscriptionTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: autoSubscriptionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct auto subscription creation configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: autoSubscriptionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify AutoSubscriptionCreation
				Expect(topic.Spec.AutoSubscriptionCreation).ShouldNot(BeNil())
				Expect(topic.Spec.AutoSubscriptionCreation.AllowAutoSubscriptionCreation).Should(Equal(true))
			})

			It("should update auto subscription creation successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: autoSubscriptionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update auto subscription creation
				topic.Spec.AutoSubscriptionCreation.AllowAutoSubscriptionCreation = false
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: autoSubscriptionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if autoSubscriptionTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: autoSubscriptionTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Schema Compatibility Strategy", Ordered, func() {
			var (
				schemaCompatibilityTopic     *v1alphav1.PulsarTopic
				schemaCompatibilityTopicName string = "test-schema-compatibility-topic"
			)

			BeforeAll(func() {
				schemaCompatibilityTopic = utils.MakePulsarTopicWithSchemaCompatibilityStrategy(
					namespaceName,
					schemaCompatibilityTopicName,
					"persistent://public/default/schema-compatibility-test",
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with schema compatibility strategy successfully", func() {
				err := k8sClient.Create(ctx, schemaCompatibilityTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: schemaCompatibilityTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct schema compatibility strategy configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: schemaCompatibilityTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify SchemaCompatibilityStrategy
				Expect(topic.Spec.SchemaCompatibilityStrategy).ShouldNot(BeNil())
				Expect(string(*topic.Spec.SchemaCompatibilityStrategy)).Should(Equal("BACKWARD"))
			})

			It("should update schema compatibility strategy successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: schemaCompatibilityTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update schema compatibility strategy
				newStrategy := v1alphav1.SchemaCompatibilityStrategy("FORWARD")
				topic.Spec.SchemaCompatibilityStrategy = &newStrategy
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: schemaCompatibilityTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			AfterAll(func() {
				if schemaCompatibilityTopic != nil {
					Eventually(func(g Gomega) {
						t := &v1alphav1.PulsarTopic{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: schemaCompatibilityTopicName}
						g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarTopic Infinite Retention Policies", Ordered, func() {
			var (
				infiniteRetentionTopic         *v1alphav1.PulsarTopic
				infiniteRetentionTopicName     string = "test-infinite-retention-topic"
				infiniteRetentionFullTopicName string = "persistent://cloud/stage/infinite-retention-test"

				infiniteTimeTopic         *v1alphav1.PulsarTopic
				infiniteTimeTopicName     string = "test-infinite-time-topic"
				infiniteTimeFullTopicName string = "persistent://cloud/stage/infinite-time-test"

				infiniteSizeTopic         *v1alphav1.PulsarTopic
				infiniteSizeTopicName     string = "test-infinite-size-topic"
				infiniteSizeFullTopicName string = "persistent://cloud/stage/infinite-size-test"

				finiteRetentionTopic         *v1alphav1.PulsarTopic
				finiteRetentionTopicName     string = "test-finite-retention-topic"
				finiteRetentionFullTopicName string = "persistent://cloud/stage/finite-retention-test"
			)

			BeforeAll(func() {
				// Create topic with both infinite retention time and size
				infiniteRetentionTopic = utils.MakePulsarTopicWithInfiniteRetention(
					namespaceName,
					infiniteRetentionTopicName,
					infiniteRetentionFullTopicName,
					pconnName,
					lifecyclePolicy,
				)

				// Create topic with infinite retention time only
				infiniteTimeTopic = utils.MakePulsarTopicWithInfiniteRetentionTime(
					namespaceName,
					infiniteTimeTopicName,
					infiniteTimeFullTopicName,
					pconnName,
					lifecyclePolicy,
				)

				// Create topic with infinite retention size only
				infiniteSizeTopic = utils.MakePulsarTopicWithInfiniteRetentionSize(
					namespaceName,
					infiniteSizeTopicName,
					infiniteSizeFullTopicName,
					pconnName,
					lifecyclePolicy,
				)

				// Create topic with finite retention for comparison
				finiteRetentionTopic = utils.MakePulsarTopicWithFiniteRetention(
					namespaceName,
					finiteRetentionTopicName,
					finiteRetentionFullTopicName,
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create topic with infinite retention (both time and size) successfully", func() {
				err := k8sClient.Create(ctx, infiniteRetentionTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: infiniteRetentionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct infinite retention configuration (both time and size)", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: infiniteRetentionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify infinite retention time
				Expect(topic.Spec.RetentionTime).ShouldNot(BeNil())
				Expect(string(*topic.Spec.RetentionTime)).Should(Equal("-1"))
				Expect(topic.Spec.RetentionTime.IsInfinite()).Should(BeTrue())

				// Verify infinite retention size
				Expect(topic.Spec.RetentionSize).ShouldNot(BeNil())
				Expect(topic.Spec.RetentionSize.String()).Should(Equal("-1"))
				Expect(rutils.IsInfiniteQuantity(topic.Spec.RetentionSize)).Should(BeTrue())
			})

			It("should create topic with infinite retention time only successfully", func() {
				err := k8sClient.Create(ctx, infiniteTimeTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: infiniteTimeTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct infinite retention time and finite size configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: infiniteTimeTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify infinite retention time
				Expect(topic.Spec.RetentionTime).ShouldNot(BeNil())
				Expect(string(*topic.Spec.RetentionTime)).Should(Equal("-1"))
				Expect(topic.Spec.RetentionTime.IsInfinite()).Should(BeTrue())

				// Verify finite retention size
				Expect(topic.Spec.RetentionSize).ShouldNot(BeNil())
				Expect(topic.Spec.RetentionSize.String()).Should(Equal("10Gi"))
				Expect(rutils.IsInfiniteQuantity(topic.Spec.RetentionSize)).Should(BeFalse())
			})

			It("should create topic with infinite retention size only successfully", func() {
				err := k8sClient.Create(ctx, infiniteSizeTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: infiniteSizeTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct finite retention time and infinite size configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: infiniteSizeTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify finite retention time
				Expect(topic.Spec.RetentionTime).ShouldNot(BeNil())
				Expect(string(*topic.Spec.RetentionTime)).Should(Equal("7d"))
				Expect(topic.Spec.RetentionTime.IsInfinite()).Should(BeFalse())

				// Verify infinite retention size
				Expect(topic.Spec.RetentionSize).ShouldNot(BeNil())
				Expect(topic.Spec.RetentionSize.String()).Should(Equal("-1"))
				Expect(rutils.IsInfiniteQuantity(topic.Spec.RetentionSize)).Should(BeTrue())
			})

			It("should create topic with finite retention for comparison successfully", func() {
				err := k8sClient.Create(ctx, finiteRetentionTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: finiteRetentionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct finite retention configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: finiteRetentionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify finite retention time
				Expect(topic.Spec.RetentionTime).ShouldNot(BeNil())
				Expect(string(*topic.Spec.RetentionTime)).Should(Equal("24h"))
				Expect(topic.Spec.RetentionTime.IsInfinite()).Should(BeFalse())

				// Verify finite retention size
				Expect(topic.Spec.RetentionSize).ShouldNot(BeNil())
				Expect(topic.Spec.RetentionSize.String()).Should(Equal("5Gi"))
				Expect(rutils.IsInfiniteQuantity(topic.Spec.RetentionSize)).Should(BeFalse())
			})

			It("should update finite retention to infinite retention successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: finiteRetentionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update to infinite retention
				infiniteTime := rutils.Duration("-1")
				infiniteSize := resource.MustParse("-1")
				topic.Spec.RetentionTime = &infiniteTime
				topic.Spec.RetentionSize = &infiniteSize
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update to infinite retention", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: finiteRetentionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have updated infinite retention configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: finiteRetentionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify updated infinite retention time
				Expect(topic.Spec.RetentionTime).ShouldNot(BeNil())
				Expect(string(*topic.Spec.RetentionTime)).Should(Equal("-1"))
				Expect(topic.Spec.RetentionTime.IsInfinite()).Should(BeTrue())

				// Verify updated infinite retention size
				Expect(topic.Spec.RetentionSize).ShouldNot(BeNil())
				Expect(topic.Spec.RetentionSize.String()).Should(Equal("-1"))
				Expect(rutils.IsInfiniteQuantity(topic.Spec.RetentionSize)).Should(BeTrue())
			})

			It("should update infinite retention back to finite retention successfully", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: finiteRetentionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Update back to finite retention
				finiteTime := rutils.Duration("12h")
				finiteSize := resource.MustParse("2Gi")
				topic.Spec.RetentionTime = &finiteTime
				topic.Spec.RetentionSize = &finiteSize
				err := k8sClient.Update(ctx, topic)
				Expect(err).Should(Succeed())
			})

			It("should be ready after update back to finite retention", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: finiteRetentionTopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have updated finite retention configuration", func() {
				topic := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: finiteRetentionTopicName}
				Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())

				// Verify updated finite retention time
				Expect(topic.Spec.RetentionTime).ShouldNot(BeNil())
				Expect(string(*topic.Spec.RetentionTime)).Should(Equal("12h"))
				Expect(topic.Spec.RetentionTime.IsInfinite()).Should(BeFalse())

				// Verify updated finite retention size
				Expect(topic.Spec.RetentionSize).ShouldNot(BeNil())
				Expect(topic.Spec.RetentionSize.String()).Should(Equal("2Gi"))
				Expect(rutils.IsInfiniteQuantity(topic.Spec.RetentionSize)).Should(BeFalse())
			})

			AfterAll(func() {
				// Clean up all test topics
				topics := []*v1alphav1.PulsarTopic{
					infiniteRetentionTopic,
					infiniteTimeTopic,
					infiniteSizeTopic,
					finiteRetentionTopic,
				}
				topicNames := []string{
					infiniteRetentionTopicName,
					infiniteTimeTopicName,
					infiniteSizeTopicName,
					finiteRetentionTopicName,
				}

				for i, topic := range topics {
					if topic != nil {
						Eventually(func(g Gomega) {
							t := &v1alphav1.PulsarTopic{}
							tns := types.NamespacedName{Namespace: namespaceName, Name: topicNames[i]}
							g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
							g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
						}).Should(Succeed())
					}
				}
			})
		})

		Context("PulsarNamespace Rate Limiting", Ordered, func() {
			var (
				rateLimitingNamespace     *v1alphav1.PulsarNamespace
				rateLimitingNamespaceName string = "test-ratelimiting-namespace"
				rateLimitingPulsarNSName  string = "cloud/ratelimiting"
			)

			BeforeAll(func() {
				rateLimitingNamespace = utils.MakePulsarNamespaceWithRateLimiting(
					namespaceName,
					rateLimitingNamespaceName,
					rateLimitingPulsarNSName,
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create namespace with rate limiting configuration successfully", func() {
				err := k8sClient.Create(ctx, rateLimitingNamespace)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should have correct dispatch rate configuration", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify DispatchRate
				Expect(ns.Spec.DispatchRate).ShouldNot(BeNil())
				Expect(*ns.Spec.DispatchRate.DispatchThrottlingRateInMsg).Should(Equal(int32(1000)))
				Expect(*ns.Spec.DispatchRate.DispatchThrottlingRateInByte).Should(Equal(int64(1048576)))
				Expect(*ns.Spec.DispatchRate.RatePeriodInSecond).Should(Equal(int32(1)))
			})

			It("should have correct subscription dispatch rate configuration", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify SubscriptionDispatchRate
				Expect(ns.Spec.SubscriptionDispatchRate).ShouldNot(BeNil())
				Expect(*ns.Spec.SubscriptionDispatchRate.DispatchThrottlingRateInMsg).Should(Equal(int32(500)))
				Expect(*ns.Spec.SubscriptionDispatchRate.DispatchThrottlingRateInByte).Should(Equal(int64(524288)))
				Expect(*ns.Spec.SubscriptionDispatchRate.RatePeriodInSecond).Should(Equal(int32(1)))
			})

			It("should have correct replicator dispatch rate configuration", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify ReplicatorDispatchRate
				Expect(ns.Spec.ReplicatorDispatchRate).ShouldNot(BeNil())
				Expect(*ns.Spec.ReplicatorDispatchRate.DispatchThrottlingRateInMsg).Should(Equal(int32(800)))
				Expect(*ns.Spec.ReplicatorDispatchRate.DispatchThrottlingRateInByte).Should(Equal(int64(838860)))
				Expect(*ns.Spec.ReplicatorDispatchRate.RatePeriodInSecond).Should(Equal(int32(1)))
			})

			It("should have correct publish rate configuration", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify PublishRate
				Expect(ns.Spec.PublishRate).ShouldNot(BeNil())
				Expect(*ns.Spec.PublishRate.PublishThrottlingRateInMsg).Should(Equal(int32(2000)))
				Expect(*ns.Spec.PublishRate.PublishThrottlingRateInByte).Should(Equal(int64(2097152)))
			})

			It("should have correct subscribe rate configuration", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify SubscribeRate
				Expect(ns.Spec.SubscribeRate).ShouldNot(BeNil())
				Expect(*ns.Spec.SubscribeRate.SubscribeThrottlingRatePerConsumer).Should(Equal(int32(10)))
				Expect(*ns.Spec.SubscribeRate.RatePeriodInSecond).Should(Equal(int32(30)))
			})

			It("should update dispatch rate successfully", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Update DispatchRate
				ns.Spec.DispatchRate.DispatchThrottlingRateInMsg = pointer.Int32(1500)
				ns.Spec.DispatchRate.DispatchThrottlingRateInByte = pointer.Int64(1572864) // 1.5MB

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after rate update", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should have updated dispatch rate values", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify updated DispatchRate
				Expect(ns.Spec.DispatchRate).ShouldNot(BeNil())
				Expect(*ns.Spec.DispatchRate.DispatchThrottlingRateInMsg).Should(Equal(int32(1500)))
				Expect(*ns.Spec.DispatchRate.DispatchThrottlingRateInByte).Should(Equal(int64(1572864)))
			})

			AfterAll(func() {
				// Clean up the rate limiting namespace
				if rateLimitingNamespace != nil {
					Eventually(func(g Gomega) {
						ns := &v1alphav1.PulsarNamespace{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: rateLimitingNamespaceName}
						g.Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, ns)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarNamespace Storage Policies", Ordered, func() {
			var (
				storagePoliciesNamespace     *v1alphav1.PulsarNamespace
				storagePoliciesNamespaceName string = "test-storage-namespace"
				storagePoliciesPulsarNSName  string = "cloud/storage"
			)

			BeforeAll(func() {
				storagePoliciesNamespace = utils.MakePulsarNamespaceWithStoragePolicies(
					namespaceName,
					storagePoliciesNamespaceName,
					storagePoliciesPulsarNSName,
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create namespace with storage policies successfully", func() {
				err := k8sClient.Create(ctx, storagePoliciesNamespace)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should have correct persistence policies configuration", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify PersistencePolicies
				Expect(ns.Spec.PersistencePolicies).ShouldNot(BeNil())
				Expect(*ns.Spec.PersistencePolicies.BookkeeperEnsemble).Should(Equal(int32(5)))
				Expect(*ns.Spec.PersistencePolicies.BookkeeperWriteQuorum).Should(Equal(int32(3)))
				Expect(*ns.Spec.PersistencePolicies.BookkeeperAckQuorum).Should(Equal(int32(2)))
				Expect(*ns.Spec.PersistencePolicies.ManagedLedgerMaxMarkDeleteRate).Should(Equal("1.5"))
			})

			It("should have correct retention policies", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify retention policies
				Expect(ns.Spec.RetentionTime).ShouldNot(BeNil())
				Expect(string(*ns.Spec.RetentionTime)).Should(Equal("48h"))
				Expect(ns.Spec.RetentionSize).ShouldNot(BeNil())

				// Verify backlog quota
				Expect(ns.Spec.BacklogQuotaLimitSize).ShouldNot(BeNil())
				Expect(ns.Spec.BacklogQuotaLimitTime).ShouldNot(BeNil())
				Expect(ns.Spec.BacklogQuotaRetentionPolicy).ShouldNot(BeNil())
				Expect(*ns.Spec.BacklogQuotaRetentionPolicy).Should(Equal("producer_request_hold"))
				Expect(ns.Spec.BacklogQuotaType).ShouldNot(BeNil())
				Expect(*ns.Spec.BacklogQuotaType).Should(Equal("destination_storage"))
			})

			It("should have correct compaction threshold", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify CompactionThreshold
				Expect(ns.Spec.CompactionThreshold).ShouldNot(BeNil())
				Expect(*ns.Spec.CompactionThreshold).Should(Equal(int64(104857600))) // 100MB
			})

			It("should have correct inactive topic policies", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify InactiveTopicPolicies
				Expect(ns.Spec.InactiveTopicPolicies).ShouldNot(BeNil())
				Expect(*ns.Spec.InactiveTopicPolicies.InactiveTopicDeleteMode).Should(Equal("delete_when_no_subscriptions"))
				Expect(*ns.Spec.InactiveTopicPolicies.MaxInactiveDurationInSeconds).Should(Equal(int32(3600)))
				Expect(*ns.Spec.InactiveTopicPolicies.DeleteWhileInactive).Should(BeTrue())
			})

			It("should have correct subscription expiration time", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify SubscriptionExpirationTime
				Expect(ns.Spec.SubscriptionExpirationTime).ShouldNot(BeNil())
				Expect(string(*ns.Spec.SubscriptionExpirationTime)).Should(Equal("7d"))
			})

			It("should reflect subscription expiration time in broker", func() {
				podName := fmt.Sprintf("%s-broker-0", brokerName)
				containerName := fmt.Sprintf("%s-broker", brokerName)
				Eventually(func(g Gomega) {
					stdout, _, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
						"./bin/pulsar-admin namespaces get-subscription-expiration-time "+storagePoliciesPulsarNSName)
					g.Expect(err).Should(Succeed())
					// 7d -> 10080 minutes
					g.Expect(stdout).Should(ContainSubstring("10080"))
				}, "30s", "200ms").Should(Succeed())
			})

			It("should update subscription expiration time to infinite", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				var infiniteExp rutils.Duration = "-1"
				ns.Spec.SubscriptionExpirationTime = &infiniteExp

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after subscription expiration update", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should reflect updated subscription expiration time", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				Expect(ns.Spec.SubscriptionExpirationTime).ShouldNot(BeNil())
				Expect(string(*ns.Spec.SubscriptionExpirationTime)).Should(Equal("-1"))
			})

			It("should reflect updated subscription expiration time in broker", func() {
				podName := fmt.Sprintf("%s-broker-0", brokerName)
				containerName := fmt.Sprintf("%s-broker", brokerName)
				Eventually(func(g Gomega) {
					stdout, _, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
						"./bin/pulsar-admin namespaces get-subscription-expiration-time "+storagePoliciesPulsarNSName)
					g.Expect(err).Should(Succeed())
					g.Expect(stdout).Should(ContainSubstring("-1"))
				}, "30s", "200ms").Should(Succeed())
			})

			It("should have correct custom properties", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify Properties
				Expect(ns.Spec.Properties).ShouldNot(BeNil())
				Expect(ns.Spec.Properties["environment"]).Should(Equal("test"))
				Expect(ns.Spec.Properties["team"]).Should(Equal("qa"))
			})

			It("should reconcile when backlog quota retention policy is unset", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				ns.Spec.BacklogQuotaRetentionPolicy = nil

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after removing backlog quota retention policy", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should allow clearing inactive topic policies and properties", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				ns.Spec.InactiveTopicPolicies = nil
				ns.Spec.Properties = nil

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after clearing inactive topic policies and properties", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should reflect cleared inactive topic policies and properties", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				Expect(ns.Spec.InactiveTopicPolicies).Should(BeNil())
				Expect(ns.Spec.Properties).Should(BeNil())
			})

			It("should update persistence policies successfully", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Update PersistencePolicies
				ns.Spec.PersistencePolicies.BookkeeperEnsemble = pointer.Int32(7)
				ns.Spec.PersistencePolicies.BookkeeperWriteQuorum = pointer.Int32(4)
				ns.Spec.PersistencePolicies.BookkeeperAckQuorum = pointer.Int32(3)

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after persistence policies update", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should have updated persistence policies values", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify updated PersistencePolicies
				Expect(ns.Spec.PersistencePolicies).ShouldNot(BeNil())
				Expect(*ns.Spec.PersistencePolicies.BookkeeperEnsemble).Should(Equal(int32(7)))
				Expect(*ns.Spec.PersistencePolicies.BookkeeperWriteQuorum).Should(Equal(int32(4)))
				Expect(*ns.Spec.PersistencePolicies.BookkeeperAckQuorum).Should(Equal(int32(3)))
			})

			It("should update compaction threshold successfully", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Update CompactionThreshold
				newThreshold := int64(209715200) // 200MB
				ns.Spec.CompactionThreshold = &newThreshold

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after compaction threshold update", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should have updated compaction threshold value", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify updated CompactionThreshold
				Expect(ns.Spec.CompactionThreshold).ShouldNot(BeNil())
				Expect(*ns.Spec.CompactionThreshold).Should(Equal(int64(209715200))) // 200MB
			})

			AfterAll(func() {
				// Clean up the storage policies namespace
				if storagePoliciesNamespace != nil {
					Eventually(func(g Gomega) {
						ns := &v1alphav1.PulsarNamespace{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: storagePoliciesNamespaceName}
						g.Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, ns)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarNamespace Security Configuration", Ordered, func() {
			var (
				securityNamespace     *v1alphav1.PulsarNamespace
				securityNamespaceName string = "test-security-namespace"
				securityPulsarNSName  string = "cloud/security"
			)

			BeforeAll(func() {
				securityNamespace = utils.MakePulsarNamespaceWithSecurityConfig(
					namespaceName,
					securityNamespaceName,
					securityPulsarNSName,
					pconnName,
					lifecyclePolicy,
				)
			})

			It("should create namespace with security configuration successfully", func() {
				err := k8sClient.Create(ctx, securityNamespace)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should have correct security configuration", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify security settings
				Expect(ns.Spec.EncryptionRequired).ShouldNot(BeNil())
				Expect(*ns.Spec.EncryptionRequired).Should(BeTrue())
				Expect(ns.Spec.ValidateProducerName).ShouldNot(BeNil())
				Expect(*ns.Spec.ValidateProducerName).Should(BeTrue())
				Expect(ns.Spec.IsAllowAutoUpdateSchema).ShouldNot(BeNil())
				Expect(*ns.Spec.IsAllowAutoUpdateSchema).Should(BeFalse())
			})

			It("should have correct anti-affinity configuration", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify anti-affinity group
				Expect(ns.Spec.AntiAffinityGroup).ShouldNot(BeNil())
				Expect(*ns.Spec.AntiAffinityGroup).Should(Equal("high-availability"))
			})

			It("should have correct schema validation enforcement", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify schema validation
				Expect(ns.Spec.SchemaValidationEnforced).ShouldNot(BeNil())
				Expect(*ns.Spec.SchemaValidationEnforced).Should(BeTrue())
			})

			It("should have security-focused rate limiting", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify security-focused rate limits
				Expect(ns.Spec.PublishRate).ShouldNot(BeNil())
				Expect(*ns.Spec.PublishRate.PublishThrottlingRateInMsg).Should(Equal(int32(1000)))
				Expect(*ns.Spec.PublishRate.PublishThrottlingRateInByte).Should(Equal(int64(1048576)))

				Expect(ns.Spec.SubscribeRate).ShouldNot(BeNil())
				Expect(*ns.Spec.SubscribeRate.SubscribeThrottlingRatePerConsumer).Should(Equal(int32(5)))
				Expect(*ns.Spec.SubscribeRate.RatePeriodInSecond).Should(Equal(int32(60)))
			})

			It("should have security-focused storage policies", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify high-replication persistence for security
				Expect(ns.Spec.PersistencePolicies).ShouldNot(BeNil())
				Expect(*ns.Spec.PersistencePolicies.BookkeeperEnsemble).Should(Equal(int32(5)))
				Expect(*ns.Spec.PersistencePolicies.BookkeeperWriteQuorum).Should(Equal(int32(3)))
				Expect(*ns.Spec.PersistencePolicies.BookkeeperAckQuorum).Should(Equal(int32(3))) // Higher ack quorum for security

				// Verify long retention for audit
				Expect(ns.Spec.RetentionTime).ShouldNot(BeNil())
				Expect(string(*ns.Spec.RetentionTime)).Should(Equal("720h")) // 30 days
			})

			It("should have security metadata in properties", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify security properties
				Expect(ns.Spec.Properties).ShouldNot(BeNil())
				Expect(ns.Spec.Properties["security-level"]).Should(Equal("high"))
				Expect(ns.Spec.Properties["compliance"]).Should(Equal("required"))
				Expect(ns.Spec.Properties["data-classification"]).Should(Equal("confidential"))
			})

			It("should have deduplication enabled", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify deduplication for data integrity
				Expect(ns.Spec.Deduplication).ShouldNot(BeNil())
				Expect(*ns.Spec.Deduplication).Should(BeTrue())
			})

			It("should update encryption requirement successfully", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Temporarily disable encryption requirement
				ns.Spec.EncryptionRequired = pointer.Bool(false)

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after security config update", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should have updated encryption requirement", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify updated encryption requirement
				Expect(ns.Spec.EncryptionRequired).ShouldNot(BeNil())
				Expect(*ns.Spec.EncryptionRequired).Should(BeFalse())
			})

			It("should update producer name validation successfully", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Disable producer name validation
				ns.Spec.ValidateProducerName = pointer.Bool(false)

				err := k8sClient.Update(ctx, ns)
				Expect(err).Should(Succeed())
			})

			It("should be ready after producer validation update", func() {
				Eventually(func() bool {
					ns := &v1alphav1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(ns)
				}, "30s", "100ms").Should(BeTrue())
			})

			It("should have updated producer name validation", func() {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
				Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())

				// Verify updated producer name validation
				Expect(ns.Spec.ValidateProducerName).ShouldNot(BeNil())
				Expect(*ns.Spec.ValidateProducerName).Should(BeFalse())
			})

			AfterAll(func() {
				// Clean up the security namespace
				if securityNamespace != nil {
					Eventually(func(g Gomega) {
						ns := &v1alphav1.PulsarNamespace{}
						tns := types.NamespacedName{Namespace: namespaceName, Name: securityNamespaceName}
						g.Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
						g.Expect(k8sClient.Delete(ctx, ns)).Should(Succeed())
					}).Should(Succeed())
				}
			})
		})

		Context("PulsarPermission operation", Label("Permissions"), func() {
			It("should grant the pulsarpermission successfully", func() {
				err := k8sClient.Create(ctx, ppermission)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarPermission{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ppermissionName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct initial state annotation", func() {
				Eventually(func() error {
					expectedState := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"ironman"},
						Actions:      []string{"produce", "consume", "functions"},
					}

					validatePermissionStateAnnotation(ctx, ppermissionName, namespaceName, expectedState)
					return nil
				}, "20s", "100ms").Should(Succeed())
			})

			It("should add a new role", func() {
				t := &v1alphav1.PulsarPermission{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ppermissionName}
				Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
				t.Spec.Roles = append(t.Spec.Roles, "spiderman")
				err := k8sClient.Update(ctx, t)
				Expect(err).Should(Succeed())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarPermission{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ppermissionName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("spiderman should exists along with ironman", func() {
				Eventually(func(g Gomega) {
					podName := fmt.Sprintf("%s-broker-0", brokerName)
					containerName := fmt.Sprintf("%s-broker", brokerName)
					stdout, _, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
						"./bin/pulsar-admin topics permissions "+ppermission.Spec.ResourceName)
					g.Expect(err).Should(Succeed())
					g.Expect(stdout).Should(Not(BeEmpty()))
					g.Expect(stdout).Should(ContainSubstring("ironman"))
					g.Expect(stdout).Should(ContainSubstring("spiderman"))
				}, "20s", "100ms").Should(Succeed())
			})

			It("should have updated state annotation with both roles", func() {
				Eventually(func() error {
					expectedState := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"ironman", "spiderman"},
						Actions:      []string{"produce", "consume", "functions"},
					}

					validatePermissionStateAnnotation(ctx, ppermissionName, namespaceName, expectedState)
					return nil
				}, "20s", "100ms").Should(Succeed())
			})

			It("should delete a role", func() {
				t := &v1alphav1.PulsarPermission{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ppermissionName}
				Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
				// remove spiderman and assign to roles
				t.Spec.Roles = []string{"ironman"}
				err := k8sClient.Update(ctx, t)
				Expect(err).Should(Succeed())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarPermission{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ppermissionName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("spiderman shouldn't exists anymore but ironman", func() {
				Eventually(func(g Gomega) {
					podName := fmt.Sprintf("%s-broker-0", brokerName)
					containerName := fmt.Sprintf("%s-broker", brokerName)
					stdout, _, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
						"./bin/pulsar-admin topics permissions "+ptopic.Spec.Name)
					g.Expect(err).Should(Succeed())
					g.Expect(stdout).Should(Not(BeEmpty()))
					g.Expect(stdout).Should(ContainSubstring("ironman"))
					g.Expect(stdout).Should(Not(ContainSubstring("spiderman")))
				}, "20s", "100ms").Should(Succeed())
			})

			It("should have updated state annotation with only ironman role", func() {
				Eventually(func() error {
					expectedState := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"ironman"},
						Actions:      []string{"produce", "consume", "functions"},
					}

					validatePermissionStateAnnotation(ctx, ppermissionName, namespaceName, expectedState)
					return nil
				}, "20s", "100ms").Should(Succeed())
			})
		})

		Context("Multiple PulsarPermissions on same Topic", Label("Permissions"), func() {
			It("should create multiple pulsarpermissions successfully", func() {
				err := k8sClient.Create(ctx, ppermission2)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
				err = k8sClient.Create(ctx, ppermission3)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("both permissions should be ready", func() {
				Eventually(func() bool {
					perm2 := &v1alphav1.PulsarPermission{}
					tns2 := types.NamespacedName{Namespace: namespaceName, Name: ppermission2Name}
					Expect(k8sClient.Get(ctx, tns2, perm2)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(perm2)
				}, "20s", "100ms").Should(BeTrue())

				Eventually(func() bool {
					perm3 := &v1alphav1.PulsarPermission{}
					tns3 := types.NamespacedName{Namespace: namespaceName, Name: ppermission3Name}
					Expect(k8sClient.Get(ctx, tns3, perm3)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(perm3)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have correct initial state annotations", func() {
				Eventually(func() error {
					expectedState2 := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"batman"},
						Actions:      []string{"produce", "consume"},
					}
					validatePermissionStateAnnotation(ctx, ppermission2Name, namespaceName, expectedState2)
					return nil
				}, "20s", "100ms").Should(Succeed())

				Eventually(func() error {
					expectedState3 := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"superman"},
						Actions:      []string{"functions"},
					}
					validatePermissionStateAnnotation(ctx, ppermission3Name, namespaceName, expectedState3)
					return nil
				}, "20s", "100ms").Should(Succeed())
			})

			It("topic should contain all roles from both permissions", func() {
				validateTopicPermissionsContain(topicName, []string{"batman", "superman"})
			})

			It("should modify first permission by adding wonderwoman", func() {
				perm2 := &v1alphav1.PulsarPermission{}
				tns2 := types.NamespacedName{Namespace: namespaceName, Name: ppermission2Name}
				Expect(k8sClient.Get(ctx, tns2, perm2)).Should(Succeed())
				perm2.Spec.Roles = append(perm2.Spec.Roles, "wonderwoman")
				err := k8sClient.Update(ctx, perm2)
				Expect(err).Should(Succeed())
			})

			It("first permission should be ready with updated roles", func() {
				Eventually(func() bool {
					perm2 := &v1alphav1.PulsarPermission{}
					tns2 := types.NamespacedName{Namespace: namespaceName, Name: ppermission2Name}
					Expect(k8sClient.Get(ctx, tns2, perm2)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(perm2)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("first permission should have updated state annotation", func() {
				Eventually(func() error {
					expectedState2 := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"batman", "wonderwoman"},
						Actions:      []string{"produce", "consume"},
					}
					validatePermissionStateAnnotation(ctx, ppermission2Name, namespaceName, expectedState2)
					return nil
				}, "20s", "100ms").Should(Succeed())
			})

			It("second permission should remain unaffected", func() {
				Eventually(func() error {
					expectedState3 := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"superman"},
						Actions:      []string{"functions"},
					}
					validatePermissionStateAnnotation(ctx, ppermission3Name, namespaceName, expectedState3)
					return nil
				}, "20s", "100ms").Should(Succeed())
			})

			It("topic should contain all roles after first permission modification", func() {
				validateTopicPermissionsContain(topicName, []string{"batman", "wonderwoman", "superman"})
			})

			It("should modify second permission actions", func() {
				perm3 := &v1alphav1.PulsarPermission{}
				tns3 := types.NamespacedName{Namespace: namespaceName, Name: ppermission3Name}
				Expect(k8sClient.Get(ctx, tns3, perm3)).Should(Succeed())
				perm3.Spec.Actions = []string{"produce"}
				err := k8sClient.Update(ctx, perm3)
				Expect(err).Should(Succeed())
			})

			It("second permission should be ready with updated actions", func() {
				Eventually(func() bool {
					perm3 := &v1alphav1.PulsarPermission{}
					tns3 := types.NamespacedName{Namespace: namespaceName, Name: ppermission3Name}
					Expect(k8sClient.Get(ctx, tns3, perm3)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(perm3)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("second permission should have updated state annotation", func() {
				Eventually(func() error {
					expectedState3 := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"superman"},
						Actions:      []string{"produce"},
					}
					validatePermissionStateAnnotation(ctx, ppermission3Name, namespaceName, expectedState3)
					return nil
				}, "20s", "100ms").Should(Succeed())
			})

			It("first permission should remain unaffected by second permission changes", func() {
				Eventually(func() error {
					expectedState2 := connection.PulsarPermissionState{
						ResourceType: "topic",
						ResourceName: topicName,
						Roles:        []string{"batman", "wonderwoman"},
						Actions:      []string{"produce", "consume"},
					}
					validatePermissionStateAnnotation(ctx, ppermission2Name, namespaceName, expectedState2)
					return nil
				}, "20s", "100ms").Should(Succeed())
			})

			It("topic should still contain all roles after all modifications", func() {
				validateTopicPermissionsContain(topicName, []string{"batman", "wonderwoman", "superman"})
			})
		})

		Context("PulsarFunction & PulsarPackage operation", func() {
			It("should create the pulsarpackage successfully", func() {
				err := k8sClient.Create(ctx, ppackage)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("the package should be ready", func() {
				Eventually(func() bool {
					p := &v1alphav1.PulsarPackage{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncName}
					Expect(k8sClient.Get(ctx, tns, p)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(p)
				}, "40s", "100ms").Should(BeTrue())
			})

			It("should create the pulsarfunction successfully", func() {
				err := k8sClient.Create(ctx, pfunc)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("the function should be ready", func() {
				Eventually(func() bool {
					f := &v1alphav1.PulsarFunction{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncName}
					Expect(k8sClient.Get(ctx, tns, f)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(f)
				}, "40s", "100ms").Should(BeTrue())
			})

			It("cleanup the pulsarfunction successfully", func() {
				Eventually(func(g Gomega) {
					t := &v1alphav1.PulsarFunction{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncName}
					g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
				}).Should(Succeed())
			})

			It("cleanup the pulsarpackage successfully", func() {
				Eventually(func(g Gomega) {
					t := &v1alphav1.PulsarPackage{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncName}
					g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		Context("PulsarSink operation", func() {
			It("should create the pulsarsink successfully", func() {
				err := k8sClient.Create(ctx, psink)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("the sink should be ready", func() {
				Eventually(func() bool {
					s := &v1alphav1.PulsarSink{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: psinkName}
					Expect(k8sClient.Get(ctx, tns, s)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(s)
				}, "40s", "100ms").Should(BeTrue())
			})

			It("cleanup the pulsarsink successfully", func() {
				Eventually(func(g Gomega) {
					t := &v1alphav1.PulsarSink{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: psinkName}
					g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		Context("PulsarSource operation", func() {
			It("should create the pulsarsource successfully", func() {
				err := k8sClient.Create(ctx, psource)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("the source should be ready", func() {
				Eventually(func() bool {
					s := &v1alphav1.PulsarSource{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: psourceName}
					Expect(k8sClient.Get(ctx, tns, s)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(s)
				}, "40s", "100ms").Should(BeTrue())
			})

			It("cleanup the pulsarsource successfully", func() {
				Eventually(func(g Gomega) {
					t := &v1alphav1.PulsarSource{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: psourceName}
					g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		Context("PulsarFunction & PulsarPackage operation with failure", func() {
			It("should create the pulsarpackage successfully", func() {
				err := k8sClient.Create(ctx, ppackage)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("the package should be ready", func() {
				Eventually(func() bool {
					p := &v1alphav1.PulsarPackage{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncName}
					Expect(k8sClient.Get(ctx, tns, p)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(p)
				}, "40s", "100ms").Should(BeTrue())
			})

			It("should create the pulsarfunction successfully with error config", func() {
				err := k8sClient.Create(ctx, pfuncfailure)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("the function should be not ready", func() {
				Eventually(func() bool {
					f := &v1alphav1.PulsarFunction{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncFailureName}
					Expect(k8sClient.Get(ctx, tns, f)).Should(Succeed())
					return !v1alphav1.IsPulsarResourceReady(f)
				}, "40s", "100ms").Should(BeTrue())
			})

			It("should update the pulsarfunction successfully with correct config", func() {
				f := &v1alphav1.PulsarFunction{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncFailureName}
				Expect(k8sClient.Get(ctx, tns, f)).Should(Succeed())
				f.Spec.Jar.URL = ppackageurl
				err := k8sClient.Update(ctx, f)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("the function should be ready after update", func() {
				Eventually(func() bool {
					f := &v1alphav1.PulsarFunction{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncFailureName}
					Expect(k8sClient.Get(ctx, tns, f)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(f)
				}, "40s", "100ms").Should(BeTrue())
			})

			It("cleanup the pulsarfunction successfully", func() {
				Eventually(func(g Gomega) {
					t := &v1alphav1.PulsarFunction{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncFailureName}
					g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
				}).Should(Succeed())
			})

			It("cleanup the pulsarpackage successfully", func() {
				Eventually(func(g Gomega) {
					t := &v1alphav1.PulsarPackage{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pfuncName}
					g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		Context("PulsarNSIsolationPolicy operation", func() {
			It("should create the pulsar ns-isolation-policy successfully", func() {
				err := k8sClient.Create(ctx, pnsisolationpolicy)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("the ns-isolation-policy should be ready", func() {
				Eventually(func() bool {
					s := &v1alphav1.PulsarNSIsolationPolicy{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pnsIsolationPolicyName}
					Expect(k8sClient.Get(ctx, tns, s)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(s)
				}, "40s", "100ms").Should(BeTrue())
			})

			It("cleanup the pulsar ns-isolation-policy successfully", func() {
				Eventually(func(g Gomega) {
					t := &v1alphav1.PulsarNSIsolationPolicy{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pnsIsolationPolicyName}
					g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
				}).Should(Succeed())
			})
		})

		Context("PulsarPackage operation with valid file URL", func() {
			It("should create the pulsarpackage successfully with valid file URL", func() {
				ppackage.Spec.FileURL = "file:///manager" // we use the manager binary as the file URL
				ppackage.Spec.PackageURL = "function://public/default/file@valid"
				ppackage.Name = "test-func-valid"
				err := k8sClient.Create(ctx, ppackage)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})
		})

		AfterAll(func() {
			Eventually(func(g Gomega) {
				t := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
				g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				t := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName2}
				g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				t := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: partitionedTopic.Name}
				g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				ns := &v1alphav1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: pnamespaceName}
				g.Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, ns)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				t := &v1alphav1.PulsarTenant{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ptenantName}
				g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				conn := &v1alphav1.PulsarConnection{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: pconnName}
				g.Expect(k8sClient.Get(ctx, tns, conn)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, conn)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				perm := &v1alphav1.PulsarPermission{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ppermissionName}
				g.Expect(k8sClient.Get(ctx, tns, perm)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, perm)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				perm2 := &v1alphav1.PulsarPermission{}
				tns2 := types.NamespacedName{Namespace: namespaceName, Name: ppermission2Name}
				g.Expect(k8sClient.Get(ctx, tns2, perm2)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, perm2)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				perm3 := &v1alphav1.PulsarPermission{}
				tns3 := types.NamespacedName{Namespace: namespaceName, Name: ppermission3Name}
				g.Expect(k8sClient.Get(ctx, tns3, perm3)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, perm3)).Should(Succeed())
			}).Should(Succeed())
		})
	})

})

func updateTopicSchema(ctx context.Context, topicName, exampleSchemaDef string) {
	topic := &v1alphav1.PulsarTopic{}
	tns := types.NamespacedName{Namespace: namespaceName, Name: topicName}
	Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())
	topic.Spec.SchemaInfo = &v1alphav1.SchemaInfo{
		Type:   "JSON",
		Schema: exampleSchemaDef,
	}
	Expect(k8sClient.Update(ctx, topic)).Should(Succeed())
}
