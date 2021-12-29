// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.
package operator_test

import (
	"context"
	_ "embed"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bkv1alpha1 "github.com/streamnative/pulsar-operators/bookkeeper-operator/api/v1alpha1"
	commonsv1alpha1 "github.com/streamnative/pulsar-operators/commons/api/v1alpha1"
	pulsarv1alpha1 "github.com/streamnative/pulsar-operators/pulsar-operator/api/v1alpha1"
	zkv1alpha1 "github.com/streamnative/pulsar-operators/zookeeper-operator/api/v1alpha1"
	v1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/tests/utils"
)

var (
	namespaceName string = "default"
	k8sClient     client.Client
	nodeIP        string
	k8sConfig     *rest.Config
	initialImage  string = "apachepulsar/pulsar:latest"
	zkClusterName string = "testzk"
	zkServer      string = zkClusterName + "-zk:2181"
	bkClusterName string = "testbk"
	brokerName    string = "testbroker"
	brokerImage   string = "apachepulsar/pulsar:latest"
	brokerAddress string = brokerName + "-broker"
	proxyName     string = "testproxy"
	proxyImage    string = "apachepulsar/pulsar:latest"
)

func TestOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operator Suite")
}

var _ = BeforeSuite(func() {
	var err error
	k8sConfig, err = utils.GetKubeConfig("")
	Expect(err).Should(Succeed())
	Expect(bkv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(pulsarv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(zkv1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())

	k8sClient, err = client.New(k8sConfig, client.Options{})
	Expect(err).Should(Succeed())
	logrus.Info("Create K8s Client successfully")

	ctx := context.TODO()
	err = utils.CreateZKCluster(ctx, k8sClient, namespaceName, zkClusterName, initialImage)
	Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
	logrus.Info("Create zk cluster successfully")

	err = utils.CreateBKCluster(ctx, k8sClient, namespaceName, bkClusterName, zkServer, initialImage)
	Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
	logrus.Info("Create bk cluster successfully")

	err = utils.CreatePulsarBroker(ctx, k8sClient, namespaceName, brokerName, brokerImage, zkServer)
	Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
	logrus.Info("Create broker successfully")

	err = utils.CreatePulsarProxy(ctx, k8sClient, namespaceName, proxyName, proxyImage, brokerAddress, commonsv1alpha1.ServiceTypeNodePort)
	if err != nil {
		logrus.Info(err)
	}
	Expect(err == nil || apierrors.IsAlreadyExists(err)).Should(BeTrue())
	logrus.Info("Create proxy successfully")
})

var _ = AfterSuite(func() {
	ctx := context.TODO()

	bName := types.NamespacedName{Namespace: namespaceName, Name: brokerName}
	err := utils.DeletePulsarBroker(ctx, k8sClient, bName)
	Expect(err).Should(Succeed())
	logrus.Info("Delete pulsarbroker successfully")

	pName := types.NamespacedName{Namespace: namespaceName, Name: proxyName}
	err = utils.DeletePulsarProxy(ctx, k8sClient, pName)
	Expect(err).Should(Succeed())
	logrus.Info("Delete pulsarproxy successfully")

	bkName := types.NamespacedName{Namespace: namespaceName, Name: bkClusterName}
	err = utils.DeleteBKCluster(ctx, k8sClient, bkName)
	Expect(err).Should(Succeed())
	logrus.Info("Delete bk cluster successfully")

	zkName := types.NamespacedName{Namespace: namespaceName, Name: zkClusterName}
	err = utils.DeleteZKCluster(ctx, k8sClient, zkName)
	Expect(err).Should(Succeed())
	logrus.Info("Delete zk cluster successfully")
})
