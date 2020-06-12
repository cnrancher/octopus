# OPCUA Adaptor

## Introduction

OPCUA Adaptor is used for connecting to and manipulating opcua devices on the edge.

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

## DeviceLink CRD
example deviceLink CRD
```yaml
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: opcua
spec:
  adaptor:
    node: edge-worker
    name: adaptors.edge.cattle.io/opcua
    parameters:
      syncInterval: 5
      timeout: 10
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "OPCUADevice"
  template:
    metadata:
      labels:
        device: opcua
    spec:
      protocol:
        url: opc.tcp://wang-2.local:53530/OPCUA/SimulationServer
        username: dadmin
        password: admin
      properties:
        - name: counter
          description: enable data collection of temperature sensor
          readOnly: true
          visitor:
            nodeID: ns=3;s=Counter
          dataType: int32
        - name: random
          description: enable data collection of temperature sensor
          readOnly: true
          visitor:
            nodeID: ns=3;s=Random
          dataType: double
        - name: constant
          description: enable data collection of temperature sensor
          readOnly: false
          visitor:
            nodeID: ns=3;s=Constant
          value: "2.33"
          dataType: float
```

### Protocol Parameters

| Parameter | Description | Type | Default |
|:--|:--|:--|:--|
| url |  Required. The URL for opc server endpoint. | string |
| username | Optional. Username for access opc server. | string |
| password | Optional. Password for access opc server. | string | 
| securityPolicy | Optional. Valid values are "None", "Basic128Rsa15", "Basic256", "Basic256Sha256", "Aes128Sha256RsaOaep", "Aes256Sha256RsaPss". | string | none |
| securityMode | Optional. Valid values are "None", "Sign", and "SignAndEncrypt". |string | none |

<!-- | certificateFile | Optional. File of the certificate to access opc server. |string|  |
     | privateKeyFile | Optional. File of the private key to access opc server. |string|  | 
-->

### Device Property

| Parameter | Description | Type | 
|:--|:--|:--|
| name | Required. The property's name. | string
| description |  Optional. The description of this property. | string
| readonly |  Optional. If it is true, the value is readonly, otherwise readwrite. | boolean
| datatype |  Required. The datatype of this property | [DataType](#DataType)
| visitor |  Required. The visitor configuration of this property | [PropertyVisitor](#PropertyVisitor)
| value |  Required. The value of this property | string

### PropertyVisitor
| Parameter | Description | Type | 
|:--|:--|:--|
| nodeID | Required. The ID of opc-ua node, e.g. "ns=1,i=1005" | string
| browseName |  Optional. The name of opc-ua node | string

### DataType

| Parameter | Description | Type | 
|:--|:--|:--|
| boolean | Property data type is boolean. | string
| int64 | Property data type is int64. | string
| int32 |  Property data type is int32. | string
| int16 |  Property data type is int16. | string
| uint64 | Property data type is uint64. | string
| uint32 |  Property data type is uint32. | string
| uint16 |  Property data type is uint16. | string
| float |  Property data type is float. | string
| double |  Property data type is double. | string
| string |  Property data type is string. | string
| byteString |  Property data type is bytestring. Will be converted to string for display. | string
| datetime |  Property data type is datetime. | string