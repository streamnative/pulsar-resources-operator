module github.com/streamnative/pulsar-resources-operator

go 1.16

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/apache/pulsar-client-go/oauth2 v0.0.0-20211114184309-7773d27562ce
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/streamnative/pulsar-operators/commons v0.0.0-20211125071320-d30ed8de4901
	github.com/streamnative/pulsarctl v0.4.3-0.20211125034318-d6febda9e96b
	go.opentelemetry.io/contrib/propagators/aws v0.20.0 // indirect
	go.uber.org/zap v1.19.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
	sigs.k8s.io/controller-runtime v0.10.0
)
