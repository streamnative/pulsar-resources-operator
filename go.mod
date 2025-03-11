module github.com/streamnative/pulsar-resources-operator

go 1.22.11

toolchain go1.24.0

require (
	github.com/apache/pulsar-client-go v0.14.0
	github.com/go-logr/logr v1.4.2
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 v2.21.0
	github.com/onsi/gomega v1.35.1
	github.com/xhit/go-str2duration/v2 v2.1.0
	go.uber.org/zap v1.27.0
	golang.org/x/oauth2 v0.21.0
	k8s.io/api v0.30.9
	k8s.io/apiextensions-apiserver v0.30.1
	k8s.io/apimachinery v0.30.9
	k8s.io/client-go v0.30.9
	k8s.io/component-base v0.30.9
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	sigs.k8s.io/controller-runtime v0.18.2
)

require (
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/evanphx/json-patch v5.9.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20241029153458-d1b30febd7db // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.2 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

// OTEL
replace (
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace => github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.19.1
	github.com/streamnative/kube-instrumentation => github.com/streamnative/kube-instrumentation v0.3.2
	go.opentelemetry.io/contrib/detectors/gcp => go.opentelemetry.io/contrib/detectors/gcp v1.19.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.44.0
	go.opentelemetry.io/contrib/propagators/aws => go.opentelemetry.io/contrib/propagators/aws v1.19.0
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace => go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0
	go.opentelemetry.io/otel/exporters/stdout => go.opentelemetry.io/otel/exporters/stdout v0.19.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace => go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.19.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.21.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.21.0
	go.opentelemetry.io/otel/sdk/trace => go.opentelemetry.io/otel/sdk/trace v1.21.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.21.0
)

// kubernetes (apiserver, controller-runtime)
replace (
	google.golang.org/grpc => google.golang.org/grpc v1.58.3

	k8s.io/api => k8s.io/api v0.30.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.30.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.30.9
	k8s.io/client-go => k8s.io/client-go v0.30.9
	k8s.io/component-base => k8s.io/component-base v0.30.9
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.110.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00
	k8s.io/utils => k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.18.6
	sigs.k8s.io/json => sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd
	sigs.k8s.io/structured-merge-diff/v4 => sigs.k8s.io/structured-merge-diff/v4 v4.4.1
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.18.0
	github.com/prometheus/common => github.com/prometheus/common v0.45.0
)

// code generate from api-server
replace k8s.io/code-generator => k8s.io/code-generator v0.30.9
