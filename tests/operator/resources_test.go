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

package operator_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	v1alphav1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/tests/utils"
)

type testJSON struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
		exampleSchemaDef           = "{\"type\":\"record\",\"name\":\"Example\",\"namespace\":\"test\"," +
			"\"fields\":[{\"name\":\"ID\",\"type\":\"int\"},{\"name\":\"Name\",\"type\":\"string\"}]}"
	)

	BeforeEach(func() {
		ctx = context.TODO()
		// use ClusterIP svc when run operator in k8s
		adminServiceURL := fmt.Sprintf("http://%s-broker.%s.svc.cluster.local:8080", brokerName, namespaceName)
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

	})

	Describe("Basic resource operations", Ordered, func() {
		Context("Check pulsar broker", func() {
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
		Context("Check pulsar proxy", func() {
			It("should create the pulsar proxy successfully", func() {
				Eventually(func() bool {
					statefulset := &v1.StatefulSet{}
					k8sClient.Get(ctx, types.NamespacedName{
						Name:      proxyName + "-proxy",
						Namespace: namespaceName,
					}, statefulset)
					return statefulset.Status.ReadyReplicas > 0 && statefulset.Status.ReadyReplicas == statefulset.Status.Replicas
				}, "600s", "100ms").Should(BeTrue())
			})

		})

		Context("PulsarConnection operation", func() {
			It("should create the pulsarconnection successfully", func() {
				err := k8sClient.Create(ctx, pconn)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})
		})

		Context("PulsarTenant operation", func() {
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

		Context("PulsarNamespace operation", func() {
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

		Context("PulsarTopic operation", func() {
			It("should create the pulsartopic successfully", func() {
				err := k8sClient.Create(ctx, ptopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
				err = k8sClient.Create(ctx, ptopic2)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})

			It("should have the schema set", func() {
				// set schema for two topics
				updateTopicSchema(ctx, ptopicName, exampleSchemaDef)
				updateTopicSchema(ctx, ptopicName2, exampleSchemaDef)

				Eventually(func(g Gomega) {
					podName := fmt.Sprintf("%s-proxy-0", proxyName)
					containerName := "pulsar-proxy"
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

			When("enable AlwaysUpdatePulsarResource", func() {
				BeforeEach(func() {
					err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
						string(feature.AlwaysUpdatePulsarResource): true,
					})
					Expect(err).ShouldNot(HaveOccurred())
				})

				AfterEach(func() {
					err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
						string(feature.AlwaysUpdatePulsarResource): false,
					})
					Expect(err).ShouldNot(HaveOccurred())
				})

				It("should always update pulsar resource", func() {
					podName := fmt.Sprintf("%s-proxy-0", proxyName)
					containerName := "pulsar-proxy"

					By("delete topic2 schema with pulsarctl")
					stdout, _, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
						"./bin/pulsarctl -s http://localhost:8080 --token=$PROXY_TOKEN schemas delete "+ptopic2.Spec.Name)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(stdout).Should(ContainSubstring("successfully"))

					By("delete topic1 schema in k8s")
					topic := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
					Expect(k8sClient.Get(ctx, tns, topic)).Should(Succeed())
					topic.Spec.SchemaInfo = nil
					Expect(k8sClient.Update(ctx, topic)).Should(Succeed())

					By("check topic1 schema is deleted in pulsar")
					Eventually(func(g Gomega) {
						_, stderr, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
							"./bin/pulsarctl -s http://localhost:8080 --token=$PROXY_TOKEN  schemas get "+ptopic.Spec.Name)
						g.Expect(err).ShouldNot(BeNil())
						g.Expect(stderr).Should(Not(BeEmpty()))
						format.MaxLength = 0
						g.Expect(stderr).Should(ContainSubstring("404"))
					}, "5s", "100ms").Should(Succeed())

					By("check topic2 schema is restored in pulsar")
					Eventually(func(g Gomega) {
						stdout, _, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
							"./bin/pulsarctl -s http://localhost:8080 --token=$PROXY_TOKEN  schemas get "+ptopic2.Spec.Name)
						g.Expect(err).Should(Succeed())
						g.Expect(stdout).Should(Not(BeEmpty()))
						format.MaxLength = 0
						g.Expect(stdout).Should(ContainSubstring("JSON"))
					}).Should(Succeed())
				})
			})
		})

		Context("PulsarPermission operation", func() {
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
		})

		AfterAll(func() {
			Eventually(func(g Gomega) {
				t := &v1alphav1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
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
