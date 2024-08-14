package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Context("Test retryer", Ordered, func() {
	var retryer *ReconcileRetryer
	var obj *corev1.Pod

	BeforeEach(func() {
		retryer = NewReconcileRetryer(2, NewEventSource(ctrl.Log.WithName("test")))
		obj = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "test-pod",
				Namespace:       "test-ns",
				UID:             "test-uid",
				ResourceVersion: "test-rv",
			},
		}
	})

	It("should trigger retry", func() {
		Expect(retryer.Source()).ShouldNot(BeClosed())
		retryer.CreateIfAbsent(obj)

		By("should trigger reconcile after delay 2 times")
		Eventually(retryer.Source(), "20s").Should(
			Receive(Equal(event.GenericEvent{Object: obj})),
		)
		retryer.CreateIfAbsent(obj)

		Eventually(retryer.Source(), "20s").Should(
			Receive(Equal(event.GenericEvent{Object: obj})),
		)

		By("should not trigger reconcile after 2 times")
		Consistently(retryer.Source(), "20s").ShouldNot(
			Receive(Equal(event.GenericEvent{Object: obj})),
		)
		Expect(retryer.Contains(obj)).Should(BeTrue())

		By("should trigger reconcile when resource version changed")
		obj.SetResourceVersion("test-rv2")
		retryer.CreateIfAbsent(obj)
		Eventually(retryer.Source(), "20s").Should(
			Receive(Equal(event.GenericEvent{Object: obj})),
		)
		Eventually(retryer.Source(), "20s").Should(
			Receive(Equal(event.GenericEvent{Object: obj})),
		)

		Consistently(retryer.Source(), "20s").ShouldNot(
			Receive(Equal(event.GenericEvent{Object: obj})),
		)

		retryer.Close()
		Expect(retryer.Source()).Should(BeClosed())
	})

	It("should remove retry", func() {
		obj2 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod2",
				Namespace: "test-ns",
				UID:       "test-uid2",
			},
		}

		Expect(retryer.Source()).ShouldNot(BeClosed())
		retryer.CreateIfAbsent(obj)
		retryer.CreateIfAbsent(obj2)

		Expect(retryer.Contains(obj)).Should(BeTrue())
		Expect(retryer.Contains(obj2)).Should(BeTrue())
		retryer.Remove(obj)
		Expect(retryer.Contains(obj)).Should(BeFalse())
		retryer.Remove(obj2)
		Expect(retryer.Contains(obj2)).Should(BeFalse())

		By("should not trigger reconcile after delay")
		Consistently(retryer.Source(), "10s").ShouldNot(
			Receive(Equal(event.GenericEvent{Object: obj})),
		)

		retryer.Close()
		Expect(retryer.Source()).Should(BeClosed())
	})
})
