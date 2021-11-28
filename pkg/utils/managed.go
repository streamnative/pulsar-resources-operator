// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package utils

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ManagedAnnotation = "cloud.streamnative.io/managed"
)

func IsManaged(object metav1.Object) bool {
	managed, exists := object.GetAnnotations()[ManagedAnnotation]
	return !exists || managed != "false"
}
