# BLE Adaptor

This is BLE(Bluetooth Low Energy) adaptor is used for bluetooth devices connection.

## Registration Information

|  Versions | Register Name | Endpoint Socket | Available |
|:---:|:---:|:---:|:---:|
|  `v1alpha1` | `adaptors.edge.cattle.io/ble` | `ble.socket` | * |

## Support Model

| Kind | Group | Version | Available | 
|:---:|:---:|:---:|:---:|
| `BluetoothDevice` | `devices.edge.cattle.io` | `v1alpha1` | * |

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
  bluetoothdevices.devices.edge.cattle.io         []                 []              [create delete get list patch update watch]
  bluetoothdevices.devices.edge.cattle.io/status  []                 []              [get patch update]
```

## BLE Device Parameters

Parameter | Description | Scheme | Required
--- | --- | --- | ---
name | Device name  | string | either device name or macAddress is required
macAddress |  Device access mac address  | string | either name or macAddress is required
properties | Device properties  | []*[DeviceProperty](#deviceproperty) | false


### DeviceProperty

Parameter | Description | Scheme | Required
--- | --- | --- | ---
name | Property name  | string | true
description |  Property description  | string | false
accessMode | Property accessMode  | *[PropertyAccessMode](#propertyaccessmode) | true
visitor | Property visitor | *[PropertyVisitor](#propertyvisitor) | true

### PropertyAccessMode
Parameter | Description | Scheme | Required
--- | --- | --- | ---
ReadOnly   | Property access mode is read only  | string | false
ReadWrite  | Property access mode is read and write  | string | false
NotifyOnly | Property access mode is notify only  | string | false

### PropertyVisitor
Parameter | Description | Scheme | Required
--- | --- | --- | ---
characteristicUUID | Property UUID  | string | true
defaultValue | Config data write to the bluetooth device(set when access mode is `ReadWrite`), for example `ON` configed in the dataWrite  | string | false
dataWrite | Responsible for converting the data from the string into []byte that is understood by the bluetooth device, for example: `"ON":[1], "OFF":[0]` | string | false
dataConverter | Responsible for converting the data being read from the bluetooth device into string | *[BluetoothDataConverter](#bluetoothdataconverter) | false

### BluetoothDataConverter
Parameter | Description | Scheme | Required
--- | --- | --- | ---
startIndex | Specifies the start index of the incoming byte stream to be converted  | int | true
endIndex | Specifies the end index of incoming byte stream to be converted | int | true
shiftLeft | Specifies the number of bits to shift left | int | false
shiftRight | Specifies the number of bits to shift right | int | false
orderOfOperations | Specifies in what order the operations | []*[BluetoothOperations](#BluetoothOperations) | false

### BluetoothOperations
Parameter | Description | Scheme | Required
--- | --- | --- | ---
operationType | Specifies the operation to be performed | *[BluetoothArithmeticOperationType](#bluetootharithmeticoperationtype) | true
operationValue | Specifies with what value the operation is to be performed | string | true

### BluetoothArithmeticOperationType
Parameter | Description | Scheme | Required
--- | --- | --- | ---
Add | Arithmetic operation of add | string | false
Subtract | Arithmetic operation of subtract | string | false
Multiply | Arithmetic operation of multiply | string | false
Divide | Arithmetic operation of divide | string | false

## Example of BLE deviceLink YAML
```YAML
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: xiaomi-temp-rs2200
spec:
  adaptor:
    node: ubuntu
    name: adaptors.edge.cattle.io/ble
    parameters:
      syncInterval: 30
      timeout: 60
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "BluetoothDevice"
  template:
    metadata:
      labels:
        device: xiaomi-temp-rs2200
    spec:
      name: "MJ_HT_V1"
      # macAddress: ""
      properties:
      - name: data
        description: XiaoMi temp sensor with temperature and humidity data
        accessMode: NotifyOnly
        visitor:
          characteristicUUID: 226c000064764566756266734470666d
```
For more BLE deviceLink examples, please refer to the [deploy/e2e](./deploy/e2e/) directory.
