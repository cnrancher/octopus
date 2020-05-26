# Agent Device Group Adaptor

Agent Device Group Adaptor is used to manage a group of devices (1…x) which are running all the same containerized applications on devices with the same architecture.

When an agentDeviceGroup is created, nodes can join the group by running shell command generated in `agentdevicegroup.status.command` on the target node. 
The token of the server is stored as secret for the adaptor to read by a `pod` running before the adaptor container starts.

The adaptor add nodes into `agentdevicegroup.status.nodes` when new nodes are labeled with `devices.edge.cattle.io/group` = `dl.metadata.name`, and run daemonSets on these nodes for applications specified in `template.spec.apps`.  

## Registration Information

|  Versions | Register Name | Endpoint Socket | Available |
|:---:|:---:|:---:|:---:|
|  `v1alpha1` | `adaptors.edge.cattle.io/agent` | `agent.socket` | * |

## Support Model

| Kind | Group | Version | Available | 
|:---:|:---:|:---:|:---:|
| `AgentDeviceGroup` | `devices.edge.cattle.io` | `v1alpha1` | * |

## Support Platform

| OS | Arch |
|:---:|:---|
| `linux` | `amd64` |
| `linux` | `arm` |
| `linux` | `arm64` |

## Usage

```shell script
kubectl apply -f ./deploy/e2e/all_in_one.yaml
```

## Authority

Grant permissions to Octopus as below <!-- kubectl describe clusterrole ... -->:

```text
  Resources                                   Non-Resource URLs  Resource Names  Verbs
  ---------                                   -----------------  --------------  -----
  agentdevicegroups.devices.edge.cattle.io         []                 []              [create delete get list patch update watch]
  agentdevicegroups.devices.edge.cattle.io/status  []                 []              [get patch update]
  secrets                                          []                 []              [create get]
  nodes                                            []                 []              [deletecollection list watch]
  endpoints                                        []                 []              [list watch]
  apps.daemonset                                   []                 []              [get watch create update delete patch]
```
## Agent Device Group Parameters

Parameter | Description | Scheme | Required
--- | --- | --- | ---
`apps` | Applications to deploy to devices in this agent device group | [App](#App) | false
`serverURL` | Agent node registration address | string | false
`deleteNodes` | Whether to delete the nodes in this group when delete this AgentDeviceGroup. Default to `false` | boolean | false

### App

Parameter | Description | Scheme | Required
--- | --- | --- | ---
`name` | app name  | string | true
`namespace` |  app namespace  | string | true
`template` |  app template for pod | [PodTemplate](#PodTemplate) | true


### PodTemplate

Parameter | Description | Scheme | Required
--- | --- | --- | ---
`metadata.labels` | labels of the pod  | map[string]string | false
`spec` | pod spec  | PodSpec | false

## Step by Step Guidance

1. Create `devicelink` for `agentdevicegroup`. Specify the group name, apps to run on the node in this group, and the node registration address, if needed.
 You can find the yaml example at the end of this doc.
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/agent-device/deploy/e2e/dl_agent_simple.yaml
kubectl edit dl nginx-devices 
```

2. Check created `agentdevicegroup`, find `status.command`, which is the command to on the target agent node to join the group.
```shell script
kubectl get agentdevicegroup nginx-devices -o yaml
```
```yaml
apiVersion: devices.edge.cattle.io/v1alpha1
kind: AgentDeviceGroup
metadata:
  annotations:
    edge.cattle.io/adaptor-name: adaptors.edge.cattle.io/agent-device
    edge.cattle.io/adaptor-node: edge-worker
  creationTimestamp: "2020-05-13T16:35:41Z"
  generation: 2
  labels:
    device: ip-172-31-31-2
  name: nginx-devices
  namespace: default
  ownerReferences:
  - apiVersion: edge.cattle.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: DeviceLink
    name: nginx-devices
    uid: 334d0cb7-8e33-446e-aec4-31f0f829426b
  resourceVersion: "1001"
  selfLink: /apis/devices.edge.cattle.io/v1alpha1/namespaces/default/agentdevicegroups/agent-device-group
  uid: 7373b6d2-58c2-4c24-9391-dca1ae711dd2
spec:
  apps:
  - name: nginx
    namespace: default
    template:
      metadata:
        labels:
          device: ip-172-31-31-2
      spec:
        containers:
        - image: nginx:1.14.2
          name: nginx
          ports:
          - containerPort: 80
  serverURL: https://example.com:6443
  deleteNodes: false
status:
  apps:
  - daemonSetStatus:
      currentNumberScheduled: 1
      desiredNumberScheduled: 1
      numberAvailable: 1
      numberMisscheduled: 0
      numberReady: 1
      observedGeneration: 1
      updatedNumberScheduled: 1
    name: nginx
    namespace: default
    updatedAt: "2020-05-13T16:39:16Z"
  command: curl -sfL https://get.k3s.io | K3S_URL=https://example.com:6443 K3S_TOKEN=XXX
    sh -s - agent --node-label devices.edge.cattle.io/group=agent-device-group
```
3.  Run the command above on your agent node. 
```shell script
 curl -sfL https://get.k3s.io | K3S_URL=https://example.com:6443 K3S_TOKEN=XXX
    sh -s - agent --node-label devices.edge.cattle.io/group=agent-device-group
```
Wait for node up and check the node label to see `devices.edge.cattle.io/group` set.
```shell script
kubectl get node ip-172-31-19-209 -o yaml
```
```yaml
labels:
  devices.edge.cattle.io/group: nginx-devices
```
4.  Check node in `agentdevicegroup.status.nodes`, and check the apps successfully deployed on the node. You can also check the status of the apps' daemonSets in `agentdevicegroup.status.apps`.
```shell script
kubectl get agent-device-group nginx-devices -o yaml
```
```yaml
status:
  apps:
  - daemonSetStatus:
      currentNumberScheduled: 1
      desiredNumberScheduled: 1
      numberAvailable: 1
      numberMisscheduled: 0
      numberReady: 1
      observedGeneration: 1
      updatedNumberScheduled: 1
    name: nginx
    namespace: default
    updatedAt: "2020-05-25T09:14:33Z"
  command: curl -sfL https://get.k3s.io | K3S_URL=https://example.com:6443 K3S_TOKEN=XXX
               sh -s - agent --node-label devices.edge.cattle.io/group=agent-device-group
  nodes:
  - ip-172-31-19-209
```

5. If we delete the `devicelink`, the daemonSets should be deleted, so as the nodes if `template.spec.deleteNodes` is specified as `true`.

```shell script
kubectl delete dl nginx-devices
```

## Example of Agent Device Group deviceLink YAML
```yaml
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: nginx-devices
spec:
  adaptor:
    node: ip-172-31-24-40
    name: adaptors.edge.cattle.io/agent-device
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "AgentDeviceGroup"
  template:
    metadata:
      labels:
        device: ip-172-31-31-2
    spec:
      deleteNodes: false
      serverURL: https://example.com:6443
      apps:
        - name: nginx
          namespace: default
          template:
            metadata:
              labels:
                device: ip-172-31-31-2
            spec:
              containers:
                - name: nginx
                  image: nginx:1.14.2
                  ports:
                    - containerPort: 80
```
