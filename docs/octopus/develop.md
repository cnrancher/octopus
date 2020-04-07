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
| `build` | Compile `octopus` according to the type and architecture of the OS, generate the binary into `bin` directory. <br/><br/> Use `CROSS=true` to compile binaries of the supported platforms(search `constant.sh` file in this repo). |
| `test` | Run unit tests. |
| `verify` | Run integration tests with a Kubernetes cluster. <br/><br/> Use `LOCAL_CLUSTER_KIND` to specify the type for local cluster, default is `k3d`. Instead of setting up a local cluster, you can also use environment variable `USE_EXISTING_CLUSTER=true` to point out an existing cluster, and then the integration tests will use the kubeconfig of the current environment to communicate with the existing cluster. |
| `package` | Package Docker image. |
| `e2e` | Run E2E tests. |
| `deploy` | Push Docker images and create manifest images for the current version. <br/><br/> Use `WITHOUT_MANIFEST=true` to prevent pushing manifest image, or `ONLY_MANIFEST=true` to push the manifest images only and `IGNORE_MISSING=true` to warn on missing images defined in platform list if needed. |

Executing a stage can run `make octopus <stage name>`, for example, when executing the `test` stage, please run `make octopus test`. To execute a stage will execute all actions in the previous sequence, if running `make octopus test`, it actually includes executing `generate`, `mod`, `lint`, `build` and `test` actions.

To run an action by adding `only` command, for example, if only run `build` action, please use `make octopus build only`.

Integrate with [`dapper`](https://github.com/rancher/dapper) via `BY` environment variable, for example, if only run `build` action via [`dapper`](https://github.com/rancher/dapper), please use `BY=dapper make octopus build only`. 

### Usage cases

Suppose to try the following example on Mac:

1. Run in the localhost, the current environment will install additional dependencies. You will get a corresponding warning if any installation fails.
    - `make octopus build`: execute `build` stage, then get a `darwin/amd64` execution binary.
    - `make octopus test only`: execute `test` action on `darwin/amd64` platform.
    - `REPO=somebody OS=linux ARCH=amd64 make octopus package`: execute `package` stage, then get a `linux/amd64` execution binary and an octopus `linux/amd64` image of `somebody` repo.
    - `LOCAL_CLUSTER_KIND=kind make octopus verify only`: execute `verify` action with [`kind`](https://github.com/kubernetes-sigs/kind) cluster.

1. Support multi-arch in the localhost.
    - `CROSS=true make octopus build only`: execute `build` action, then get all execution binaries of supported platform.
    - `CROSS=true make octopus test only`: _crossed testing isn't supported currently_.
    - `CROSS=true REPO=somebody make octopus package only`: execute `package` action, then get all supported platform images of `somebody` repo.
        + `make octopus package only`: _packaging `darwin` platform image isn't supported currently_.
    - `CROSS=true REPO=somebody make octopus deploy only`: execute `deploy` action, then push all supported platform images to `somebody` repo, also create [manifest images](https://docs.docker.com/engine/reference/commandline/manifest/) for the current version.
        + `make octopus deploy only`: _deploying `darwin` platform image isn't supported currently_.
    
1. In [`dapper`](https://github.com/rancher/dapper) mode, no additional dependencies are required in the current environment, which suitable for constructing CI / CD and good to the portability of the environment.
    - `BY=dapper make octopus build`: execute `build` stage, then get a `linux/amd64` execution binary.
    - `BY=dapper make octopus test`: execute `test` stage on `linux/amd64` platform.
    - `BY=dapper REPO=somebody make octopus package only`: execute `package` action and get an octopus `linux/amd64` image of `sombody` repo.

### Notes

In [`dapper`](https://github.com/rancher/dapper) mode:
- Using `USE_EXISTING_CLUSTER=true` is **NOT ALLOWED**.
- Using `LOCAL_CLUSTER_KIND=kind` instead of `k3d` a local cluster until fixed [k3d:issue#143](https://github.com/rancher/k3d/issues/143).

## Build management of Adaptors

The build management of Adaptors is the similar to Octopus, except that the execution is different. Executing a stage of any adaptor can run `make adaptor <adaptor name> <stage name>`. Please view [Develop Adaptors](../adaptors/develop.md) for more details.

## Contributor workflow

Contributing is welcome, please view [Contributing](../../CONTRIBUTING.md) for more details.
