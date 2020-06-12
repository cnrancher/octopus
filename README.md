# Octopus

[![Build Status](http://drone-pandaria.cnrancher.com/api/badges/cnrancher/octopus/status.svg)](http://drone-pandaria.cnrancher.com/cnrancher/octopus)

Octopus is an edge device management system based on Kubernetes, it is very lightweight and does not need to replace any of the basic components of the Kubernetes clusters. With Octopus deployed, the cluster can have the ability to manage any edge device as a resource.

<!-- toc -->

- [Idea](#idea)
- [Workflow](#workflow)
- [Walkthrough](#walkthrough)
    + [Deploy Octopus](#deploy-octopus)
    + [Deploy Device Model & Device Adaptor](#deploy-device-model--device-adaptor)
    + [Create DeviceLink](#create-devicelink)
    + [Manage Device](#manage-device)
- [Documentation](#documentation)
- [License](#license)

<!-- /toc -->

## Idea

Like the real octopus, Octopus consists of `brain` and `limb`s. The `brain` only needs to deploy one or choose the leader, it is responsible for processing some relatively centralized information, such as judging whether the node exists, whether the device model(type) exists, etc. The `limb`s need to deploy on edge nodes that the device can connect to, they talk to devices in the real world through `adaptors`. Therefore, Octopus manages devices by managing the device connections(`DeviceLink`).

## Workflow

```text
                                                                                                                   
    │          metadata         │                    edge node                      │      devices      │          
   ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─         
    │                           │                                                   │                   │          
                                                                                                                   
                                                                                                        │          
                                                                                                                   
        ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐                                          ┌───────────┐                    │          
          <<Device Model>>                                          ┌─▶│  adaptor  ├┐  6                           
     ┌──│        CRD        │                                     4 │  └┬──────────┘│◀──┐               │          
     │   ─ ─ ─ ─ ─ ─ ─ ─ ─ ─                                        │   └───────────┘   │                          
     │                                                              │                   │     .         │          
    1│                                                              │                   └───▶( )          user     
     │  ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐                                       │                     5   '         │          
     │       DeviceLink                                             │                                              
     │  ├───────────────────┤                                       │                                   │          
     └─▷│       Model       │                                       │                                              
        ├───────────────────┤                                       │                                   │          
        │      Adaptor      │                                       │                                              
        ├───────────────────┤                                       │                                   │          
        │     Template      │─────────────┬─────────────────┐       │                                              
        └───────────────────┘            2│                3│       │                                   │          
                                          │                 │       │                                  ─ ─         
                                          ▼                 │       └─────┐                             │          
                                ┌───────────────────┐       │             │                                        
                                │       brain       │       │             │                             │          
                                └───────────────────┘       │             │                                        
                                 │                          │             │                             │          
                                 ├─▣  node existed?         │             │                                        
                                 │   ────────────────       │             │                             │          
                                 │                          │             │                                        
                                 └─▣  model existed?        │             │                             │          
                                     ────────────────       │             │                                        
                                                            │             │                             │          
                                                            │             │                               octopus  
                                                            ▼             │                             │          
        ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐                     ┌───────────────────┐   │                                        
          <<Device Model>>             ┌──────────│       limb        ├┐  │ 7                           │          
        │     Instance      │          │          └┬──────────────────┘│◀─┘                                        
        ┌───────────────────┐   8      │           └┬──────────────────┘                                │          
        │       Spec        │◀─────────┘            │                                                              
        ├───────────────────┤                       ├─▣ adaptor existed?                                │          
        │      Status       │                       │   ─────────────────                                          
        └───────────────────┘                       │                                                   │          
                                                    ├─▣  device created?                                           
                                                    │   ─────────────────                               │          
                                                    │                                                              
                                                    └─▣ device connected?                               │          
                                                        ─────────────────                              ─ ─         
                                                                                                        │          
```

## Walkthrough

In this walkthrough, we try to use Octopus to manage a dummy device. We will perform the following steps in order:

1. Deploy Octopus
1. Deploy Device Model & Device Adaptor
1. Create DeviceLink
1. Manage Device

### Deploy Octopus

There are [two ways](docs/octopus/deployment.md) to deploy Octopus, for convenience, we use the deployment manifest file to bring up the Octopus. The installer YAML file is under the [`deploy/e2e`](./deploy/e2e) directory:

```shell script
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/deploy/e2e/all_in_one.yaml
namespace/octopus-system created
customresourcedefinition.apiextensions.k8s.io/devicelinks.edge.cattle.io created
role.rbac.authorization.k8s.io/octopus-leader-election-role created
clusterrole.rbac.authorization.k8s.io/octopus-manager-role created
rolebinding.rbac.authorization.k8s.io/octopus-leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/octopus-manager-rolebinding created
service/octopus-brain created
service/octopus-limb created
deployment.apps/octopus-brain created
daemonset.apps/octopus-limb created

```

After installed, we can verify the status of Octopus as below:

```shell script
$ kubectl get all -n octopus-system
NAME                                 READY   STATUS    RESTARTS   AGE
pod/octopus-limb-w8vcf               1/1     Running   0          14s
pod/octopus-limb-862kh               1/1     Running   0          14s
pod/octopus-limb-797d8               1/1     Running   0          14s
pod/octopus-limb-8w462               1/1     Running   0          14s
pod/octopus-brain-65fdb4ff99-zvw62   1/1     Running   0          14s

NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/octopus-brain   ClusterIP   10.43.92.81    <none>        8080/TCP   14s
service/octopus-limb    ClusterIP   10.43.143.49   <none>        8080/TCP   14s

NAME                          DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/octopus-limb   4         4         4       4            4           <none>          14s

NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/octopus-brain   1/1     1            1           14s

NAME                                       DESIRED   CURRENT   READY   AGE
replicaset.apps/octopus-brain-65fdb4ff99   1         1         1       14s

```

### Deploy Device Model & Device Adaptor

Octopus has prepared a dummy adaptor for testing, which does not need to be connected to a real device. So we can imagine that the dummy device is a realistic device in here.

At first, we need to describe the device as a resource in Kubernetes. This description process is modeling the device. In Kubernetes, the best way to describe resources is to use [CustomResourceDefinitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions), so **defining a device model in Octopus is actually defining the CustomResourceDefinitions.** Take a quick look at this `DummySpecialDevice` model(pretend it is a smart fan): 

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.5
    devices.edge.cattle.io/description: dummy device description
    devices.edge.cattle.io/device-property: ""
    devices.edge.cattle.io/enable: "true"
    devices.edge.cattle.io/icon: ""
  labels:
    app.kubernetes.io/name: octopus-adaptor-dummy
    app.kubernetes.io/version: master
  name: dummyspecialdevices.devices.edge.cattle.io
spec:
  group: devices.edge.cattle.io
  names:
    kind: DummySpecialDevice
    listKind: DummySpecialDeviceList
    plural: dummyspecialdevices
    singular: dummyspecialdevice
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        description: DummySpecialDevice is the Schema for the dummy special device
          API.
        properties:
          ...
          spec:
            description: DummySpecialDeviceSpec defines the desired state of DummySpecialDevice.
            properties:
              gear:
                description: Specifies how fast the dummy special device should be.
                enum:
                - slow
                - middle
                - fast
                type: string
              "on":
                description: Turn on the dummy special device.
                type: boolean
              protocol:
                description: Protocol for accessing the dummy special device.
                properties:
                  location:
                    type: string
                required:
                - location
                type: object
            required:
            - "on"
            - protocol
            type: object
          status:
            description: DummySpecialDeviceStatus defines the observed state of DummySpecialDevice.
            properties:
              gear:
                description: Reports the current gear of dummy special device.
                enum:
                - slow
                - middle
                - fast
                type: string
              rotatingSpeed:
                description: Reports the detail number of speed of dummy special device.
                format: int32
                type: integer
            type: object
        type: object
    ...
status:
  ...

```

The dummy adaptor installer YAML file is under the [`adaptors/dummy/deploy/e2e`](./adaptors/dummy/deploy/e2e) directory, the `all_in_one.yaml` includes the device model and the device adaptor, we can apply them into the cluster directly:

```shell script
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/dummy/deploy/e2e/all_in_one.yaml
customresourcedefinition.apiextensions.k8s.io/dummyspecialdevices.devices.edge.cattle.io created
customresourcedefinition.apiextensions.k8s.io/dummyprotocoldevices.devices.edge.cattle.io created
clusterrole.rbac.authorization.k8s.io/octopus-adaptor-dummy-manager-role created
clusterrolebinding.rbac.authorization.k8s.io/octopus-adaptor-dummy-manager-rolebinding created
daemonset.apps/octopus-adaptor-dummy-adaptor created

$ kubectl get all -n octopus-system
NAME                                      READY   STATUS    RESTARTS   AGE
pod/octopus-limb-w8vcf                    1/1     Running   0          2m27s
pod/octopus-limb-862kh                    1/1     Running   0          2m27s
pod/octopus-limb-797d8                    1/1     Running   0          2m27s
pod/octopus-limb-8w462                    1/1     Running   0          2m27s
pod/octopus-brain-65fdb4ff99-zvw62        1/1     Running   0          2m27s
pod/octopus-adaptor-dummy-adaptor-6xcdz   1/1     Running   0          21s
pod/octopus-adaptor-dummy-adaptor-mmk5l   1/1     Running   0          21s
pod/octopus-adaptor-dummy-adaptor-xnjrf   1/1     Running   0          21s
pod/octopus-adaptor-dummy-adaptor-srsjz   1/1     Running   0          21s

NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/octopus-brain   ClusterIP   10.43.92.81    <none>        8080/TCP   2m27s
service/octopus-limb    ClusterIP   10.43.143.49   <none>        8080/TCP   2m27s

NAME                                           DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/octopus-limb                    4         4         4       4            4           <none>          2m27s
daemonset.apps/octopus-adaptor-dummy-adaptor   4         4         4       4            4           <none>          21s

NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/octopus-brain   1/1     1            1           2m27s

NAME                                       DESIRED   CURRENT   READY   AGE
replicaset.apps/octopus-brain-65fdb4ff99   1         1         1       2m27s

```

It is worth noting that we have granted the permission to Octopus for managing `DummySpecialDevice`/`DummyProtocolDevice`:

```shell script
$ kubectl get clusterrolebinding | grep octopus
octopus-manager-rolebinding                            2m49s
octopus-adaptor-dummy-manager-rolebinding              43s

```

### Create DeviceLink

Next, we are going to connect a device via `DeviceLink`. A link mainly consists of 3 fields: `adaptor`, `model` and `template(device spec)`:

- `adaptor` describes how to access the device, this accessing process calls **Adaptation**. In order to adapt a device, we need to indicate the name of the adaptor and the name of the device-connectable node.
- `model` describes the model of device, it is the [TypeMeta](https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go) of the device model CRD.
- `template(device spec)` describes the desired status of device, it is determined by the device model CRD.

In addition, we can also use the `references` field to refer the [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/) and [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) under the same Namespace, even use the downward API to fetch the information in `DeviceLink`.

We can imagine that there is a device named `living-room-fan` on the `edge-worker` node, and then we can connect it by Octopus:

```yaml
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: living-room-fan
  namespace: default
spec:
  adaptor:
    node: edge-worker
    name: adaptors.edge.cattle.io/dummy
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "DummySpecialDevice"
  template:
    metadata:
      labels:
        device: living-room-fan
    spec:
      protocol:
        location: "living_room"
      gear: slow
      "on": true

```

After deployed the above `DeviceLink` into a cluster, we could find that there are [several states](./docs/octopus/state_of_devicelink.md) of a link. If the **DeviceConnected** `PHASE` is on **Healthy** `STATUS`, we can query the same name instance of device model CRD, now the device is in our cluster:

```shell script
$ kubectl get devicelink living-room-fan -n default
NAME              KIND                 NODE          ADAPTOR                         PHASE             STATUS    AGE
living-room-fan   DummySpecialDevice   edge-worker   adaptors.edge.cattle.io/dummy   DeviceConnected   Healthy   10s

$ kubectl get dummyspecialdevice living-room-fan -n default -w
NAME              GEAR   SPEED   AGE
living-room-fan   slow   10      32s
living-room-fan   slow   11      33s
living-room-fan   slow   12      36s

```

### Manage Device

When we want to stop the device, we can do this as below:

```shell script
$ kubectl patch devicelink living-room-fan -n default --type merge --patch '{"spec":{"template":{"spec":{"on":false}}}}'
devicelink.edge.cattle.io/living-room-fan patched

$ kubectl get devicelink living-room-fan -n default
  NAME              KIND                 NODE          ADAPTOR                         PHASE             STATUS    AGE
  living-room-fan   DummySpecialDevice   edge-worker   adaptors.edge.cattle.io/dummy   DeviceConnected   Healthy   89s

$ kubectl get dummyspecialdevice living-room-fan -n default
NAME              GEAR   SPEED   AGE
living-room-fan                  117s

```

## Documentation

<!-- toc -->
- Octopus
    - [How to deploy](./docs/octopus/deployment.md)
    - [How to develop](./docs/octopus/develop.md)
    - [How to monitor](./docs/octopus/monitoring.md)
    - [The state transition of DeviceLink](./docs/octopus/state_of_devicelink.md)
- Adaptors
    - [How it works](./docs/adaptors/design_of_adaptor.md)
    - [How to develop](./docs/adaptors/develop.md)
<!-- /toc -->

## License

Octopus is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.
