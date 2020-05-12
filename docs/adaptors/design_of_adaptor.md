# Design of Adaptor

<!-- toc -->

- [Idea](#idea)
- [Available Adaptor List](#available-adaptor-list)

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

The access management of adaptors takes inspiration from [Kubernetes Device Plugins management](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/). The workflow includes the following steps:

1. The `limb` starts a gRPC service with a Unix socket on host path to receive registration requests from adaptors: <a id="registration"></a>
    ```proto
    // Registration is the service advertised by the Limb,
    // any adaptor start its service until Limb approved this register request.
    service Registration {
        rpc Register (RegisterRequest) returns (Empty) {}
    }
    
    message RegisterRequest {
        // Name of the adaptor in the form `adaptor-vendor.com/adaptor-vendor`.
        string name = 1;
        // Version of the API the adaptor was built against.
        string version = 2;
        // Name of the unix socket the adaptor is listening on, it's in the form `*.socket`.
        string endpoint = 3;
    }
    ```
1. The adaptor starts a gRPC service with a Unix socket under host path `/var/lib/octopus/adaptors`, that implements the following interfaces: <a id="connection"></a>
    ```proto
    // Connection is the service advertised by the adaptor.
    service Connection {
        rpc Connect (stream ConnectRequest) returns (stream ConnectResponse) {}
    }
    
    message ConnectRequest {
        // Parameters for the connection, it's in form JSON bytes.
        bytes parameters = 1;
        // Model for the device.
        k8s.io.apimachinery.pkg.apis.meta.v1.TypeMeta model = 2;
        // Desired device, it's in form JSON bytes.
        bytes device = 3;
    }
    
    message ConnectResponse {
        // Observed device, it's in form JSON bytes.
        bytes device = 1;
    }
    ```
1. The adaptor registers itself with the `limb` through the Unix socket at host path `/var/lib/octopus/adaptors/limb.socket`.
1. After successfully registering itself, the adaptor runs in serving mode, during which it keeps connecting devices and reports back to the `limb` upon any device state changes.

## Available Adaptor List

- [dummy](../../adaptors/dummy)
- [ble](../../adaptors/ble)
- [modbus](../../adaptors/modbus)
- [opcua](../../adaptors/opcua)
- [mqtt](../../adaptors/mqtt)
