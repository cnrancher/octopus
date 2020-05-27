#Quick Start

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
5.start the testdevice roomlight in the testdata/testdevice/roomlight directory
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