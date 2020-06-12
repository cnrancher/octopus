# Modbus Adaptor

## Introduction

Adaptor的作用是连接和控制边缘设备与limb进行通信，Modbus Adaptor是对于使用modbus通信协议的设备的adaptor实现。
Modbus Adaptor支持TCP和RTU通信协议，用户构建DeviceLink CRD指定相应参数，并添加到集群内，即可接入设备，实现对设备属性的读写。

## Modbus Protocol

Modbus协议是一个master/slave架构的协议。
有一个节点是master节点，其他使用Modbus协议参与通信的节点是slave节点。
每一个slave设备都有一个唯一的地址。
所有设备都会收到命令，但只有指定位置的设备会执行及回应指令（地址0例外，指定地址0的指令是广播指令，所有收到指令的设备都会运行，不过不回应指令）。
在Modbus Adaptor中，adaptor作为master连接modbus slave设备。

所有的Modbus命令包含了检查码，以确定到达的命令没有被破坏。
基本的ModBus命令能指令一个RTU改变它的寄存器的某个值，控制或者读取一个I/O端口，以及指挥设备回送一个或者多个其寄存器中的数据。

## Registers Operation
Modbus设备有四种寄存器，可读可写的线圈寄存器（位操作），保持寄存器（字操作），和只读的离散输入寄存器（位操作），输入寄存器（字操作）。

**线圈寄存器**：可以类比为开关量，每一个bit对应一个信号的开关状态。所以一个byte可以同时控制8路的信号。比如控制外部8路io的高低。 线圈寄存器支持读也支持写。

**离散输入寄存器**：离散输入寄存器就相当于线圈寄存器的只读模式，也是每个bit表示一个开关量，而只能读取输入的开关信号，是不能写的。

**保持寄存器**：单位不再是bit而是两个byte，可以存放具体的数据量的，并且是可读写的。比如设置时间年月日，不但可以写也可以读出来现在的时间。

**输入寄存器**：和保持寄存器类似，但也只支持读而不能写。一个寄存器也是占据两个byte的空间。类比我我通过读取输入寄存器获取现在的AD采集值。  
   
Modbus功能码可以分为位操作和字操作两类。位操作的最小单位为BIT，字操作的最小单位为两个字节。

**位操作指令**:   读线圈状态01H，读(离散)输入状态02H，写单个线圈06H和写多个线圈0FH。

**字操作指令**:   读保持寄存器03H，写单个寄存器06H，写多个保持寄存器10H。


## DeviceLink CRD
定义设备链接（DeviceLink）
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
| ip | 设备的IP地址 | string
| port | 设备的端口 | int
| slaveId | 访问寄存器值时的标识字段 | int

#### RTU Config

| Parameter | Description | Type | Default |
|:--|:--|:--|:--|
| serialPort | 设备连接的串口，不同边缘节点操作系统下可选择不同的值。(e.g. /dev/ttyS0) | string |
| slaveId | 访问寄存器值时的标识字段 | int |
| baudRate | 每秒钟传送码元符号的个数，衡量数据传输速率的指标 | int | 19200 |
| dataBits | 衡量通信中实际数据位的参数 (5, 6, 7 or 8) | int | 8  |
| parity | 一种简单的校错方式，判断是否有噪声干扰通信或者是否存在传输和接收数据不同步 (N - None, E - Even, O - Odd) |string | E |
| stopBits | 用于表示单个数据包的最后一位 (1 or 2)|int| 1 |

### Property Visitor
| Parameter | Description | Type | 
|:--|:--|:--|
| register | 线圈寄存器 (CoilRegister)、离散输入寄存器 (DiscreteInputRegister)、保持寄存器 (HoldingRegister)或输入寄存器 (InputRegister)| string
| offset | 寄存器偏移地址 | int
| quantity | 寄存器的个数 | int