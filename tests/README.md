# Integration tests

`tests` is a separate Go module included by the repository's `go.work` file.

Run `go mod download` from this directory to prefetch its dependencies without rewriting module files.


## Requirements
- Pulsar Operator installed
- A pulsar cluster installed without authentication and authorization


## Run tests

```bash
go run github.com/onsi/ginkgo/v2/ginkgo --trace ./operator
```

Optionally, if you have an external pulsar cluster (e.g. deployed on minikube) and you want to test the operator without deploying it in kubernetes:

1. Run the code in a terminal

```bash
make install
make run
```

2. In another terminal run

```bash
# your admin service url
export ADMIN_SERVICE_URL=http://localhost:80
# your pulsar namespace
export NAMESPACE=pulsar
# your pulsar broker name
export BROKER_NAME=pulsar-mini
# your pulsar proxy url
export PROXY_URL=pulsar://localhost:6650

cd tests
go run github.com/onsi/ginkgo/v2/ginkgo --trace ./operator
```
