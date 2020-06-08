# Design of Adaptor

<!-- toc -->

- [Idea](#idea)
    + [APIs](#apis)
        + [Registration](#registration)
        + [Connection](#connection)
- [Available Adaptors](#available-adaptors)

<!-- /toc -->

## Idea

The Octopus has strong scalability, this ability reflects in the design of the device model and the adapter especially.

Since device model can be defined via CRD, the device model can be defined as either a dedicated device such as a fan, LED, etc., or a general protocol device such as a Bluetooth device, Modbus device, and so on:

```text
                                         ┌──────────────────────┐
                                         │   MideaAirPurifier   │
                                         └──────────────────────┘
                                                                 
                                                                 
                                         ┌──────────────────────┐
                                         │ MideaAirConditioning │
                                         └──────────────────────┘
                                                                 
                                                                 
                                         ┌──────────────────────┐
                                         │  XiaoMiAirPurifier   │
                                         └──────────────────────┘
                                                                 
                                                                 
                                         ┌──────────────────────┐
                                         │ XiaoMiWeighingScale  │
                                         └──────────────────────┘
                                                                 
                                                                 
                                         ╔══════════════════════╗
                                         ║      Bluetooth       ║
                                         ╚══════════════════════╝
                                                                 
                                                                 
                                         ╔══════════════════════╗
                                         ║        Modbus        ║
                                         ╚══════════════════════╝
```

At the same time, the implementation of adapter can be connected to a single device or multiple devices:

```text
                                         ┌──────────────────────┐                                           
                              ┌──────────│   MideaAirPurifier   │──────────┐                                
                              │          └──────────────────────┘          │                                
                              │                                            │                                
                              │                                            │                                
                   .          │          ┌──────────────────────┐          │           .                    
                  ( )◀────────┤          │ MideaAirConditioning │──────────┴─────────▶( )                   
                   '          │          └──────────────────────┘                      '                    
   adaptors.smarthome.io/airpurifier                                      adaptors.media.io/smarthome       
                              │                                                                             
                              │          ┌──────────────────────┐                                           
                              └──────────│  XiaoMiAirPurifier   │──────────┐                                
                                         └──────────────────────┘          │                                
                                                                           │                                
                                                                           │                                
                                         ┌──────────────────────┐          │            .                   
                                         │ XiaoMiWeighingScale  │──────────┴──────────▶( )                  
                                         └──────────────────────┘                       '                   
                                                                          adaptors.xiaomi.io/smarthome      
                                                                                                            
                   .                     ╔══════════════════════╗                                           
                  ( )◀═══════════════════║      Bluetooth       ║                                           
                   '                     ╚══════════════════════╝                                           
    adaptors.edge.cattle.io/bluetooth                                                                       
                                                                                                            
                                         ╔══════════════════════╗                       .                   
                                         ║        Modbus        ║═════════════════════▶( )                  
                                         ╚══════════════════════╝                       '                   
                                                                         adaptors.edge.cattle.io/modbus     
```

Please view [here](./develop.md) for more detail about developing an adaptor.

### APIs

The access management of adaptors takes inspiration from [Kubernetes Device Plugins management](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/). The available version of access management APIs is `v1alpha1`.

|  Versions | Available | Current |
|:---:|:---:|:---:|
|  [`v1alpha1`](./design_of_adaptor.md) | * | * |

Following the below steps can allow an adaptor to interact with `limb`:

1. The `limb` starts a gRPC service with a Unix socket on host path to receive registration requests from adaptors:
    ```proto
    // Registration is the service advertised by the Limb,
    // any adaptor start its service until Limb approved this register request.
    service Registration {
        rpc Register (RegisterRequest) returns (Empty) {}
    }
    
    message RegisterRequest {
        // Name of the adaptor in the form `adaptor-vendor.com/adaptor-name`.
        string name = 1;
        // Version of the API the adaptor was built against.
        string version = 2;
        // Name of the unix socket the adaptor is listening on, it's in the form `*.sock`.
        string endpoint = 3;
    }
    ```
1. The adaptor starts a gRPC service with a Unix socket under host path `/var/lib/octopus/adaptors`, that implements the following interfaces:
    ```proto
    // Connection is the service advertised by the adaptor.
    service Connection {
        rpc Connect (stream ConnectRequest) returns (stream ConnectResponse) {}
    }
    
    message ConnectRequestReferenceEntry {
        map<string, bytes> items = 1;
    }
    
    message ConnectRequest {
        // [Deprecated] Parameters for the connection, it's in form JSON bytes.
        bytes parameters = 1;
        // Model for the device.
        k8s.io.apimachinery.pkg.apis.meta.v1.TypeMeta model = 2;
        // Desired device, it's in form JSON bytes.
        bytes device = 3;
        // References for the device, i.e: Secret, ConfigMap and Downward API.
        map<string, ConnectRequestReferenceEntry> references = 4;
    }
    
    message ConnectResponse {
        // Observed device, it's in form JSON bytes.
        bytes device = 1;
    }
    ```
1. The adaptor registers itself with the `limb` through the Unix socket at host path `/var/lib/octopus/adaptors/limb.sock`.
1. After successfully registering itself, the adaptor runs in serving mode, during which it keeps connecting devices and reports back to the `limb` upon any device state changes.

#### Registration

The **Registration** can let the `limb` to know the existence of an adaptor, on this phase, the `limb` acts as a server and the adaptor acts as a client. The adaptor constructs a registration request with its `name`, `version` and accessing `endpoint`, and then request to `limb`. After successful registered, the `limb` will keep watching the adaptor and notify those DeviceLinks related to the registered adaptor.

- The `name` is the name of the adaptor, it's strongly recommended that to use this pattern `adaptor-vendor.com/adaptor-name` to named an adaptor, each adaptor must have one unique `name`.
    > The second adaptor who has the same `name` will overwrite the previous one.
- The `version` is the API version of accessing management, for now, it's fixed in `v1alpha1`.
- The accessing `endpiont` is the name of unix socket, each adaptor must have one unique `endpoint`. 
    > The second adaptor who has the same registered `endpoint` will never register successfully until the previous one quits.

#### Connection

The **Connection** can let the `limb` to connect to an adaptor, on this phase, the adaptor acts as a server and the `limb` acts as a client. The `limb` constructs a connection request with the `parameters`, `model`, `device` and `references`, and then request to the target adaptor.

- The `parameters` is the parameters used for connection, it's in form JSON bytes.
    > This `parameters` field has been **DEPRECATED**, it should define the connection parameter as a part of device model.
- The `model` is the device's model, it's useful to help adaptor to distinguish multiple models, or maintain the compatibility if there are difference versions in one model.
- The `device` is the device's instance, it's in form JSON bytes, which is JSON bytes of a complete `model` instance and contains `spec` and `status` data.
    > The adaptor should select the corresponding deserialized receiving object according to the `model` to receive this field's data.
    > Since the receiving object (device instance) is a legal CRD instance, it strictly conforms to the resource management convention of Kubernetes, so a device can be uniquely identified by Namespace and Name.
- The `references` is the reference sources of the device, it allows the device to use the ConfigMap and Secret under the same Namespace, or downward API of the parent DeviceLink instance.

## Available Adaptors

- [dummy](../../adaptors/dummy)
- [ble](../../adaptors/ble)
- [modbus](../../adaptors/modbus)
- [opcua](../../adaptors/opcua)
- [mqtt](../../adaptors/mqtt)
