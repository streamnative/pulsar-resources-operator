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
		partitionedTopic = &v1alphav1.PulsarTopic{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespaceName,
				Name:      "test-partitioned-topic",
			},
			Spec: v1alphav1.PulsarTopicSpec{
				Name:       "persistent://cloud/stage/partitioned-topic",
				Partitions: pointer.Int32Ptr(1),
				ConnectionRef: corev1.LocalObjectReference{
					Name: pconnName,
				},
			},
		}
		ppackage          *v1alphav1.PulsarPackage
		ppackageurl       string = "function://public/default/api-examples@v3.2.3.3"
		pfuncName         string = "test-func"
		psinkName         string = "test-sink"
		psourceName       string = "test-source"
		pfunc             *v1alphav1.PulsarFunction
		psinkpackageurl   string = "builtin://data-generator"
		psink             *v1alphav1.PulsarSink
		psource           *v1alphav1.PulsarSource
		psourcepackageurl string = "builtin://data-generator"
	)

	BeforeEach(func() {
		Expect(feature.SetFeatureGates()).ShouldNot(HaveOccurred())
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
		ppackage = utils.MakePulsarPackage(namespaceName, pfuncName, ppackageurl, pconnName, lifecyclePolicy)
		pfunc = utils.MakePulsarFunction(namespaceName, pfuncName, ppackageurl, pconnName, lifecyclePolicy)
		psink = utils.MakePulsarSink(namespaceName, psinkName, psinkpackageurl, pconnName, lifecyclePolicy)
		psource = utils.MakePulsarSource(namespaceName, psourceName, psourcepackageurl, pconnName, lifecyclePolicy)
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

		Context("PulsarTopic operation", Ordered, func() {
			It("should create the pulsartopic successfully", func() {
				err := k8sClient.Create(ctx, ptopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
				err = k8sClient.Create(ctx, ptopic2)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
				err = k8sClient.Create(ctx, partitionedTopic)
				Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
			})

			It("should be ready", func() {
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

				By("check topic1 schema is deleted in pulsar")
				Eventually(func(g Gomega) {
					_, stderr, err := utils.ExecInPod(k8sConfig, namespaceName, podName, containerName,
						"./bin/pulsarctl -s http://localhost:8080 --token=$PROXY_TOKEN  schemas get "+ptopic.Spec.Name)
					g.Expect(err).ShouldNot(BeNil())
					g.Expect(stderr).Should(Not(BeEmpty()))
					format.MaxLength = 0
					g.Expect(stderr).Should(ContainSubstring("404"))
				}, "5s", "100ms").Should(Succeed())
			})

			It("should increase the partitions successfully", func() {
				curTopic := &v1alphav1.PulsarTopic{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Namespace: partitionedTopic.Namespace,
					Name:      partitionedTopic.Name,
				}, curTopic)).ShouldNot(HaveOccurred())

				curTopic.Spec.Partitions = pointer.Int32Ptr(2)
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
