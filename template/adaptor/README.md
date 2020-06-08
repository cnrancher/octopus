# Template Adaptor

<!-- describe what the adaptor does in here -->

## Registration Information

|  Versions | Register Name | Endpoint Socket | Available |
|:---:|:---:|:---:|:---:|
|  `v1alpha1` | `adaptors.edge.cattle.io/template` | `template.sock` | * |

## Support Model

| Kind | Group | Version | Available | 
|:---:|:---:|:---:|:---:|
| `TemplateDevice` | `devices.edge.cattle.io` | `v1alpha1` | * |

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

Grant permissions to Octopus as below <!-- kubectl describe clusterrole ... -->:

```text
  Resources                                   Non-Resource URLs  Resource Names  Verbs
  ---------                                   -----------------  --------------  -----
  templatedevices.devices.edge.cattle.io         []                 []              [create delete get list patch update watch]
  templatedevices.devices.edge.cattle.io/status  []                 []              [get patch update]
```

Permissions obtained from cluster as below: 

```text
none
```
