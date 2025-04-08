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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Context("TestEventSource", func() {
	It("should trigger reconcile after delay", func() {
		source := NewEventSource(ctrl.Log.WithName("test"))
		key := "test"
		Expect(source.Events).ShouldNot(BeClosed())
		source.CreateIfAbsent(3*time.Second, &corev1.Pod{}, key)
		Eventually(source.Events, "10s").Should(
			Receive(Equal(event.GenericEvent{Object: &corev1.Pod{}})),
		)
		// key should be removed from the event map
		Expect(source.Contains(key)).Should(BeFalse())
		source.Close()
		Expect(source.Events).Should(BeClosed())
	})

	It("should be idempotent for the same key", func() {
		source := NewEventSource(ctrl.Log.WithName("test"))
		key := "test"
		Expect(source.Events).ShouldNot(BeClosed())

		source.CreateIfAbsent(3*time.Second, &corev1.Pod{}, key)
		source.CreateIfAbsent(10*time.Second, &corev1.Pod{}, key)

		Eventually(source.Events, "10s").
			Should(Receive(Equal(event.GenericEvent{Object: &corev1.Pod{}})))
		// key should be removed from the event map
		Expect(source.Contains(key)).Should(BeFalse())
		source.Close()
		Expect(source.Events).Should(BeClosed())
	})

	It("should be removed", func() {
		source := NewEventSource(ctrl.Log.WithName("test"))
		key := "test"
		Expect(source.Events).ShouldNot(BeClosed())

		source.CreateIfAbsent(2*time.Second, &corev1.Pod{}, key)
		source.Remove(key)
		// key should be removed from the event map
		Expect(source.Contains(key)).Should(BeFalse())
		Consistently(source.Events, "3s").
			ShouldNot(Receive(Equal(event.GenericEvent{Object: &corev1.Pod{}})))
		source.Close()
		Expect(source.Events).Should(BeClosed())
	})
})
