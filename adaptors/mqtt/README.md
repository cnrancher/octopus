# mqtt Adaptor

<!-- describe what the adaptor does in here -->

## Registration Information

|  Versions | Register Name | Endpoint Socket | Available |
|:---:|:---:|:---:|:---:|
|  `v1alpha1` | `adaptors.edge.cattle.io/mqtt` | `mqtt.sock` | * |

## Support Model

| Kind | Group | Version | Available | 
|:---:|:---:|:---:|:---:|
| `mqttDevice` | `devices.edge.cattle.io` | `v1alpha1` | * |

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
  mqttdevices.devices.edge.cattle.io         []                 []              [create delete get list patch update watch]
  mqttdevices.devices.edge.cattle.io/status  []                 []              [get patch update]
```

### DeviceMqttConfig

Parameter | Description | Scheme | Required
--- | --- | --- | ---
broker | MQTT broker url string | string  | true
username | MQTT username | string | true
password | MQTT user password | string | true

### SubInfo

Parameter | Description | Scheme | Required
--- | --- | --- | ---
topic | topic name  | string | true
payloadType |  MQTT payload type (json)  | string | true
qos | qos of MQTT | int | true

### PubInfo

Parameter | Description | Scheme | Required
--- | --- | --- | ---
topic | topic name  | string | true
qos | qos of MQTT | int | true

### ValueProps

Parameter | Description | Scheme | Required
--- | --- | --- | ---
valueType | Reports the type of property | string | false
intValue | Reports the value of int type | int | false
stringValue | Reports the value of string type | string | false
floatValue | Reports the value of float type | float | false
booleanValue | Reports the value of boolean type | bool | false
arrayValue | Reports the value of array type | RawExtension | false
objectValue | Reports the value of object type | RawExtension | false

### DeviceSpecProperty

Parameter | Description | Scheme | Required
--- | --- | --- | ---
name | Property name  | string | true
description |  Property description  | string | false
jsonPath | jsonpath of value ,for more complete information please check out [GJSON Syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md).| string | true
subInfo | subecribe info adaptor | [SubInfo](#SubInfo) | true
value | valueProps of property | [ValueProps](#ValueProps) | false

### DevicePropertyStatus

Parameter | Description | Scheme | Required
--- | --- | --- | ---
name | property name | string | true
description | property describe | string | false
value | valueProps of property | [ValueProps](#ValueProps) | true
updateAt | property status update time | string | true

## Example of MQTT deviceLink YAML
```YAML
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: mqtt-test
spec:
  adaptor:
    node: k3d-k3s-default-server
    name: adaptors.edge.cattle.io/mqtt
    parameters:
      syncInterval: 5
      timeout: 10
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "MqttDevice"
  template:
    metadata:
      labels:
        device: mqtt-test
    spec:
      config:
        broker: "tcp://192.168.8.246:1883"
        password: parchk123
        username: parchk
      properties:
        - name: "switch"
          description: "the room light switch"
          jsonPath: "switch"
          subInfo:
              topic: "device/room/light"
              payloadType: "json"
              qos: 2
        - name: "brightness"
          description: "the room light brightness"
          jsonPath: "brightness"
          subInfo:
              topic: "device/room/light"
              payloadType: "json"
              qos: 2
        - name: "power"
          description: "the room light power"
          jsonPath: "power"
          subInfo:
              topic: "device/room/light"
              payloadType: "json"
              qos: 2


```
For more MQTT deviceLink examples, please refer to the [deploy/e2e](./deploy/e2e/) directory.

### JSON Path Syntax

for more complete information please check out [GJSON Syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md).

### Quick Start

1.Install and run k3s (Can use [k3d](https://github.com/rancher/k3d) to get your cluster up and running quickly)
```shell script
k3d create
```
2.deploy octopus in your k3s cluster use [all_in_one_without_webhook.yaml](../../../../deploy/e2e)
```shell script
kubectl apply -f all_in_one_without_webhook.yaml
```
3.deploy MQTT adaptor use [all_in_one.yaml](../../deploy/e2e)
```shell script
kubectl apply -f all_in_one.yaml
```
4.Change the MQTT setting in the [roomlightcase1.yaml](../../deploy/e2e) file to your own MQTT broker
```yaml
    spec:
      config:
        broker: "tcp://192.168.8.246:1883"
        password: parchk123
        username: parchk
```
5.start the testdevice roomlight in the test/testdata/testdevice/roomlight directory
```shell script
cd ./testdata/testdevice/roomlight
go build
./roomlight -b "tcp://192.168.8.246:1883"
```
6.deploy the DeviceLink use [roomlightcase1.yaml](../../deploy/e2e)
```shell script
kubeclt apply -f roomlightcase1.yaml
```
7.check the resource status of devices in the clusters
```shell script
kubeclt get MqttDevice mqtt-test -oyaml
```

if all right you will get the resource info like this:
```yaml
apiVersion: "devices.edge.cattle.io/v1alpha1"
kind: "MqttDevice"
metadata: 
  creationTimestamp: 
  name: "testDevice"
spec: 
  config: 
    broker: ""
    password: ""
    username: ""
  properties: 
  - description: "test property"
    jsonPath: "power"
    name: "test_property"
    pubInfo: 
      qos: "0"
      topic: ""
    subInfo: 
      payloadType: "json"
      qos: "2"
      topic: "test/abc"
    value: 
      valueType: ""
status: 
  properties: 
  - description: "test property"
    name: "test_property"
    updateAt: "2020-05-20T09:04:46Z"
    value: 
      objectValue: 
        electricQuantity: "19.99"
        powerDissipation: "10KWH"
      valueType: "object"
```
For example, if you want to modify the value of a device's attributes, you can modify DeviceLinke's attributes,use cmd
```
kubectl edit dl mqtt-test
```
and add spec property value type like this:
```yaml
    spec:
      config:
        broker: tcp://192.168.8.246:1883
        password: parchk123
        username: parchk
      properties:
      - description: the room light switch
        jsonPath: switch
        name: switch
        subInfo:
          payloadType: json
          qos: 2
          topic: device/room/light
        value:
          stringValue: "on"
          valueType: string

```

