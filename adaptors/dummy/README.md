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
- [Support Platform](#support-platform)
- [Usage](#Usage)
- [Authority](#authority)

<!-- /toc -->

## Registration Information

|  Versions | Register Name | Endpoint Socket | Available |
|:---:|:---:|:---:|:---:|
|  `v1alpha1` | `adaptors.edge.cattle.io/dummy` | `dummy.socket` | * |

## Support Model

| Kind | Group | Version | Available | 
|:---:|:---:|:---:|:---:|
| [`DummySpecialDevice`](#dummyspecialdevice) | `devices.edge.cattle.io` | `v1alpha1` | * |

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
  dummyspecialdevices.devices.edge.cattle.io          []                 []              [create delete get list patch update watch]
  dummyspecialdevices.devices.edge.cattle.io/status   []                 []              [get patch update]
```

Permissions obtained from cluster as below: 

```text
none
```
