# Octopus

[![Build Status](http://drone-pandaria.cnrancher.com/api/badges/cnrancher/octopus/status.svg)](http://drone-pandaria.cnrancher.com/cnrancher/octopus)
[![Go Report Card](https://goreportcard.com/badge/github.com/cnrancher/octopus)](https://goreportcard.com/report/github.com/cnrancher/octopus)

Octopus is a light-weight and cloud-native device management system for Kubernetes and k3s, it does not need to replace any basic components of the Kubernetes cluster. After Octopus deployed, the cluster can have the ability to manage edge devices as custom k8s resources.

<!-- toc -->

- [Idea](#idea)
- [Documentation](#documentation)
- [Quick-start](#quick-start)
- [Source code](#source-code)
- [License](#license)

<!-- /toc -->

## Idea

Like a real octopus, Octopus consists of the `brain` and `limbs`. The `brain` only needs to deploy one or automatically select a leader in HA mode. It only needs to process relatively concentrated information, such as verifying the existence of nodes and the existence of device models (types). Limbs need to be deployed on each edge node that can be connected to the device, and they communicate with the actual device through the device adaptor (Adaptors). Therefore, Octopus uses a DeviceLink YAML file (a custom-defined k8s object) to configure and manage its device connections.

For more details please refer to the [official documentation](https://cnrancher.github.io/docs-octopus/eng/).

## Documentation

<!-- toc -->
- Octopus
    - [About Octopus](https://cnrancher.github.io/docs-octopus/docs/en/about)
    - [Quick-start guide](https://cnrancher.github.io/docs-octopus/docs/en/quick-start)
    - [How to develop](https://cnrancher.github.io/docs-octopus/docs/en/develop)
    - [How to monitor](https://cnrancher.github.io/docs-octopus/docs/en/monitoring)
    - [The state transition of DeviceLink](https://cnrancher.github.io/docs-octopus/docs/en/devicelink/state-of-dl)
- Adaptors
    - [How it works](https://cnrancher.github.io/docs-octopus/docs/en/adaptors/adaptor)
    - [How to develop](https://cnrancher.github.io/docs-octopus/docs/en/adaptors/develop)
- Contribution
    - [How to contribute](./CONTRIBUTING.md)
<!-- /toc -->

## Quick-start

There are two ways to deploy the Octopus, for quick-start, you can use the manifest YAML file to bring up the Octopus. The installer YAML file is under the [deploy/e2e](./deploy/e2e) directory on Github.
```shell script
# install octopus
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/deploy/e2e/all_in_one.yaml

# install ui
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus-api-server/master/deploy/e2e/all_in_one.yaml

# install adaptors
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/modbus/deploy/e2e/all_in_one.yaml
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/opcua/deploy/e2e/all_in_one.yaml
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/mqtt/deploy/e2e/all_in_one.yaml
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/ble/deploy/e2e/all_in_one.yaml
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/dummy/deploy/e2e/all_in_one.yaml
```

Optionally, you can use this [repository](https://github.com/cnrancher/octopus-chart) hosts official Helm charts for Octopus. These charts are used to deploy Octopus to the Kubernetes/k3s Cluster.
```shell script
# add octopus helm repo
$ helm repo add octopus http://charts.cnrancher.com/octopus
$ helm repo update

# create octopus-system namespace
$ kubectl create ns octopus-system

# install octopus, ui and adaptors
$ helm install --namespace octopus-system octopus octopus/octopus
```

## Source code
Octopus is 100% open source software. Project source code is spread across a number of repos:

| Name | Repo Address |
|:---|:---|
| Octopus UI | https://github.com/cnrancher/octopus-ui |
| Octopus API Server | https://github.com/cnrancher/octopus-api-server |
| Octopus Chart | https://github.com/cnrancher/octopus-chart |
| Octopus Simulator | https://github.com/cnrancher/octopus-simulator |
| Octopus Docs | https://github.com/cnrancher/docs-octopus |

## License
Copyright (c) 2020 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at [LICENSE](./LICENSE) file for details.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
