# tests

tests is an individul module beside pulsar resources operator

`go mod tidy` to download modules for tests


## Requirements
- Pulsar Operator installed
- A pulsar cluster installed without authentication and authorization


## Run tests

`ginkgo --trace --progress ./operator`

Optionally, if you have an external pulsar cluster (e.g. deployed on minikube) and you want to test the operator without deploying it in kubernetes:

1. Run the code in a terminal

```bash
make install
go run .
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
export PROXY_URL=http://localhost:80

ginkgo --trace --progress ./operator
```
