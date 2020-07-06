# Modbus Adaptor

## Introduction

[Modbus](https://www.modbustools.com/modbus.html) is a master/slave protocol, the device requesting the information is called the Modbus master and the devices supplying information are Modbus slaves. 
In a standard Modbus network, there is one master and up to 247 slaves, each with a unique slave address from 1 to 247. 
The master can also write information to the slaves.

Modbus adaptor support both TCP and RTU protocol, it acting as the master node and connects to or manipulating the Modbus slave devices on the edge side.

## Documentation

Please see the [official docs](https://cnrancher.github.io/docs-octopus/docs/en/adaptors/modbus) site for complete documentation on Modbus Adaptor.
