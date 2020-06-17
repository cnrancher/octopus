# Modbus Adaptor

## Introduction

[Modbus](https://www.modbustools.com/modbus.html) is a master/slave protocol, the device requesting the information is called the Modbus master and the devices supplying information are Modbus slaves. 
In a standard Modbus network, there is one master and up to 247 slaves, each with a unique slave address from 1 to 247. 
The master can also write information to the slaves.

Modbus adaptor support both TCP and RTU protocol, it acting as the master node and connects to or manipulating the Modbus slave devices on the edge side.

### Registers Operation

- **Coil Registers**: readable and writable, 1 bit (off/on)

- **Discrete Input Registers**: readable, 1 bit (off/on)

- **Input Registers**: readable, 16 bits (0 to 65,535), essentially measurements and statuses

- **Holding Registers**: readable and writable, 16 bits (0 to 65,535), essentially configuration values


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

## Modbus DeviceLink YAML

example of modbus deviceLink YAML
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

### Modbus Device Spec

Parameter | Description | Scheme | Required
--- | --- | --- | ---
parameters | Parameter of the modbus device| *[ModbusDeviceParamters](#modbusdeviceparamters) | false
protocol | Protocol for accessing the modbus device  | *[ModbusProtocolConfig](#modbusprotocolconfig) | true
properties | Device properties  | []*[DeviceProperty](#deviceproperty) | false
extension | Integrate with deivce MQTT extension  | *[DeviceExtension](#deviceextension) | false

#### ModbusDeviceParamters

Parameter | Description | Scheme | Required
--- | --- | --- | ---
syncInterval | Device properties sync interval, default to `5s`  | string | false
timeout |  Device connection timeout, default to `10s` | string | false

#### ModbusProtocolConfig

Parameter | Description | Scheme | Required
--- | --- | --- | ---
rtu | Modbus RTU protocol config  | *[ModbusConfigRTU](#modbusconfigrtu)| false
tcp | Modbus TCP protocol config  | *[ModbusConfigTCP](#modbusconfigtcp)| false

#### ModbusConfigRTU

Parameter | Description | Scheme | Required
--- | --- | --- | ---
serialPort | Device path (e.g. /dev/ttyS0) | string | true
slaveId | Slave id of the device | int | true
baudRate | Baud rate, a measurement of transmission speed, default to `19200` | int | false
dataBits | Data bits (5, 6, 7 or 8), default to `0` | int | false
parity | N - None, E - Even, O - Odd (default E) (The use of no parity requires 2 stop bits.) | string | false
stopBits | Stop bits: 1 or 2 (default 1) | int | false

#### ModbusConfigTCP

Parameter | Description | Scheme | Required
--- | --- | --- | ---
ip | IP address of the device | string | true
port | TCP port of the device | int | true
slaveId | Slave id of the device | int | true

#### DeviceProperty

Parameter | Description | Scheme | Required
--- | --- | --- | ---
name | Property name | string | true
description | Property description  | string | false
readOnly | Check if the device property is readonly, default to false | boolean | false
dataType | Property data type, options are `int, string, float, boolean` | string | true
visitor | Property visitor config | *[PropertyVisitor](#propertyvisitor) | true
value | Set desired value of the property | string | false

#### PropertyVisitor

Parameter | Description | Scheme |  Required
--- | --- | --- | ---
register | CoilRegister, DiscreteInputRegister, HoldingRegister, or InputRegister | string | true
offset | Offset indicates the starting register number to read/write data | int | true
quantity | Limit number of registers to read/write | int | true
orderOfOperations | The quantity of registers | []*[ModbusOperations](#modbusoperations) | false

#### ModbusOperations

Parameter | Description | Scheme |  Required
--- | --- | --- | ---
operationType | Arithmetic operation type(`Add, Subtract, Multiply, Divide`) | string | false
operationValue | Arithmetic operation value | string | false

#### DeviceExtension

- reference the [example YAML](#modbus-devicelink-yaml) of modbus device for MQTT integration.
- check [Integrate with MQTT Documentation](https://github.com/cnrancher/octopus/blob/master/docs/adaptors/integrate_with_mqtt.md) for more details.
