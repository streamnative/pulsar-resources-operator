module github.com/streamnative/pulsar-resources-operator/tests

go 1.16

replace github.com/streamnative/pulsar-resources-operator => ../

require (
	github.com/mittwald/go-helm-client v0.8.2
	github.com/onsi/ginkgo v1.16.6-0.20211118180735-4e1925ba4c95
	github.com/onsi/gomega v1.17.0
	github.com/sirupsen/logrus v1.8.1
	github.com/streamnative/pulsar-resources-operator v0.0.0-00010101000000-000000000000
	helm.sh/helm/v3 v3.7.1
	k8s.io/api v0.22.4
	k8s.io/apimachinery v0.22.4
	k8s.io/client-go v0.22.4
	sigs.k8s.io/controller-runtime v0.10.3
)
