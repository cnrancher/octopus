# Develop Adaptor

<!-- toc -->

- [Scaffold](#scaffold)
- [Build management](#build-management)
- [Contributor workflow](#contributor-workflow)

<!-- /toc -->

## Scaffold

Octopus provides a simple way to develop a new adaptor, with running `make template-adaptor`, get a scaffold under `adaptors` directory. The overlay of the scaffold is as follows:

```text
tree -d adaptors/<adaptor-name>
├── api                             ---  device model CRD
│   └── v1alpha1                    ------  implement the logic*
├── bin                             ---  output of `go build`
├── cmd                             ---  command entry code
│   └── <adaptor-name>              ------  implement the logic*
├── deploy                          ---  deployment manifest
│   ├── e2e                         ------  output of `kubectl kustomize` and demo cases
│   └── manifests                   ------  overlay for kustomize
├── dist                            ---  output of `go test` and versioned deployment manifest
├── hack                            ---  bash scripts for make rules
├── pkg                             ---  core code
│   ├── adaptor
│   └── <adaptor-name>              ------  implement the logic*
└── test                            ---  test code
    ├── e2e
    └── integration
```

## Build management

Adaptor follows the build management of Octopus, please view [Develop Octopus](../octopus/develop.md) for more details. Same as Octopus, Adaptor's management process consists of several stages, a stage consists of several actions. For convenience, the name of the action also represents the current stage. The overall flow relationship of action looks as below:
                                        
```text
        generate -> mod -> lint -> build -> package -> deploy
                                       \ -> test -> verify -> e2e
```

Executing a stage for an Adaptor can run `make adaptor <adatpor-name> <stage name>`, for example, when executing the `test` stage for [dummy](../../adaptors/dummy) adaptor, please run `make adaptor dummy test`. To execute a stage will execute all actions in the previous sequence, if running `make adaptor dummy test`, it actually includes executing `generate`, `mod`, `lint`, `build` and `test` actions.

To run an action by adding `only` command, for example, if only run `build` action, please use `make adaptor <adatpor-name> build only`.

Integrate with [`dapper`](https://github.com/rancher/dapper) via `BY` environment variable, for example, if only run `build` action via [`dapper`](https://github.com/rancher/dapper), please use `BY=dapper make adaptor <adatpor-name> build only`. 

## Contributor workflow

Contributing is welcome, please view [CONTRIBUTING](../../CONTRIBUTING.md) for more details.
