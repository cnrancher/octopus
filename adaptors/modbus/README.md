# Modbus Adaptor

## Introduction

Modbus Adaptor is used for connecting to and manipulating modbus devices on the edge.
Modbus Adaptor supports TCP and RTU protocol.

## Registration Information

|  Versions | Register Name | Endpoint Socket | Available |
|:---:|:---:|:---:|:---:|
|  `v1alpha1` | `adaptors.edge.cattle.io/modbus` | `modbus.sock` | * |

## Support Model

| Kind | Group | Version | Available | 
|:---:|:---:|:---:|:---:|
| `ModbusDevice` | `devices.edge.cattle.io` | `v1alpha1` | * |

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
  modbusdevices.devices.edge.cattle.io         []                 []              [create delete get list patch update watch]
  modbusdevices.devices.edge.cattle.io/status  []                 []              [get patch update]
```

## Modbus Protocol

Modbus is a master/slave protocol. 
The device requesting the information is called the Modbus Master and the devices supplying information are Modbus Slaves. 
In a standard Modbus network, there is one Master and up to 247 Slaves, each with a unique Slave Address from 1 to 247. 
The Master can also write information to the Slaves.

In Modbus Adaptor, the adaptor as the master connects to modbus slave devices。

## Registers Operation
**Coil Registers**: readable and writable, 1 bit (off/on)

**Discrete Input Registers**: readable, 1 bit (off/on)

**Input Registers**: readable, 16 bits (0 to 65,535), essentially measurements and statuses

**Holding Registers**: readable and writable, 16 bits (0 to 65,535), essentially configuration values

## DeviceLink CRD
example deviceLink CRD
```yaml
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: modbus-tcp
spec:
  adaptor:
    node: edge-worker
    name: adaptors.edge.cattle.io/modbus
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "ModbusDevice"
  template:
    metadata:
      labels:
        device: modbus-tcp
    spec:
      protocol:
        tcp:
          ip: 192.168.1.3
          port: 502
          slaveID: 1
      properties:
        - name: temperature
          description: data collection of temperature sensor
          readOnly: false
          visitor:
            register: HoldingRegister
            offset: 2
            quantity: 8
          value: "33.3"
          dataType: float
        - name: temperature-enable
          description: enable data collection of temperature sensor
          readOnly: false
          visitor:
            register: CoilRegister
            offset: 2
            quantity: 1
          value: "true"
          dataType: boolean

```

### Parameters
#### TCP Config

| Parameter | Description | Type | 
|:--|:--|:--|
| ip | ip address of the device | string
| port | tcp port of the device | int
| slaveId | slave id of the device | int

#### RTU Config

| Parameter | Description | Type | Default |
|:--|:--|:--|:--|
| serialPort | Device path (e.g. /dev/ttyS0) | string |
| slaveId | slave id of the device | int |
| baudRate | baud rate, a measurement of transmission speed | int | 19200 |
| dataBits | data bits (5, 6, 7 or 8) | int | 8  |
| parity | N - None, E - Even, O - Odd (default E) (The use of no parity requires 2 stop bits.) |string | E |
| stopBits | 1 or 2 |int| 1 |

### Property Visitor
| Parameter | Description | Type | 
|:--|:--|:--|
| register | CoilRegister, DiscreteInputRegister, HoldingRegister, or InputRegister | string
| offset | Offset indicates the starting register number to read/write data | int
| quantity | Limit number of registers to read/write | int