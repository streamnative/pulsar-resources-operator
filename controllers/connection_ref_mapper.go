// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package controllers

import (
	pulsarv1alpha1 "github.com/streamnative/resources-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PulsarConnectionRefMapper struct {
}

func (p *PulsarConnectionRefMapper) Map(object client.Object) []reconcile.Request {
	var ref *corev1.LocalObjectReference = nil
	if tenant, ok := object.(*pulsarv1alpha1.PulsarTenant); ok {
		ref = &tenant.Spec.ConnectionRef
	} else if namespace, ok := object.(*pulsarv1alpha1.PulsarNamespace); ok {
		ref = &namespace.Spec.ConnectionRef
	} else if topic, ok := object.(*pulsarv1alpha1.PulsarTopic); ok {
		ref = &topic.Spec.ConnectionRef
	} else if permission, ok := object.(*pulsarv1alpha1.PulsarPermission); ok {
		ref = &permission.Spec.ConnectionRef
	}
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

func ConnectionRefMapper(object client.Object) []reconcile.Request {
	var ref *corev1.LocalObjectReference = nil
	if tenant, ok := object.(*pulsarv1alpha1.PulsarTenant); ok {
		ref = &tenant.Spec.ConnectionRef
	} else if namespace, ok := object.(*pulsarv1alpha1.PulsarNamespace); ok {
		ref = &namespace.Spec.ConnectionRef
	} else if topic, ok := object.(*pulsarv1alpha1.PulsarTopic); ok {
		ref = &topic.Spec.ConnectionRef
	} else if permission, ok := object.(*pulsarv1alpha1.PulsarPermission); ok {
		ref = &permission.Spec.ConnectionRef
	}
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
