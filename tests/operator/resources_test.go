// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package operator_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	pulsarv1alpha1 "github.com/streamnative/resources-operator/api/v1alpha1"
	"github.com/streamnative/resources-operator/tests/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Resources", func() {
	var (
		ctx                 context.Context
		pconn               *pulsarv1alpha1.PulsarConnection
		pconnName           string = "test-connection"
		ptenant             *pulsarv1alpha1.PulsarTenant
		ptenantName         string = "test-tenant"
		tenantName          string = "cloud"
		lifecyclePolicy     pulsarv1alpha1.PulsarResourceLifeCyclePolicy
		pnamespace          *pulsarv1alpha1.PulsarNamespace
		pnamespaceName      string = "test-namespace"
		pulsarNamespaceName string = "cloud/stage"
		ptopic              *pulsarv1alpha1.PulsarTopic
		ptopicName          string = "test-topic"
		topicName           string = "persistent://cloud/stage/user"
		nodePort            int32  = 30080
	)

	BeforeEach(func() {
		ctx = context.TODO()
		// Use this url on gke
		adminServiceURL := fmt.Sprintf("http://%s-broker.%s.svc.cluster.local:8080", snHelmReleaseName, namespaceName)
		// Use this url on kind cluster
		// adminServiceURL := fmt.Sprintf("http://%s:%d", nodeIP, nodePort)
		tokenSecretName := fmt.Sprintf("%s-vault-secret-env-injection", snHelmReleaseName)
		pconn = utils.MakePulsarConnection(namespaceName, pconnName, adminServiceURL, tokenSecretName)
		adminRoles := []string{"superman"}
		allowedClusters := []string{}
		lifecyclePolicy = pulsarv1alpha1.CleanUpAfterDeletion
		ptenant = utils.MakePulsarTenant(namespaceName, ptenantName, tenantName, pconnName, adminRoles, allowedClusters, lifecyclePolicy)
		pnamespace = utils.MakePulsarNamespace(namespaceName, pnamespaceName, pulsarNamespaceName, pconnName, lifecyclePolicy)
		ptopic = utils.MakePulsarTopic(namespaceName, ptopicName, topicName, pconnName, lifecyclePolicy)
	})

	Describe("Basic resource operations", Ordered, func() {
		Context("Check SN Platform", func() {
			It("should create the sn-platform successfully", func() {
				Eventually(func() bool {
					statefulset := &v1.StatefulSet{}
					k8sClient.Get(ctx, types.NamespacedName{
						Name:      fmt.Sprintf("%s-proxy", snHelmReleaseName),
						Namespace: namespaceName,
					}, statefulset)
					logrus.Info(statefulset)
					return statefulset.Status.ReadyReplicas > 0 && statefulset.Status.ReadyReplicas == statefulset.Status.Replicas
				}, 20*time.Minute, 30*time.Second).Should(BeTrue())
				logrus.Infof("sn chart %s installed", snHelmReleaseName)
			})

			It("should update the type of service proxy to NodePort", func() {
				svc := &corev1.Service{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-proxy", snHelmReleaseName),
					Namespace: namespaceName,
				}, svc)
				Expect(err).Should(Succeed())
				svc.Spec.Type = corev1.ServiceTypeNodePort
				svc.Spec.Ports[0].NodePort = nodePort
				Expect(k8sClient.Update(ctx, svc)).Should(Succeed())
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
					t := &pulsarv1alpha1.PulsarTenant{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ptenantName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return pulsarv1alpha1.IsPulsarResourceReady(t)
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
					ns := &pulsarv1alpha1.PulsarNamespace{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: pnamespaceName}
					Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
					return pulsarv1alpha1.IsPulsarResourceReady(ns)
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
					t := &pulsarv1alpha1.PulsarTopic{}
					tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
					Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
					return pulsarv1alpha1.IsPulsarResourceReady(t)
				}, "20s", "100ms").Should(BeTrue())
			})
		})

		// Context("PulsarPermission operation", func() {
		// 	It("should grant the pulsarpermission successfully", func() {
		// 		err := k8sClient.Create(ctx, ptopic)
		// 		Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
		// 	})

		// 	It("should be ready", func() {
		// 		Eventually(func() bool {
		// 			t := &pulsarv1alpha1.PulsarNamespace{}
		// 			tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
		// 			Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
		// 			return pulsarv1alpha1.IsPulsarResourceReady(t)
		// 		}, "20s", "100ms").Should(BeTrue())
		// 	})
		// })
		AfterAll(func() {
			Eventually(func(g Gomega) {
				t := &pulsarv1alpha1.PulsarTopic{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ptopicName}
				g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				ns := &pulsarv1alpha1.PulsarNamespace{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: pnamespaceName}
				g.Expect(k8sClient.Get(ctx, tns, ns)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, ns)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				t := &pulsarv1alpha1.PulsarTenant{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: ptenantName}
				g.Expect(k8sClient.Get(ctx, tns, t)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, t)).Should(Succeed())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				conn := &pulsarv1alpha1.PulsarConnection{}
				tns := types.NamespacedName{Namespace: namespaceName, Name: pconnName}
				g.Expect(k8sClient.Get(ctx, tns, conn)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, conn)).Should(Succeed())
			}).Should(Succeed())

			// Eventually(func(g Gomega) {
			// 	perm := &pulsarv1alpha1.PulsarPermission{}
			// 	tns := types.NamespacedName{Namespace: namespaceName, Name: ppermissionName}
			// 	g.Expect(k8sClient.Get(ctx, tns, perm)).Should(Succeed())
			// 	g.Expect(k8sClient.Delete(ctx, perm)).Should(Succeed())
			// }).Should(Succeed())
		})
	})

})
