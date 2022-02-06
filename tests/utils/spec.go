// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.
package utils

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bkv1alpha1 "github.com/streamnative/pulsar-operators/bookkeeper-operator/api/v1alpha1"
	commonv1alpha1 "github.com/streamnative/pulsar-operators/commons/api/v1alpha1"
	pulsarv1alpha1 "github.com/streamnative/pulsar-operators/pulsar-operator/api/v1alpha1"
	zkv1alpha1 "github.com/streamnative/pulsar-operators/zookeeper-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-operators/zookeeper-operator/pkg/utils"
	"github.com/streamnative/pulsar-resources-operator/api/v1alpha2"
)

// MakeZKCluster will generate a object of ZooKeeperCluster
func MakeZKCluster(namespace, name, image string) *zkv1alpha1.ZooKeeperCluster {
	zk := &zkv1alpha1.ZooKeeperCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Annotations: map[string]string{
				utils.AnnotationIgnoreLeaderCheck: "true",
			},
		},
		Spec: zkv1alpha1.ZooKeeperClusterSpec{
			Replicas:        3,
			Image:           image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Persistence: &zkv1alpha1.PersistentVolume{
				VolumeReclaimPolicy: zkv1alpha1.VolumeReclaimPolicyDelete,
				Data: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceStorage: resource.MustParse("1G"),
						},
					},
				},
				DataLog: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceStorage: resource.MustParse("1G"),
						},
					},
				},
			},

			Pod: zkv1alpha1.PodPolicy{
				Resources: corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:    resource.MustParse("50m"),
						corev1.ResourceMemory: resource.MustParse("384Mi"),
					},
					Limits: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:    resource.MustParse("500m"),
						corev1.ResourceMemory: resource.MustParse("384Mi"),
					},
				},
				JvmOptions: &zkv1alpha1.JVMOptions{
					MemoryOptions: []string{
						"-Xms64m",
						"-Xmx64m",
						"-XX:MaxDirectMemorySize=64m",
					},
				},
			},
		},
	}
	return zk
}

// MakeBookKeeperCluster will generate a object of BookKeeperCluster
func MakeBookKeeperCluster(namespace, name, zkServers, image string, replicas int32) *bkv1alpha1.BookKeeperCluster {
	bk := &bkv1alpha1.BookKeeperCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: bkv1alpha1.BookKeeperClusterSpec{
			Replicas:        &replicas,
			ZkServers:       zkServers,
			Image:           image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Pod: bkv1alpha1.PodPolicy{
				Resources: corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:    resource.MustParse("50m"),
						corev1.ResourceMemory: resource.MustParse("512Mi"),
					},
					Limits: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:    resource.MustParse("500m"),
						corev1.ResourceMemory: resource.MustParse("512Mi"),
					},
				},
			},
			Storage: &bkv1alpha1.BookieStorage{
				VolumeReclaimPolicy: bkv1alpha1.VolumeReclaimPolicyDelete,
				Journal: bkv1alpha1.StorageSpec{
					NumDirsPerVolume: 1,
					NumVolumes:       1,
					VolumeClaimSpec: &corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(bkv1alpha1.DefaultVolumeSize),
							},
						},
					},
				},
				Ledger: bkv1alpha1.StorageSpec{
					NumDirsPerVolume: 1,
					NumVolumes:       1,
					VolumeClaimSpec: &corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(bkv1alpha1.DefaultVolumeSize),
							},
						},
					},
				},
			},
		},
	}
	return bk

}

// MakePulsarConnection will generate a object of PulsarConnection, without authentication
func MakePulsarConnection(namespace, name, adminServiceURL string) *v1alpha2.PulsarConnection {
	return &v1alpha2.PulsarConnection{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha2.PulsarConnectionSpec{
			AdminServiceURL: adminServiceURL,
		},
	}
}

// MakePulsarTenant will generate a object of PulsarTenant
func MakePulsarTenant(namespace, name, tenantName, connectionName string, adminRoles, allowedClusters []string, policy v1alpha2.PulsarResourceLifeCyclePolicy) *v1alpha2.PulsarTenant {
	return &v1alpha2.PulsarTenant{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha2.PulsarTenantSpec{
			Name: tenantName,
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			AdminRoles:      adminRoles,
			AllowedClusters: allowedClusters,
			LifecyclePolicy: policy,
		},
	}
}

// MakePulsarNamespace will generate a object of PulsarNamespace
func MakePulsarNamespace(namespace, name, namespaceName, connectionName string, policy v1alpha2.PulsarResourceLifeCyclePolicy) *v1alpha2.PulsarNamespace {
	backlogSize := resource.MustParse("1Gi")
	bundle := int32(16)
	return &v1alpha2.PulsarNamespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha2.PulsarNamespaceSpec{
			Name: namespaceName,
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			BacklogQuotaLimitTime: &metav1.Duration{
				Duration: time.Hour * 24,
			},
			BacklogQuotaLimitSize: &backlogSize,
			Bundles:               &bundle,
			MessageTTL: &metav1.Duration{
				Duration: time.Hour * 1,
			},
		},
	}
}

// MakePulsarTopic will generate a object of PulsarTopic
func MakePulsarTopic(namespace, name, topicName, connectionName string, policy v1alpha2.PulsarResourceLifeCyclePolicy) *v1alpha2.PulsarTopic {
	return &v1alpha2.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha2.PulsarTopicSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			Name: topicName,
		},
	}
}

// MakePulsarPermission will generate a object of PulsarPermission
func MakePulsarPermission(namespace, name, resourceName, connectionName string, resourceType v1alpha2.PulsarResourceType,
	roles, actions []string, policy v1alpha2.PulsarResourceLifeCyclePolicy) *v1alpha2.PulsarPermission {
	return &v1alpha2.PulsarPermission{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1alpha2.PulsarPermissionSpec{
			ConnectionRef: corev1.LocalObjectReference{
				Name: connectionName,
			},
			ResourceName: resourceName,
			ResoureType:  resourceType,
			Roles:        roles,
			Actions:      actions,
		},
	}
}

// MakePulsarBroker will generate a object of PulsarBroker
func MakePulsarBroker(namespace, name, image, zkServer string, replicas int32) *pulsarv1alpha1.PulsarBroker {
	return &pulsarv1alpha1.PulsarBroker{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: pulsarv1alpha1.PulsarBrokerSpec{
			ZkServers: zkServer,
			Image:     image,
			Replicas:  replicas,
		},
	}
}

// MakePulsarProxy will generate a object of PulsarProxy, expose specific type service
// default config: tls disabled
func MakePulsarProxy(namespace, name, image, brokerAddress string, replicas int32, externalService commonv1alpha1.ExternalServiceTemplate) *pulsarv1alpha1.PulsarProxy {
	st := &commonv1alpha1.ServiceTemplate{
		ObjectMeta: commonv1alpha1.ObjectMeta{
			Name: name + "-proxy-external",
		},
	}
	externalService.ServiceTemplate = st

	return &pulsarv1alpha1.PulsarProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: pulsarv1alpha1.PulsarProxySpec{
			APIObjects: &pulsarv1alpha1.ProxyAPIObjects{
				ExternalService: &externalService,
			},
			BrokerAddress: brokerAddress,
			Image:         image,
			Replicas:      replicas,
			DNSNames:      []string{},
			Config: &pulsarv1alpha1.PulsarProxyConfig{
				TLS: &pulsarv1alpha1.TLSConfig{
					Enabled: false,
				},
			},
		},
	}
}
