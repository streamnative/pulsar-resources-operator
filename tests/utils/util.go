// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.
package utils

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bkv1alpha1 "github.com/streamnative/pulsar-operators/bookkeeper-operator/api/v1alpha1"
	commonv1alpha1 "github.com/streamnative/pulsar-operators/commons/api/v1alpha1"
	pulsarv1alpha1 "github.com/streamnative/pulsar-operators/pulsar-operator/api/v1alpha1"
	zkv1alpha1 "github.com/streamnative/pulsar-operators/zookeeper-operator/api/v1alpha1"
)

// CreateBKCluster creates a bookkeeper cluster in k8s, with 3 replicas.
func CreateBKCluster(ctx context.Context, k8sClient client.Client, namespace, clusterName, zkServer, image string) error {
	bk := MakeBookKeeperCluster(namespace, clusterName, zkServer, image, 3)
	return k8sClient.Create(ctx, bk)
}

// DeleteBKCluster will delete the bookkeeper cluster in k8s
func DeleteBKCluster(ctx context.Context, k8sClient client.Client, clusterName types.NamespacedName) error {
	cur := &bkv1alpha1.BookKeeperCluster{}
	err := k8sClient.Get(ctx, clusterName, cur)
	if err != nil {
		return err
	}
	return k8sClient.Delete(ctx, cur)
}

// CreateBKCluster creates a zookeeper cluster in k8s, with 3 replicas.
func CreateZKCluster(ctx context.Context, k8sClient client.Client, namespace, clusterName, image string) error {
	zk := MakeZKCluster(namespace, clusterName, image)
	return k8sClient.Create(ctx, zk)
}

// DeleteZKCluster will delete the zookeeper cluster in k8s
func DeleteZKCluster(ctx context.Context, k8sClient client.Client, clusterName types.NamespacedName) error {
	cur := &zkv1alpha1.ZooKeeperCluster{}
	err := k8sClient.Get(ctx, clusterName, cur)
	if err != nil {
		return err
	}
	return k8sClient.Delete(ctx, cur)
}

// CreatePulsarBroker creates a pulsarbroker in k8s
func CreatePulsarBroker(ctx context.Context, k8sClient client.Client, namespace, brokerName, image, zkServer string) error {
	broker := MakePulsarBroker(namespace, brokerName, image, zkServer, 1)
	return k8sClient.Create(ctx, broker)
}

// DeletePulsarBroker will delete the pulsarbroker in k8s
func DeletePulsarBroker(ctx context.Context, k8sClient client.Client, brokerName types.NamespacedName) error {
	cur := &pulsarv1alpha1.PulsarBroker{}
	err := k8sClient.Get(ctx, brokerName, cur)
	if err != nil {
		return err
	}
	return k8sClient.Delete(ctx, cur)
}

// CreatePulsarProxy creates a pulsarproxy in k8s
func CreatePulsarProxy(ctx context.Context, k8sClient client.Client, namespace, proxyName, image, brokerAddress string, serviceType commonv1alpha1.ServiceType) error {
	var externalService commonv1alpha1.ExternalServiceTemplate
	switch serviceType {
	case commonv1alpha1.ServiceTypeClusterIP:
		externalService = commonv1alpha1.ExternalServiceTemplate{
			Type: commonv1alpha1.ServiceTypeClusterIP,
		}
	case commonv1alpha1.ServiceTypeNodePort:
		externalService = commonv1alpha1.ExternalServiceTemplate{
			Type: commonv1alpha1.ServiceTypeNodePort,
			Ports: []commonv1alpha1.ServicePort{
				{Name: "http", Port: pointer.Int32Ptr(8080), NodePort: pointer.Int32Ptr(30080)},
			},
		}
	case commonv1alpha1.ServiceTypeLoadBalancer:
		externalService = commonv1alpha1.ExternalServiceTemplate{
			Type: commonv1alpha1.ServiceTypeLoadBalancer,
		}
	}
	proxy := MakePulsarProxy(namespace, proxyName, image, brokerAddress, 1, externalService)
	return k8sClient.Create(ctx, proxy)
}

// DeletePulsarProxy will delete the pulsarproxy in k8s
func DeletePulsarProxy(ctx context.Context, k8sClient client.Client, brokerName types.NamespacedName) error {
	cur := &pulsarv1alpha1.PulsarProxy{}
	err := k8sClient.Get(ctx, brokerName, cur)
	if err != nil {
		return err
	}
	return k8sClient.Delete(ctx, cur)
}
