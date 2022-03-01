# tests

tests is an individul module beside pulsar resources operator

`go mod tidy` to download modules for tests


# Requirements
- Pulsar Operator installed
- A pulsar cluster installed without authentication and authorization


# Run tests

`ginkgo --trace --progress ./operator`