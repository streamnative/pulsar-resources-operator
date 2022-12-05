module github.com/streamnative/pulsar-resources-operator/tests

go 1.16

require (
	// pulsar-client-go latest version is v0.9.0,
	// it removed the go.mod from oauth2, this will cause the issue of ambiguous import
	github.com/apache/pulsar-client-go v0.8.1
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.19.0
	github.com/sirupsen/logrus v1.8.1
	github.com/streamnative/pulsar-resources-operator v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.0 // indirect
	golang.org/x/sys v0.0.0-20220204135822-1c1b9b1eba6a // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.11.0
)

replace (
	github.com/streamnative/pulsar-operators/commons => github.com/streamnative/pulsar-operators/commons v0.0.0-20211228075820-f17e933b1e80
	github.com/streamnative/pulsar-resources-operator => ../
	k8s.io/client-go => k8s.io/client-go v0.22.4
)
