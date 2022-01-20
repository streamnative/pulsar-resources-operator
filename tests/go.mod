module github.com/streamnative/pulsar-resources-operator/tests

go 1.16

require (
	github.com/onsi/ginkgo/v2 v2.0.0
	github.com/onsi/gomega v1.17.0
	github.com/sirupsen/logrus v1.8.1
	github.com/streamnative/pulsar-operators/bookkeeper-operator v0.0.0-20211228075820-f17e933b1e80
	github.com/streamnative/pulsar-operators/commons v0.0.0
	github.com/streamnative/pulsar-operators/pulsar-operator v0.0.0-20211228075820-f17e933b1e80
	github.com/streamnative/pulsar-operators/zookeeper-operator v0.0.0-20211228075820-f17e933b1e80
	github.com/streamnative/pulsar-resources-operator v0.0.0-00010101000000-000000000000
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	k8s.io/api v0.22.4
	k8s.io/apimachinery v0.22.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a
	sigs.k8s.io/controller-runtime v0.10.3
)

replace (
	github.com/streamnative/pulsar-operators/commons => github.com/streamnative/pulsar-operators/commons v0.0.0-20211228075820-f17e933b1e80
	github.com/streamnative/pulsar-operators/pulsar-operator => ../../pulsar-operators/pulsar-operator
	github.com/streamnative/pulsar-resources-operator => ../
	k8s.io/client-go => k8s.io/client-go v0.22.4
)
