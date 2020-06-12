# Test steps

## Deploy octopus

1. Start a k3s cluster

2. Deploy octopus in your k3s cluster
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/deploy/e2e/all_in_one_without_webhook.yaml
```

3. deploy opcua adaptor
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/opcua/deploy/e2e/test/all_in_one.yaml
```

## Run mock server
[Open62541](https://github.com/open62541/open62541) is a open source OPC-UA implementation written in C++. We can run a container using its official mock server [docker image](https://hub.docker.com/r/open62541/open62541).

```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/opcua/deploy/e2e/test/server.yaml
```

## Apply devicelink
Configure the correct parameters and apply the devicelink.
```shell script
kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus/master/adaptors/opcua/deploy/e2e/test_opcua.yaml
```
This `devicelink` has the three properties readable from the mock server.

## Check device status
Make sure the `devicelink` status to be `healthy`. And check `status` from the device `yaml`. 
```shell script
kubectl get dl
kubectl get opcuadevice opcua-open -o yaml
```
We can get the following status if the values are not overwritten in `dl.spec.template.spec`.
```yaml
status:
  properties:
  - dataType: datetime
    name: datetime
    updatedAt: "2020-06-03T07:06:41Z"
    value: 2020-06-03 07:06:41.269882 +0000 UTC
  - dataType: int32
    name: integer
    updatedAt: "2020-06-03T07:06:41Z"
    value: "42"
  - dataType: byteString
    name: string
    updatedAt: "2020-06-03T07:06:41Z"
    value: test123
```