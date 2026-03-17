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

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("APIServerConnectionReconciler", func() {
	const (
		testNamespace  = "test-snc"
		connectionName = "test-connection"
	)

	var (
		reconciler *APIServerConnectionReconciler
		ctx        context.Context
		ns         *corev1.Namespace
	)

	BeforeEach(func() {
		ctx = context.Background()
		reconciler = &APIServerConnectionReconciler{
			Client: k8sClient,
		}

		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		err := k8sClient.Create(ctx, ns)
		if err != nil {
			// namespace may already exist from previous test
			_ = err
		}
	})

	Describe("listDependentResources", func() {
		AfterEach(func() {
			// Clean up all resources in the namespace
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.ComputeFlinkDeployment{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.ComputeWorkspace{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.Secret{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.ServiceAccountBinding{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.ServiceAccount{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.APIKey{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.RoleBinding{}, client.InNamespace(testNamespace))
		})

		It("should return empty when no dependent resources exist", func() {
			remaining, err := reconciler.listDependentResources(ctx, testNamespace, connectionName)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(BeEmpty())
		})

		It("should return dependent resources that reference the connection", func() {
			// Create a workspace referencing the connection
			ws := &resourcev1alpha1.ComputeWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ws",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ComputeWorkspaceSpec{
					APIServerRef: corev1.LocalObjectReference{Name: connectionName},
				},
			}
			Expect(k8sClient.Create(ctx, ws)).To(Succeed())

			// Create a secret referencing the connection
			secret := &resourcev1alpha1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.SecretSpec{
					APIServerRef: corev1.LocalObjectReference{Name: connectionName},
					Data:         map[string]string{"key": "value"},
				},
			}
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			remaining, err := reconciler.listDependentResources(ctx, testNamespace, connectionName)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(HaveLen(2))
			Expect(remaining).To(ContainElement("ComputeWorkspace/test-ws"))
			Expect(remaining).To(ContainElement("Secret/test-secret"))
		})

		It("should not return resources referencing a different connection", func() {
			ws := &resourcev1alpha1.ComputeWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ws-other",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ComputeWorkspaceSpec{
					APIServerRef: corev1.LocalObjectReference{Name: "other-connection"},
				},
			}
			Expect(k8sClient.Create(ctx, ws)).To(Succeed())

			remaining, err := reconciler.listDependentResources(ctx, testNamespace, connectionName)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(BeEmpty())
		})

		It("should return multiple resource types", func() {
			ws := &resourcev1alpha1.ComputeWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ws-multi",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ComputeWorkspaceSpec{
					APIServerRef: corev1.LocalObjectReference{Name: connectionName},
				},
			}
			Expect(k8sClient.Create(ctx, ws)).To(Succeed())

			sa := &resourcev1alpha1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sa",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ServiceAccountSpec{
					APIServerRef: corev1.LocalObjectReference{Name: connectionName},
				},
			}
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			apiKey := &resourcev1alpha1.APIKey{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-apikey",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.APIKeySpec{
					APIServerRef:       corev1.LocalObjectReference{Name: connectionName},
					ServiceAccountName: "test-sa",
				},
			}
			Expect(k8sClient.Create(ctx, apiKey)).To(Succeed())

			remaining, err := reconciler.listDependentResources(ctx, testNamespace, connectionName)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(HaveLen(3))
			Expect(remaining).To(ContainElement("ComputeWorkspace/test-ws-multi"))
			Expect(remaining).To(ContainElement("ServiceAccount/test-sa"))
			Expect(remaining).To(ContainElement("APIKey/test-apikey"))
		})

		It("should return FlinkDeployments that inherit the connection from a workspace", func() {
			ws := &resourcev1alpha1.ComputeWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ws-inherited",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ComputeWorkspaceSpec{
					APIServerRef: corev1.LocalObjectReference{Name: connectionName},
				},
			}
			Expect(k8sClient.Create(ctx, ws)).To(Succeed())

			deployment := &resourcev1alpha1.ComputeFlinkDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-flink",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ComputeFlinkDeploymentSpec{
					WorkspaceName: "test-ws-inherited",
				},
			}
			Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

			remaining, err := reconciler.listDependentResources(ctx, testNamespace, connectionName)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(ContainElement("ComputeWorkspace/test-ws-inherited"))
			Expect(remaining).To(ContainElement("ComputeFlinkDeployment/test-flink"))
		})

		It("should return ServiceAccountBindings that inherit the connection from a service account", func() {
			sa := &resourcev1alpha1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sa-inherited",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ServiceAccountSpec{
					APIServerRef: corev1.LocalObjectReference{Name: connectionName},
				},
			}
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			binding := &resourcev1alpha1.ServiceAccountBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-binding",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ServiceAccountBindingSpec{
					ServiceAccountName: "test-sa-inherited",
					PoolMemberRefs: []resourcev1alpha1.PoolMemberReference{
						{
							Namespace: testNamespace,
							Name:      "pool-member",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, binding)).To(Succeed())

			remaining, err := reconciler.listDependentResources(ctx, testNamespace, connectionName)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(ContainElement("ServiceAccount/test-sa-inherited"))
			Expect(remaining).To(ContainElement("ServiceAccountBinding/test-binding"))
		})
	})

	Describe("connectionNameForDependentResource", func() {
		AfterEach(func() {
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.ComputeFlinkDeployment{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.ComputeWorkspace{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.ServiceAccountBinding{}, client.InNamespace(testNamespace))
			_ = k8sClient.DeleteAllOf(ctx, &resourcev1alpha1.ServiceAccount{}, client.InNamespace(testNamespace))
		})

		It("should resolve a FlinkDeployment connection through its workspace", func() {
			ws := &resourcev1alpha1.ComputeWorkspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ws-map",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ComputeWorkspaceSpec{
					APIServerRef: corev1.LocalObjectReference{Name: connectionName},
				},
			}
			Expect(k8sClient.Create(ctx, ws)).To(Succeed())

			deployment := &resourcev1alpha1.ComputeFlinkDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-flink-map",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ComputeFlinkDeploymentSpec{
					WorkspaceName: ws.Name,
				},
			}
			Expect(reconciler.connectionNameForDependentResource(ctx, deployment)).To(Equal(connectionName))
		})

		It("should resolve a ServiceAccountBinding connection through its service account", func() {
			sa := &resourcev1alpha1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sa-map",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ServiceAccountSpec{
					APIServerRef: corev1.LocalObjectReference{Name: connectionName},
				},
			}
			Expect(k8sClient.Create(ctx, sa)).To(Succeed())

			binding := &resourcev1alpha1.ServiceAccountBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-binding-map",
					Namespace: testNamespace,
				},
				Spec: resourcev1alpha1.ServiceAccountBindingSpec{
					ServiceAccountName: sa.Name,
					PoolMemberRefs: []resourcev1alpha1.PoolMemberReference{
						{
							Namespace: testNamespace,
							Name:      "pool-member",
						},
					},
				},
			}
			Expect(reconciler.connectionNameForDependentResource(ctx, binding)).To(Equal(connectionName))
		})
	})
})
