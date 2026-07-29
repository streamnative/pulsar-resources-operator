<!--
	
	  Copyright 2023 StreamNative, Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.

-->

# Contributing guidelines

## Contributing steps
1. Submit an issue describing your proposed change.
2. Discuss and wait for proposal to be accepted.
3. Fork this repo, develop and test your code changes.
4. Submit a pull request.

## Conventions

Please read through below conventions before contributions.

### PullRequest conventions

- Use [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/) to standardize PR title.

### Code conventions

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
- Know and avoid [Go landmines](https://gist.github.com/lavalamp/4bd23295a9f32706a48f)
- Commenting
  - [Go's commenting conventions](http://blog.golang.org/godoc-documenting-go-code)
  - If reviewers ask questions about why the code is the way it is, that's a sign that comments might be helpful.
- Naming
  - Please consider package name when selecting an interface name, and avoid redundancy. For example, `storage.Interface` is better than `storage.StorageInterface`.
  - Do not use uppercase characters, underscores, or dashes in package names.
  - Please consider parent directory name when choosing a package name. For example, `pkg/controllers/autoscaler/foo.go` should say `package autoscaler` not `package autoscalercontroller`.
      - Unless there's a good reason, the `package foo` line should match the name of the directory in which the `.go` file exists.
      - Importers can use a different name if they need to disambiguate.
  - Locks should be called `lock` and should never be embedded (always `lock sync.Mutex`). When multiple locks are present, give each lock a distinct name following Go conventions: `stateLock`, `mapLock` etc.
  
### Folder and file conventions

- All filenames should be lowercase.
- Go source files and directories use underscores, not dashes.
  - Package directories should generally avoid using separators as much as possible. When package names are multiple words, they usually should be in nested subdirectories.
- Documentation filenames are lowercase and currently follow the existing `docs/*.md` underscore convention. Match neighboring files when adding new documents.
- All source files should add a license at the beginning.


### How to work locally

1. Install the Go version declared by `go.work` and clone this repository.
2. Run unit and envtest coverage with `make test`.
3. For end-to-end tests, create a cluster, for example `minikube start --memory=8192 --cpus=4`.
4. [Deploy Apache Pulsar](https://pulsar.apache.org/docs/4.0.x/getting-started-helm/#step-1-install-pulsar-helm-chart).
5. Apply the operator CRDs with `make install`.
6. Run the operator locally with `make run`.
7. In another terminal, run the integration suite:

   ```shell
   cd tests
   go run github.com/onsi/ginkgo/v2/ginkgo --trace ./operator
   ```

   Set `ADMIN_SERVICE_URL`, `NAMESPACE`, `BROKER_NAME`, and `PROXY_URL` as described in [`tests/README.md`](tests/README.md) when defaults do not match the test cluster.
