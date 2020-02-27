# Develop Octopus

<!-- toc -->

- [Build management of Octopus](#build-management-of-octopus)
- [Build management of Adaptors](#build-management-of-adaptors)
- [Contributor workflow](#contributor-workflow)

<!-- /toc -->

## Build management of Octopus

Octopus takes inspiration from [Maven](https://maven.apache.org/) and provides a set of project build management based on [make](https://www.gnu.org/software/make/manual/make.html). Build management process consists of several stages, a stage consists of several actions. For convenience, the name of the action also represents the current stage. The overall flow relationship of action looks as below:

```text
    generate -> mod -> lint -> build -> test -> verify
                         \ = = = = = = = =  package  = = = = = = = = > e2e -> deploy
                            \ -> build -> test -> containerize -> /
```

Explanation of each action:

| Action | Usage |
|---:|:---|
| `generate` | Generate deployment manifests and deepcopy/runtime.Object implementations of `octopus` via [`controller-gen`](https://github.com/kubernetes-sigs/controller-tools/blob/master/cmd/controller-gen/main.go); Generate proto files of `adaptor` interfaces via [`protoc`](https://github.com/protocolbuffers/protobuf). |
| `mod` | Download `octopus` dependencies. |
| `lint` | Verify `octopus` via [`golangci-lint`](https://github.com/golangci/golangci-lint), roll back to `go fmt` and `go vet` if the installation fails. |
| `build` | Compile `octopus` according to the type and architecture of the OS, generate the binary into `bin` directory. |
| `test` | Run unit tests. |
| `verify` | Run integration tests. |
| `containerize` | Package Docker container. |
| `package` | Use [`dapper`](https://github.com/rancher/dapper) to execute `build`, `test` and `containerize` actions. |
| `e2e` | Run E2E tests. |
| `deploy` | Push Docker container. | 

Executing a stage can run `make octopus <stage name>`, for example, when executing the `test` stage, please run `make octopus test`. To execute a stage will execute all actions in the previous sequence, if running `make octopus test`, it actually includes executing `generate`, `mod`, `lint`, `build` and `test` actions.

To run an action by adding `only` command, for example, if only run `build` action, please run `make octopus build only`.

## Build management of Adaptors

The build management of Adaptors is the similar to Octopus, except that the execution is different. Executing a stage of any adaptor can run `make adaptor <adaptor name> <stage name>`. Please view [Develop Adaptors](../adaptors/develop.md) for more details.

## Contributor workflow

Contributing is welcome, please view [Contributing](../../CONTRIBUTING.md) for more details.
