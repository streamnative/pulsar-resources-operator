// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

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
