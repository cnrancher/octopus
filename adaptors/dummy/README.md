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
- [Support Platform](#support-platform)
- [Usage](#Usage)
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
| protocol |  Protocol for accessing the dummy special device. | [DummySpecialDeviceProtocol](#dummyspecialdeviceprotocol) | true |
| on | Turn on the dummy special device | bool | true |
| gear | Specifies how fast the dummy special device should be. | [DummySpecialDeviceGear](#dummyspecialdevicegear) | false |

#### DummySpecialDeviceStatus

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
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
| protocol | Protocol for accessing the dummy protocol device. | [DummyProtocolDeviceProtocol](#dummyprotocoldeviceprotocol) | true |
| props | Describes the desired properties. | map[string][DummyProtocolDeviceSpecProps](#dummyprotocoldevicespecprops) | false |

#### DummyProtocolDeviceStatus

| Field | Description | Schema | Required |
|:---|:---|:---|:---:|
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
