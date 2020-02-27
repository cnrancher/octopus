# State of Device Link

<!-- toc -->

- [Main Flow](#main-flow)
- [Correct behaviors](#correct-behaviors)
  - [State of DeviceConnected ](#state-of-deviceconnected )

<!-- /toc -->

## Main Flow

In the status of `DeviceLink`, there are several conditions used for tracing the states of the link. Next, let us come to understand the transition of these conditions.

There is a `DeviceLink` description as below:

```yaml
apiVersion: edge.cattle.io/v1alpha1
kind: DeviceLink
metadata:
  name: example
  namespace: octopus-test
spec:
  adaptor:
    node: edge-worker
    name: adaptors.edge.cattle.io/dummy
    parameters:
      ip: 192.168.2.47
  model:
    apiVersion: "devices.edge.cattle.io/v1alpha1"
    kind: "DummyDevice"
  template:
    metadata:
      labels:
        device: example
    spec:
      gear: slow
      "on": true
```

Suppose we create the above `DeviceLink` into the cluster, the `brain` will detect if the Node corresponding to `spec.adaptor.node` is available:

```text
┌─────────────────────┐   if the node is available? 
│     NodeExisted     │───────────────┐             
└─────────────────────┘               │             
                                      ▼             
                                      .             
                                     ( ) brain      
                                      '             
```

If the Node is available, the `brain` will detect if the CRD corresponding to `spec.model` is available:

```text
┌─────────────────────┐   if the node is available? 
│     NodeExisted     │───────────────┐             
└─────────────────────┘               │             
                                      ▼             
                          yes         .             
           ┌─────────────────────────( ) brain      
           │                          '             
           ▼                                        
┌─────────────────────┐   if the model is available?  
│    ModelExisted     │───────────────┐             
└─────────────────────┘               │             
                                      ▼             
                                      .             
                                     ( ) brain      
                                      '             
```

*It is worth noting that the CRD used as a device model requires an annotation: `devices.edge.cattle.io/enable:true`.* If the CRD is available, the `limb` will detect if the adaptor corresponding to `spec.adaptor.name` is available:

```text
           │                                        
           ▼                                           
┌─────────────────────┐   if the model is available?   
│    ModelExisted     │───────────────┐                
└─────────────────────┘               │                
                                      ▼                
                          yes         .                
           ┌─────────────────────────( ) brain         
           │                          '                
           ▼                                           
┌─────────────────────┐   if the adaptor is available? 
│   AdaptorExisted    │───────────────┐                
└─────────────────────┘               │                
                                      ▼                
                                      .                
                                     ( ) limb          
                                      '                
```

You can view [Design of Adaptor](../adaptors/design_of_adaptor.md) to learn how `limb` detects an adaptor. If the adaptor is available, the `limb` will try to create a device instance related with `spec.model`:

```text
           │                                          
           ▼                                           
┌─────────────────────┐   if the adaptor is available? 
│   AdaptorExisted    │───────────────┐                
└─────────────────────┘               │                
                                      ▼                
                          yes         .                
           ┌─────────────────────────( ) limb          
           │                          '                
           ▼                                           
┌─────────────────────┐   create an instance of model  
│    DeviceCreated    │───────────────┐                
└─────────────────────┘               │                
                                      ▼                
                                      .                
                                     ( ) limb          
                                      '                
```

After the device instance is successfully created, the `limb` will use the `spec.adaptor.parameters` and `spec.template.spec` to connect that real device via adaptor:

```text
           │                                   
           ▼                                           
┌─────────────────────┐   create an instance of model  
│    DeviceCreated    │───────────────┐                
└─────────────────────┘               │                
                                      ▼                
                          success     .                
           ┌─────────────────────────( ) limb          
           │                          '                
           ▼                                           
┌─────────────────────┐   connect the real device      
│   DeviceConnected   │───────────────┐                
└─────────────────────┘               │                
                                      ▼                
                                      .                
                                     ( ) limb          
                                      '                
```

If the connection is healthy, the corresponding device instance will synchronize the status from the real one. This is the process of all state flow:

```text
┌─────────────────────┐   if the node is available?               
│     NodeExisted     │───────────────┐                           
└─────────────────────┘               │                           
                                      ▼                           
                          yes         .                           
           ┌─────────────────────────( ) brain                    
           │                          '                           
           ▼                                                      
┌─────────────────────┐   if the model is available?              
│    ModelExisted     │───────────────┐                           
└─────────────────────┘               │                           
                                      ▼                           
                          yes         .                           
           ┌─────────────────────────( ) brain                    
           │                          '                           
           ▼                                                      
┌─────────────────────┐   if the adaptor is available?            
│   AdaptorExisted    │───────────────┐                           
└─────────────────────┘               │                           
                                      ▼                           
                          yes         .                           
           ┌─────────────────────────( ) limb                     
           │                          '                           
           ▼                                                      
┌─────────────────────┐   create an instance of model             
│    DeviceCreated    │───────────────┐                           
└─────────────────────┘               │                           
                                      ▼                           
                          success     .                           
           ┌─────────────────────────( ) limb                     
           │                          '                           
           ▼                                                      
┌─────────────────────┐   connect the real device                 
│   DeviceConnected   │───────────────┐                           
└─────────────────────┘               │                           
           ▲                          ▼                           
           │              healthy     .                           
           └─────────────────────────( ) limb ─────────┐          
                                      '                │          
                                      ▲                ▼          
                                      │                .          
                                      └───────────────( ) adaptor 
                                                       '          
```

**The flow of states is serialized, which means that if the previous state is not ready(unsuccessful, false), it will not flow to the next state.**



## Correct behaviors

The main flow is not always going forward, some detection logic can adjust it to show the current state, we called them *corrections*. Some corrections are automatic, but some need manual intervention.

| State | Operator | Correction Logic |
|:---:|:---:|:---|
| `NodeExisted` | `brain` | If the Node has been deleted/drained/cordoned, the `brain` will adjust the main flow back to `NodeExisted` and mark it unavailable. <br/><br/> When the Node becomes available again, the `brain` will trigger the main flow to start again. |
| `ModelExisted` | `brain` | If the CRD(device model) has been deleted/disabled, the `brain` will adjust the main flow back to `ModelExisted` and mark it unavailable. <br/><br/> When the CRD becomes available again, the `brain` will trigger the main flow to start from model detection. |
| `AdaptorExisted` | `limb` | If the adaptor has been deleted, the `limb` will adjust the main flow back to `AdaptorExisted` and mark it unavailable. <br/><br/> When the adaptor becomes available again, the `limb` will trigger the main flow to start from adaptor detection. |
| `DeviceConnected` | `limb` maybe  | Accidental deletion of device instance will not be immediately perceived by `limb`, because the `limb` doesn't list-watch these instances. <br/><br/> If the deleted device has already connected(`DeviceConnected` was healthy), and the implementation of adapter is to synchronize status from the real device in real time or interval, it could have a chance to be recreated by `limb`. The `limb` will trigger the main flow to start from device creation again. <br/><br/> **Otherwise, the link needs to be modified/rebuilt manually.** |

### State of DeviceConnected 

Before talking about `DeviceConnected`, it needs to know that the device connection management of Octopus divides into two parts, one is the connection between `limb` and adaptor, another one is the connection between adaptor and real device:

```text
┌──────────┐   c1   ┌─────────┐   c2    .               
│   limb   │◀──────▶│ adaptor │◀──────▶( ) real device  
└──────────┘        └─────────┘         '                    
```

The `c1` is based on [gRPC](https://grpc.io/), but `c2` is determined by adaptor.  When both `c1` and `c2` are healthy, the `limb` will state the `DeviceConnected` in **Healthy**.

The `limb` can sense the changes in `c1`, if the `c1` closes unexpectedly, `limb` can trigger the main flow to start from device connection again.  

However, if the `c2` closes unexpectedly, the `limb` cannot perceive. The adaptor is responsible for notifying this, usually an *ERROR* will be sent to `limb`.  Then, the `limb` will state the `DeviceConnected` in **Unhealthy**.

If the interrupted `c2` is not reconnected, the link will remain **Unhealthy**. This depends on the adaptation of the `adapter`. Therefore, the implementation of adaptor should reconnect `c2` as much as possible. Otherwise, the situation should be made clear in the *ERROR* message.
