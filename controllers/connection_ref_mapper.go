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
	namespace := object.GetNamespace()
	if ref.Namespace != "" {
		namespace = ref.Namespace
	}
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      ref.Name,
			},
		},
	}
}

// var _ handler.Mapper = &PulsarConnectionRefMapper{}

// ConnectionRefMapper maps resource object to PulsarConnection request
func ConnectionRefMapper(ctx context.Context, object client.Object) []reconcile.Request {
	ref := getConnectionRef(object)
	if ref == nil {
		return nil
	}
	namespace := object.GetNamespace()
	if ref.Namespace != "" {
		namespace = ref.Namespace
	}
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      ref.Name,
			},
		},
	}
}

func getConnectionRef(object client.Object) *pulsarv1alpha1.PulsarConnectionRef {
	switch v := object.(type) {
	case *pulsarv1alpha1.PulsarTenant:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarNamespace:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarTopic:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarPermission:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarGeoReplication:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarFunction:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarSource:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarSink:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarPackage:
		return &v.Spec.ConnectionRef
	case *pulsarv1alpha1.PulsarNSIsolationPolicy:
		return &v.Spec.ConnectionRef
	default:
		return nil
	}
}
