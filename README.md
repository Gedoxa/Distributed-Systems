# A simple network simulation

This project simulates a distributed system using TCP/IP for communication between processes. Each process can send and receive messages to/from other processes with a specified delay.

## Structure

The project consists of a single Go file that contains all the necessary functions to simulate a simple network system. The system configuration is read from a text file.

## Configuration

The system configuration is specified in a text file named `config.txt`. The first line of this file specifies the minimum and maximum delay for sending messages (in milliseconds). Each subsequent line represents a process in the system, with the format: `ID IP Port`.

Here's an example configuration:

```
100 200
1 127.0.0.1 8001
2 127.0.0.1 8002
3 127.0.0.1 8003
4 127.0.0.1 8004
```

This configuration specifies a system with 4 processes. The minimum delay for sending messages is 100 milliseconds, and the maximum delay is 200 milliseconds. The processes have IDs 1 through 4, and they all run on the local machine (127.0.0.1), with ports 8001 through 8004.

## Usage

To run the simulation, simply execute the Go file:

```bash
go run mp1.go

send [destinationID] [message]

send 2 Hello, world!

```
