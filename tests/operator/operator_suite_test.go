// Copyright 2022 StreamNative
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

package operator_test

import (
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/tests/utils"
)

var (
	namespaceName string = "default"
	k8sClient     client.Client
	k8sConfig     *rest.Config
	brokerName    string = "test-pulsar"
	proxyURL      string = fmt.Sprintf("http://%s-broker.%s.svc.cluster.local:6650", brokerName, namespaceName)
	pulsarClient  pulsar.Client
)

func TestOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operator Suite")
}

var _ = BeforeSuite(func() {
	var err error
	k8sConfig, err = utils.GetKubeConfig("")
	Expect(err).Should(Succeed())
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).Should(Succeed())

	k8sClient, err = client.New(k8sConfig, client.Options{})
	Expect(err).Should(Succeed())
	logrus.Info("Create K8s Client successfully")

	pulsarClient, err = pulsar.NewClient(pulsar.ClientOptions{
		URL:               proxyURL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	Expect(err).Should(BeNil())
})

var _ = AfterSuite(func() {
	pulsarClient.Close()
})
