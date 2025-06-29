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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	v1alphav1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
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
