// Copyright 2026 StreamNative
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
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

var _ = Describe("PulsarTopic Boolean Field Persistence", func() {
	Context("when PulsarTopic status has GeoReplicationEnabled set to false", func() {
		It("should preserve false value in JSON serialization", func() {
			ctx := context.Background()
			namespace := "default"
			topicName := "test-geo-replication-false"
			connectionName := "test-connection"

			connection := &resourcev1alpha1.PulsarConnection{
				ObjectMeta: metav1.ObjectMeta{
					Name:      connectionName,
					Namespace: namespace,
				},
				Spec: resourcev1alpha1.PulsarConnectionSpec{
					AdminServiceURL:  "http://localhost:8080",
					BrokerServiceURL: "pulsar://localhost:6650",
				},
			}
			Expect(k8sClient.Create(ctx, connection)).Should(Succeed())

			topic := &resourcev1alpha1.PulsarTopic{
				ObjectMeta: metav1.ObjectMeta{
					Name:      topicName,
					Namespace: namespace,
				},
				Spec: resourcev1alpha1.PulsarTopicSpec{
					Name: "persistent://public/default/" + topicName,
					ConnectionRef: corev1.LocalObjectReference{
						Name: connectionName,
					},
				},
				Status: resourcev1alpha1.PulsarTopicStatus{
					GeoReplicationEnabled: false,
				},
			}
			Expect(k8sClient.Create(ctx, topic)).Should(Succeed())

			// Get the created topic and verify status
			createdTopic := &resourcev1alpha1.PulsarTopic{}
			topicKey := types.NamespacedName{Name: topicName, Namespace: namespace}
			Expect(k8sClient.Get(ctx, topicKey, createdTopic)).Should(Succeed())
			Expect(createdTopic.Status.GeoReplicationEnabled).Should(Equal(false))

			// Verify the serialization
			statusJSON, err := json.Marshal(createdTopic.Status)
			Expect(err).Should(Succeed())
			var statusMap map[string]interface{}
			Expect(json.Unmarshal(statusJSON, &statusMap)).Should(Succeed())
			Expect(statusMap).Should(HaveKey("geoReplicationEnabled"))
			geoRepValue, exists := statusMap["geoReplicationEnabled"]
			Expect(exists).Should(BeTrue())
			Expect(geoRepValue).Should(Equal(false))
		})
	})
})
