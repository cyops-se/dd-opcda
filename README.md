![example usage](./assets/cyops.png)
# dd-opcda - Simple OPC DA data one way replicator
**cyops-se**: *This application is part of the cyops.se community and use the same language and terminology. If there are acronyms or descriptions here that are unknown or ambiguous, please visit the [documentations](https://github.com/cyops-se/docs) site to see if it is explained there. You are welcome to help us improve the content regardless if you find what you are looking for or not*.

# Table of Contents
* [Introduction](#introduction)
* [Overview](#overview)
* [Example configuration (Windows)](#example-configuration)
* [User interface](#user-interface)

## Introduction
This very basic program that periodically collects tags from one or more OPC DA sources and sends them in JSON format over UDP to a single-cast receiver IP without expecting any response.

The primary use of this program is for replicating real time data from sensitive systems through a data diode to a potentially hostile network in order to achieve complete isolation of the sensitive network.

On the outer side of the data diode, use [dd-inserter](https://github.com/cyops-se/dd-inserter) to store the data in a tiemseries database or forward them through a message queue.

**IMPORTANT NOTE!**  
```dd-opcda``` is dependent on OPC core components (usually provided by the local OPC server) and the Grabox OPC wrapper. Please refer to the [CLI Examples](#cli-examples) section for more information! Without these pre-requisites, the application will fail!

## Overview
Today it is almost impossible to keep an IACS isolated over time as the businesses operating them find the information in them valuable and even critical to run the business efficiently. Fortunately, there are ways to meet that need without compromising the security architecture. One simple way is to use a data diode as illustrated below.

![example usage](./assets/diode-1.png)

The dd-opcda application can connect to one or more OPC DA servers and extract groups of tags with different sampling frequencies that are then sent over UDP to the receiver on the other end of the data diode (called end-point). It is also possible to send files and there is a local cache with process data (default 7 days), stored in 5 minutes intervals, that can be resend through the diode if necessary.

A web user interface is built into the application that provide an administrative interface for:
- browsing OPC DA server tag trees
- selection of tags to be replicated
- manage sampling groups
- manage end-points (the host running dd-inserter on the other side of the data diode)
- manage selected tags
- resend cached values
- transfer files
- see system logs
- manage users
- manage settings

A SQLite3 database is used to store application parameters and tag meta data. This database is created automatically and must be located in the current working directory of the application. If the application is configured as a Windows service, that directory is typically c:\windows\system32

***dd-opcda is currently only able to connect to local OPC servers***

Tag values are collected together with current time and quality and are sent in batches of 10 to avoid packet fragmentation.

## Example configuration
This example assumes you have a simple packet forwarding data diode which simply accepts packets at one port and mirrors it to other port without any possibility for data to go in the opposite direction. See [basic example](./EXAMPLE.md).

## User interface
A simple user interface is provided to configure, operate and monitor the application health. See [the user interface section](./USERINTERFACE.md) for more information.