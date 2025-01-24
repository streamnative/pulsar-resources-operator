module github.com/streamnative/pulsar-resources-operator

go 1.22.7

require (
	github.com/apache/pulsar-client-go v0.13.0-candidate-1.0.20240827063209-953d9eab0794
	github.com/go-logr/logr v1.4.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 v2.6.0
	github.com/onsi/gomega v1.24.1
	github.com/xhit/go-str2duration/v2 v2.1.0
	go.uber.org/zap v1.24.0
	k8s.io/api v0.26.0
	k8s.io/apiextensions-apiserver v0.26.0
	k8s.io/apimachinery v0.26.0
	k8s.io/client-go v0.26.0
	k8s.io/component-base v0.26.0
	k8s.io/utils v0.0.0-20221128185143-99ec85e7a448
	sigs.k8s.io/controller-runtime v0.14.0
	github.com/streamnative/cloud-api-server v1.25.2-0.20250102073730-2589f39cb5f2
)

require (
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/oauth2 v0.12.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/term v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

// OTEL
replace (
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace => github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.19.1
	github.com/streamnative/kube-instrumentation => github.com/streamnative/kube-instrumentation v0.3.2
	go.opentelemetry.io/contrib/detectors/gcp => go.opentelemetry.io/contrib/detectors/gcp v1.19.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.44.0
	go.opentelemetry.io/contrib/propagators/aws => go.opentelemetry.io/contrib/propagators/aws v1.19.0
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace => go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0
	go.opentelemetry.io/otel/exporters/stdout => go.opentelemetry.io/otel/exporters/stdout v0.19.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace => go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.19.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.19.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.19.0
	go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.19.0
	go.opentelemetry.io/otel/sdk/trace => go.opentelemetry.io/otel/sdk/trace v1.19.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.19.0
)

// Pulsar Operator
replace (
	github.com/operator-framework/api => github.com/operator-framework/api v0.14.0
	github.com/streamnative/sn-operator/api => github.com/streamnative/sn-operator/api v0.8.0-rc.27
	github.com/streamnative/sn-operator/api/commons => github.com/streamnative/sn-operator/api/commons v0.8.0-rc.27
	github.com/streamnative/sn-operator/pkg/commons => github.com/streamnative/sn-operator/pkg/commons v0.8.0-rc.27
)

// kubernetes (apiserver, controller-runtime)
replace (
	github.com/google/cel-go => github.com/google/cel-go v0.16.1
	go.etcd.io/etcd/api/v3 => go.etcd.io/etcd/api/v3 v3.5.9
	go.etcd.io/etcd/client/pkg/v3 => go.etcd.io/etcd/client/pkg/v3 v3.5.9
	go.etcd.io/etcd/client/v3 => go.etcd.io/etcd/client/v3 v3.5.9
	google.golang.org/grpc => google.golang.org/grpc v1.56.3
	k8s.io/api => k8s.io/api v0.28.15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.28.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.28.15
	k8s.io/apiserver => k8s.io/apiserver v0.28.15
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.28.15
	k8s.io/client-go => k8s.io/client-go v0.28.15
	k8s.io/code-generator => k8s.io/code-generator v0.28.15
	k8s.io/component-base => k8s.io/component-base v0.28.15
	k8s.io/gengo => k8s.io/gengo v0.0.0-20220902162205-c0856e24416d
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.100.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.28.15
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9
	k8s.io/kubectl => k8s.io/kubectl v0.28.15
	k8s.io/utils => k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.16.6
	sigs.k8s.io/json => sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd
	sigs.k8s.io/structured-merge-diff/v4 => sigs.k8s.io/structured-merge-diff/v4 v4.2.3
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.3.0
)

// pinned to v0.11.1 to align with kubernetes dependencies
replace sigs.k8s.io/external-dns => sigs.k8s.io/external-dns v0.11.1