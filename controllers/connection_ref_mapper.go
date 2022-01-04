// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package controllers

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pulsarv1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// PulsarConnectionRefMapper maps resource requests to PulsarConnection
type PulsarConnectionRefMapper struct {
}

// Map maps resource object to PulsarConnection request
func (p *PulsarConnectionRefMapper) Map(object client.Object) []reconcile.Request {
	ref := getConnectionRef(object)
	if ref == nil {
		return nil
	}
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: object.GetNamespace(),
				Name:      ref.Name,
			},
		},
	}
}

// var _ handler.Mapper = &PulsarConnectionRefMapper{}

// ConnectionRefMapper maps resource object to PulsarConnection request
func ConnectionRefMapper(object client.Object) []reconcile.Request {
	ref := getConnectionRef(object)
	if ref == nil {
		return nil
	}
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: object.GetNamespace(),
				Name:      ref.Name,
			},
		},
	}
}

func getConnectionRef(object client.Object) *corev1.LocalObjectReference {
	switch v := object.(type) {
	case *pulsarv1alpha1.PulsarTenant:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarNamespace:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarTopic:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarPermission:
		return &v.Spec.ConnectionRef
	default:
		return nil
	}
}
