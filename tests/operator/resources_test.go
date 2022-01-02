// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package operator_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	v1alphav1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/tests/utils"
)

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
		nodePort            int32  = 30080
		ppermission         *v1alphav1.PulsarPermission
		ppermissionName     string = "test-permission"
	)

	BeforeEach(func() {
		ctx = context.TODO()
		// use ClusterIP svc when run operator in k8s
		// adminServiceURL := fmt.Sprintf("http://%s-broker.%s.svc.cluster.local:8080", brokerName, namespaceName)
		// use NodePort svc when cluster is kind cluster and run operator locally, the nodePort need to be setup in kind
		adminServiceURL := fmt.Sprintf("http://127.0.0.1:%d", nodePort)
		pconn = utils.MakePulsarConnection(namespaceName, pconnName, adminServiceURL)
		adminRoles := []string{"superman"}
		allowedClusters := []string{}
		lifecyclePolicy = v1alphav1.CleanUpAfterDeletion
		ptenant = utils.MakePulsarTenant(namespaceName, ptenantName, tenantName, pconnName, adminRoles, allowedClusters, lifecyclePolicy)
		pnamespace = utils.MakePulsarNamespace(namespaceName, pnamespaceName, pulsarNamespaceName, pconnName, lifecyclePolicy)
		ptopic = utils.MakePulsarTopic(namespaceName, ptopicName, topicName, pconnName, lifecyclePolicy)
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

			It("should create NodePort service successfully", func() {
				svc := &corev1.Service{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-proxy-external", proxyName),
					Namespace: namespaceName,
				}, svc)

				Expect(err).Should(Succeed())
				Expect(svc.Spec.Type == corev1.ServiceTypeNodePort).Should(BeTrue())
				Expect(svc.Spec.Ports[1].NodePort == nodePort).Should(BeTrue())
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
			})

			It("should be ready", func() {
				Eventually(func() bool {
					t := &v1alphav1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return v1alphav1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
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
