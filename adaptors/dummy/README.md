# Dummy Adaptor

This is for experience or testing.

<!-- toc -->

- [Registration Information](#registration-information)
- [Support Model](#support-model)
    + [DummySpecialDevice](#dummyspecialdevice)
        + [DummySpecialDeviceSpec](#dummyspecialdevicespec)
        + [DummySpecialDeviceStatus](#dummyspecialdevicestatus)
        + [DummySpecialDeviceProtocol](#dummyspecialdeviceprotocol)
        + [DummySpecialDeviceGear](#dummyspecialdevicegear)
    + [DummyProtocolDevice](#dummyprotocoldevice)
        + [DummyProtocolDeviceSpec](#dummyprotocoldevicespec)
        + [DummyProtocolDeviceStatus](#dummyprotocoldevicestatus)
        + [DummyProtocolDeviceProtocol](#dummyprotocoldeviceprotocol)
        + [DummyProtocolDeviceSpecProps](#dummyprotocoldevicespecprops)
        + [DummyProtocolDevicesStatusProps](#dummyprotocoldevicestatusprops)
        + [DummyProtocolDevicePropertyType](#dummyprotocoldevicepropertytype)
    + [DeviceExtensionSpec](#deviceextensionspec)
    + [DeviceExtensionStatus](#deviceextensionstatus)
- [Support Platform](#support-platform)
- [Usage](#usage)
    + [Demo](#demo)
- [Authority](#authority)

<!-- /toc -->

## Registration Information

|  Versions | Register Name | Endpoint Socket | Available |
|:---:|:---:|:---:|:---:|
|  `v1alpha1` | `adaptors.edge.cattle.io/dummy` | `dummy.sock` | * |

## Support Model

| Kind | Group | Version | Available | 
|:---:|:---:|:---:|:---:|
| [`DummySpecialDevice`](#dummyspecialdevice) | `devices.edge.cattle.io` | `v1alpha1` | * |
| [`DummyProtocolDevice`](#dummyprotocoldevice) | `devices.edge.cattle.io` | `v1alpha1` | * |

### DummySpecialDevice

The `DummySpecialDevice` can be considered as a fake fan.

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| metadata | | [metav1.ObjectMeta](https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L110) | false |
| spec | Defines the desired state of DummySpecialDevice. | [DummySpecialDeviceSpec](#dummyspecialdevicespec) | true |
| status | Defines the observed state of DummySpecialDevice. | [DummySpecialDeviceStatus](#dummyspecialdevicestatus) | false |

#### DummySpecialDeviceSpec

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| extension | Specifies the extension of device. | [DeviceExtensionSpec](#deviceextensionspec) | false |
| protocol |  Protocol for accessing the dummy special device. | [DummySpecialDeviceProtocol](#dummyspecialdeviceprotocol) | true |
| on | Turn on the dummy special device | bool | true |
| gear | Specifies how fast the dummy special device should be. | [DummySpecialDeviceGear](#dummyspecialdevicegear) | false |

#### DummySpecialDeviceStatus

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| extension | Reports the extension of device. | [DeviceExtensionStatus](#deviceextensionstatus) | false |
| gear | Reports the current gear of dummy special device. | [DummySpecialDeviceGear](#dummyspecialdevicegear) | false |
| rotatingSpeed | Reports the detail number of speed of dummy special device. | int32 | false |

#### DummySpecialDeviceProtocol

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| location | Specifies where to locate the dummy special device. | string | true |

#### DummySpecialDeviceGear

DummySpecialDeviceGear defines how fast the dummy special device should be.

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| slow | Starts from 0 and increases every three seconds until 100. | string | false |
| middle | Starts from 100 and increases every two seconds until 200. | string | false |
| fast | Starts from 200 and increases every one second until 300. | string | false |

### DummyProtocolDevice

The `DummyProtocolDevice` can be considered as a chaos protocol robot, it will change its attribute values every two seconds.

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| metadata | | [metav1.ObjectMeta](https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L110) | false |
| spec | Defines the desired state of DummyProtocolDevice. | [DummyProtocolDeviceSpec](#dummyprotocoldevicespec) | true |
| status | Defines the observed state of DummyProtocolDevice. | [DummyProtocolDeviceStatus](#dummyprotocoldevicestatus) | false |

#### DummyProtocolDeviceSpec

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| extension | Specifies the extension of device. | [DeviceExtensionSpec](#deviceextensionspec) | false |
| protocol | Protocol for accessing the dummy protocol device. | [DummyProtocolDeviceProtocol](#dummyprotocoldeviceprotocol) | true |
| props | Describes the desired properties. | map[string][DummyProtocolDeviceSpecProps](#dummyprotocoldevicespecprops) | false |

#### DummyProtocolDeviceStatus

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| extension | Reports the extension of device. | [DeviceExtensionStatus](#deviceextensionstatus) | false |
| props | Reports the observed value of the desired properties. | map[string][DummyProtocolDeviceStatusProps](#dummyprotocoldevicestatusprops) | false |

#### DummyProtocolDeviceProtocol

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| ip | Specifies where to connect the dummy protocol device. | string | true |

#### DummyProtocolDeviceSpecProps

> `DummyProtocolDeviceSpecObjectOrArrayProps` is the same as `DummyProtocolDeviceSpecProps`.
> The existence of `DummyProtocolDeviceSpecObjectOrArrayProps` is to combat the object circular reference.

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| type | Describes the type of property. | [DummyProtocolDevicePropertyType](#dummyprotocoldevicepropertytype) | true |
| description | Outlines the property. | string | false |
| readOnly | Configures the property is readOnly or not. | bool | false |
| arrayProps | Describes item properties of the array type. | *[DummyProtocolDeviceSpecObjectOrArrayProps](#dummyprotocoldevicespecprops) | false | 
| objectProps | Describes properties of the object type. | map[string][DummyProtocolDeviceSpecObjectOrArrayProps](#dummyprotocoldevicespecprops) | false |

#### DummyProtocolDeviceStatusProps

> `DummyProtocolDeviceStatusObjectOrArrayProps` is the same as `DummyProtocolDeviceStatusProps`.
> The existence of `DummyProtocolDeviceStatusObjectOrArrayProps` is to combat the object circular reference.

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| type | Reports the type of property. | [DummyProtocolDevicePropertyType](#dummyprotocoldevicepropertytype) | true |
| intValue | Reports the value of int type. | *int | false |
| stringValue | Reports the value of string type. | *string | false |
| floatValue | Reports the value of float type. | *[resource.Quantity](https://github.com/kubernetes/apimachinery/blob/master/pkg/api/resource/quantity.go) [kubernetes-sigs/controller-tools/issues#245](https://github.com/kubernetes-sigs/controller-tools/issues/245#issuecomment-550030238) | false |
| booleanValue | Reports the value of bool type. | *bool | false |
| arrayValue | Reports the value of array type. | [][DummyProtocolDeviceStatusObjectOrArrayProps](#dummyprotocoldevicestatusprops) | false | 
| objectValue | Reports the value of object type. | map[string][DummyProtocolDeviceStatusObjectOrArrayProps](#dummyprotocoldevicestatusprops) | false |

#### DummyProtocolDevicePropertyType

DummyProtocolDevicePropertyType describes the type of property.

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| string | | string | false |
| int | | string | false |
| float | | string | false |
| boolean | | string | false |
| array | | string | false |
| object | | string | false |

### DeviceExtensionSpec

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| mqtt | Specifies the MQTT settings. | *[v1alpha1.MQTTOptionsSpec](../../docs/adaptors/integrate_with_mqtt.md#specification) | true |

### DeviceExtensionStatus

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
| mqtt | Reports the MQTT settings. | *[v1alpha1.MQTTOptionsStatus](../../docs/adaptors/integrate_with_mqtt.md#status) | true |

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

### Demo

1. Create a [DeviceLink](./deploy/e2e/dl_specialdevice.yaml) to connect the DummySpecialDevice, which simulates a fan of living room. 

    ```shell script
    kubectl apply -f ./deploy/e2e/dl_specialdevice.yaml
    ```
    
    Synchronize the above created fan's status to remote MQTT broker server.
    
    ```shell script
    # create a Generic Secret to store the CA for connecting test.mosquitto.org.
    kubectl create secret generic living-room-fan-mqtt-ca --from-file=ca.crt=./test/integration/physical/testdata/mosquitto.org.crt
    # create a TLS Secret to store the TLS/SSL keypair for connecting test.mosquitto.org.
    kubectl create secret tls living-room-fan-mqtt-tls --key ./test/integration/physical/testdata/client-key.pem --cert ./test/integration/physical/testdata/client.crt
    # publish status to test.mosquitto.org
    kubectl apply -f ./deploy/e2e/dl_specialdevice_with_mqtt.yaml
    ```
    
    Use [`mosquitto_sub`](https://mosquitto.org/man/mosquitto_sub-1.html) tool to watch the synchronized status.
    
    ```shell script
    # get mqtt broker server
    kubectl get dummyspecialdevices.devices.edge.cattle.io living-room-fan -o jsonpath="{.status.extension.mqtt.client.server}"
    # get topic name
    kubectl get dummyspecialdevices.devices.edge.cattle.io living-room-fan -o jsonpath="{.status.extension.mqtt.message.topicName}"
    # use mosquitto_sub
    mosquitto_sub -h {the host of mqtt broker server} -p {the port of mqtt broker server} -t {the topic name}
    # mosquitto_sub -h test.mosquitto.org -p 1883 -t cattle.io/octopus/default/living-room-fan 
    ```
   
1. Create a [DeviceLink](./deploy/e2e/dl_protocoldevice.yaml) to connect the DummyProtocolDevice, which simulates an intelligent property-filled robot, it can fill the desired properties randomly in 2 seconds.

    ```shell script
    kubectl apply -f ./deploy/e2e/dl_protocoldevice.yaml
    ```
   
    Synchronize the above created robot's answers to remote MQTT broker server.
        
    ```shell script
    # publish status to test.mosquitto.org
    kubectl apply -f ./deploy/e2e/dl_protocoldevice_with_mqtt.yaml
    ```
    
    Use [`mosquitto_sub`](https://mosquitto.org/man/mosquitto_sub-1.html) tool to watch the synchronized answers.
    
    ```shell script
    # get mqtt broker server
    kubectl get dummyprotocoldevices.devices.edge.cattle.io localhost-robot -o jsonpath="{.status.extension.mqtt.client.server}"
    # get topic name
    kubectl get dummyprotocoldevices.devices.edge.cattle.io localhost-robot -o jsonpath="{.status.extension.mqtt.message.topicName}"
    # use mosquitto_sub
    mosquitto_sub -h {the host of mqtt broker server} -p {the port of mqtt broker server} -t {the topic name}
    # mosquitto_sub -h test.mosquitto.org -p 1883 -t cattle.io/octopus/835aea2e-5f80-4d14-88f5-40c4bda41aa3
    ```

## Authority

Grant permissions to Octopus as below:

```text
  Resources                                           Non-Resource URLs  Resource Names  Verbs
  ---------                                           -----------------  --------------  -----
  dummyprotocoldevices.devices.edge.cattle.io         []                 []              [create delete get list patch update watch]
  dummyspecialdevices.devices.edge.cattle.io          []                 []              [create delete get list patch update watch]
  dummyprotocoldevices.devices.edge.cattle.io/status  []                 []              [get patch update]
  dummyspecialdevices.devices.edge.cattle.io/status   []                 []              [get patch update]
```

Permissions obtained from cluster as below: 

```text
none
```
