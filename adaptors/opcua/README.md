# OPC-UA Adaptor

## Introduction

[OPC Unified Architecture](https://opcfoundation.org/about/opc-technologies/opc-ua/) (OPC-UA) is a machine to machine communication protocol for industrial automation developed by the OPC Foundation.

OPC-UA adaptor implements the [gopcua](https://github.com/gopcua/opcua) and focus on communicating with the industrial OPC-UA equipment and systems for data collection and data manipulation on the edge side.

## Registration Information

|  Versions | Register Name | Endpoint Socket | Available |
|:---:|:---:|:---:|:---:|
|  `v1alpha1` | `adaptors.edge.cattle.io/opcua` | `opcua.sock` | * |

## Support Model

| Kind | Group | Version | Available | 
|:---:|:---:|:---:|:---:|
| `OPCUADevice` | `devices.edge.cattle.io` | `v1alpha1` | * |

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

Grant permissions to Octopus as below:

```text
  Resources                                   Non-Resource URLs  Resource Names  Verbs
  ---------                                   -----------------  --------------  -----
  opcuadevices.devices.edge.cattle.io         []                 []              [create delete get list patch update watch]
  opcuadevices.devices.edge.cattle.io/status  []                 []              [get patch update]
```

## OPC-UA DeviceLink YAML
below is an example of OPC-UA deviceLink YAML
```yaml
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: opcua-open
spec:
  adaptor:
    node: edge-worker
    name: adaptors.edge.cattle.io/opcua
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "OPCUADevice"
  template:
    metadata:
      labels:
        device: opcua-open
    spec:
      parameters:
        syncInterval: 10s
        timeout: 10s
      extension:
        mqtt:
          client:
            server: tcp://test.mosquitto.org:1883
          message:
            topic:
              prefix: cattle.io/octopus
              with: nn # namespace/name
      protocol:
        url: opc.tcp://192.168.64.5:30839/
      properties:
        - name: datetime
          description: the current datetime
          readOnly: true
          visitor:
            nodeID: ns=0;i=2258
          dataType: datetime
        - name: integer
          description: mock number. Default value is 42
          readOnly: false
          visitor:
            nodeID: ns=1;s=the.answer
          dataType: int32
          value: "1"
        - name: string
          description: mock byte string. Default value is "test123"
          readOnly: false
          visitor:
            nodeID: ns=1;s=myByteString
          dataType: byteString
          value: "newString"
```

### OPC-UA Device Spec

Parameter | Description | Scheme | Required
--- | --- | --- | ---
parameters | Parameter of the opcua device| *[DeviceParamters](#deviceparamters) | false
protocol | Protocol for accessing the opcua device  | *[ProtocolConfig](#protocolconfig) | true
properties | Device properties  | []*[DeviceProperty](#deviceproperty) | false
extension | Integrate with deivce MQTT extension  | *[DeviceExtension](#deviceextension) | false

#### DeviceParamters

Parameter | Description | Scheme | Required
--- | --- | --- | ---
syncInterval | Device properties sync interval, default to `5s`  | string | false
timeout |  Device connection timeout, default to `10s` | string | false

#### ProtocolConfig

Parameter | Description | Scheme | Required
--- | --- | --- | ---
url | The URL for opc-ua server endpoint | string | true
username | Username for accessing opc-ua server | string | false
password | Password for opc-ua server endpoint | string | false
securityPolicy | Defaults to `None`. Valid values are `None, Basic128Rsa15, Basic256, Basic256Sha256, Aes128Sha256RsaOaep, Aes256Sha256RsaPss`. | string | false
securityMode | Defaults to `None`. Valid values are `None, Sign, and SignAndEncrypt`. | string | false
certificateFile | Certificate file for accessing opc-ua server | string | true
privateKeyFile | PrivateKey file for accessing opc-ua server | string | true

#### DeviceProperty

Parameter | Description | Scheme | Required
--- | --- | --- | ---
name | Property name | string | true
description | Property description  | string | false
readOnly | Check if the device property is readonly, otherwise readwrite, default to false | boolean | false
dataType | The datatype of this property | *[PropertyDataType](#propertydatatype) | true
visitor | The visitor configuration of this property | *[PropertyVisitor](#propertyvisitor) | true
value | Set desired value of the property | string | false

#### PropertyVisitor

Parameter | Description | Scheme |  Required
--- | --- | --- | ---
nodeID | The ID of opc-ua node, e.g. "ns=1,i=1005" | string | true
browseName | The name of opc-ua node | string | false


#### PropertyDataType

Parameter | Description | Scheme
--- | --- | --- 
boolean | Property data type is boolean. | string
int64 | Property data type is int64. | string
int32 |  Property data type is int32. | string
int16 |  Property data type is int16. | string
uint64 | Property data type is uint64. | string
uint32 |  Property data type is uint32. | string
uint16 |  Property data type is uint16. | string
float |  Property data type is float. | string
double |  Property data type is double. | string
string |  Property data type is string. | string
byteString |  Property data type is bytestring. Will be converted to string for display. | string
datetime |  Property data type is datetime. | string


#### DeviceExtension

- reference the [example YAML](#opc-ua-devicelink-yaml) of opcua device for MQTT integration.
- check [Integrate with MQTT Documentation](https://github.com/cnrancher/octopus/blob/master/docs/adaptors/integrate_with_mqtt.md) for more details.
