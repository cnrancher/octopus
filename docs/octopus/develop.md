# Develop Octopus

<!-- toc -->

- [Build management of Octopus](#build-management-of-octopus)
- [Build management of Adaptors](#build-management-of-adaptors)
- [Contributor workflow](#contributor-workflow)

<!-- /toc -->

## Build management of Octopus

Octopus takes inspiration from [Maven](https://maven.apache.org/) and provides a set of project build management based on [make](https://www.gnu.org/software/make/manual/make.html). Build management process consists of several stages, a stage consists of several actions. For convenience, the name of the action also represents the current stage. The overall flow relationship of action looks as below:

```text
          generate -> mod -> lint -> build -> package -> deploy
                                         \ -> test -> verify -> e2e
```

Explanation of each action:

| Action | Usage |
|---:|:---|
| `generate` | Generate deployment manifests and deepcopy/runtime.Object implementations of `octopus` via [`controller-gen`](https://github.com/kubernetes-sigs/controller-tools/blob/master/cmd/controller-gen/main.go); Generate proto files of `adaptor` interfaces via [`protoc`](https://github.com/protocolbuffers/protobuf). |
| `mod` | Download `octopus` dependencies. |
| `lint` | Verify `octopus` via [`golangci-lint`](https://github.com/golangci/golangci-lint), roll back to `go fmt` and `go vet` if the installation fails. |
| `build` | Compile `octopus` according to the type and architecture of the OS, generate the binary into `bin` directory. <br/><br/> Use `CROSS=true` to compile binaries of the supported platforms(search the `constant.sh` in this repo). |
| `test` | Run unit tests. |
| `verify` | Run integration tests. |
| `package` | Package Docker image. |
| `e2e` | Run E2E tests. |
| `deploy` | Push Docker images, and create manifest images. |

Executing a stage can run `make octopus <stage name>`, for example, when executing the `test` stage, please run `make octopus test`. To execute a stage will execute all actions in the previous sequence, if running `make octopus test`, it actually includes executing `generate`, `mod`, `lint`, `build` and `test` actions.

To run an action by adding `only` command, for example, if only run `build` action, please use `make octopus build only`.

Integrate with [`dapper`](https://github.com/rancher/dapper) via `BY` environment variable, for example, if only run `build` action via [`dapper`](https://github.com/rancher/dapper), please use `BY=dapper make octopus build only`

### Usage example

1. `make octopus build` on Mac: execute `generate`, `mod`, `lint` and `build` stages, and get a `darwin/amd64` execution binary on `bin` directory.
1. `CROSS=true make octopus build only` on Mac: execute `build` stage, and get all execution binaries of supported platform on `bin` directory.
1. `BY=dapper make octopus build only`: execute `build` stage with [`dapper`](https://github.com/rancher/dapper), and get a `linux/amd64` execution binary on `bin` directory.
1. `make octopus test only` on Mac: execute `test` stage, and run the unit testing on `darwin/amd64` platform.
1. `CROSS=true make octopus test only` on Mac: execute `test` stage, _crossed testing isn't supported currently_.
1. `REPO=somebody OS=linux ARCH=amd64 make octopus package`: execute `package` stage, then get a `linux/amd64` execution binary on `bin` directory, also get an octopus `linux/amd64` image of `somebody` repo.
1. `CROSS=true REPO=somebody make octopus package only`: execute `package` stage, then get all execution binaries of supported platform on `bin` directory, also get all supported platform images of `somebody` repo.
1. `REPO=somebody make octopus deploy only`: execute `deploy` stage, then push all supported platform images to Docker hub, and create manifest image for the current version and `latest`.

## Build management of Adaptors

The build management of Adaptors is the similar to Octopus, except that the execution is different. Executing a stage of any adaptor can run `make adaptor <adaptor name> <stage name>`. Please view [Develop Adaptors](../adaptors/develop.md) for more details.

## Contributor workflow

Contributing is welcome, please view [Contributing](../../CONTRIBUTING.md) for more details.
