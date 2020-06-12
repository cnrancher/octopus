# Test steps

## Deploy octopus

1. Start a k3s cluster

2. Deploy octopus in your k3s cluster
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/deploy/e2e/all_in_one_without_webhook.yaml
```

3. deploy modbus adaptor
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/modbus/deploy/e2e/all_in_one.yaml
```

## Run mock server

This server is a mock server. Holding register 0 and 1 are random integers between 0 and 400, updating every 5 seconds.
Holding register 5 is the alert limitation of the two values, to be set by the user. Coil register 0 is the alert.

### Deploy the Modbus TCP server
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/modbus-server/master/deploy/modbus-tcp.yaml
```

### Deploy the Modbus RTU server
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/modbus-server/master/deploy/modbus-rtu.yaml
```

## Apply devicelink
Change the corresponding node and protocol config, and apply the `devicelink`.
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/modbus-server/master/deploy/thermometer-tcp.yaml
```
or RTU
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/modbus-server/master/deploy/thermometer-rtu.yaml
```
This is a mock thermometer. Holding register 0 is considered to be temperature, while holding register 1 is considered to be humidity, and the values are divided by 10. 

## Check device status
```shell script
kubectl get modbusdevice thermometer-tcp -o yaml
```
or
```shell script
kubectl get modbusdevice thermometer-rtu -o yaml
```
```yaml
status:
  properties:
  - dataType: float
    name: temperature
    updatedAt: "2020-06-01T02:43:43Z"
    value: "14"
  - dataType: float
    name: humidity
    updatedAt: "2020-06-01T02:43:43Z"
    value: "18.6"
  - dataType: boolean
    name: alert
    updatedAt: "2020-06-01T02:43:43Z"
    value: "false"
  - dataType: float
    name: limitation
    updatedAt: "2020-06-01T02:43:43Z"
    value: "20"
```