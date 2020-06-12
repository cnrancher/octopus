# Deploy Modbus RTU device

## Deploy the modbusdevice model and run the Modbus Adaptor
```shell script
# deploy modbus adaptor and modbusdevice model
kubectl apply -f adaptors/modbus/deploy/e2e/all_in_one.yaml

# confirm the modbus adaptor deployment
kubectl get daemonset octopus-adaptor-modbus-adaptor -n octopus-system
```

## Connect Modbus device to edge node
Please ensure that Modbus device is connecting to your edge node. 
If the device is not accessible to any node of the remote cluster, you can create a virtual machine on your local PC and join the cluster.

### Create a virtual machine and mount the device from host PC
For example, we can use [VirtualBox](https://www.virtualbox.org/wiki/Downloads) to create a virtual machine and join the cluster as a worker. 
With the device connected to the local PC, we enable the serial port/USB as applicable on the virtual machine.

## Configure the deviceLink with the serial port of the edge node
Find the mounted serial port of the device on the edge node from `/dev` directory. 
Configure the path to the deviceLink's `spec.template.spec.protocol.rtu.serialPort` parameter. Remember to configure the correct edge node. 
```yaml
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: modbus-rtu
spec:
  adaptor:
    node: test  #node name
    name: adaptors.edge.cattle.io/modbus
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "ModbusDevice"
  template:
    metadata:
      labels:
        device: modbus-rtu
    spec:
      parameters:
        syncInterval: 5
        timeout: 10
      protocol:
        rtu:
          serialPort: /dev/ttyUSB0  #serial port
          slaveID: 1
          parity: "N"
          stopBits: 1
          dataBits: 8
          baudRate: 9600
      properties:
        - name: temperature
          description: data collection of temperature sensor
          readOnly: true
          visitor:
            register: HoldingRegister
            offset: 0
            quantity: 1
            orderOfOperations:
              - operationType: Divide
                operationValue: "10"
          dataType: float
```

 ## Deploy the deviceLink
```shell script
# deploy a devicelink
kubectl apply -f adaptors/modbus/deploy/e2e/dl.yaml
 
# confirm the state of devicelink
kubectl get dl modbus-rtu -n default

# watch the device instance
kubectl get modbusdevice modbus-rtu -n default -w
```