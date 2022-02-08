// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

package utils

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// ManagedAnnotation indicates the object is managed by the controller
	ManagedAnnotation = "cloud.streamnative.io/managed"
)

// IsManaged returns true if the object is under control of the controller by checking
// the specific annotation
func IsManaged(object metav1.Object) bool {
	managed, exists := object.GetAnnotations()[ManagedAnnotation]
	return !exists || managed != "false"
}
