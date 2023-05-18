The key components of the program are as follows:

## Process Struct:

This ‘struct’ represents a process in the system. Each process is identified by an ID and has an associated IP and port for communication.

## Config Struct:

This ‘struct’ represents the configuration of the system, which includes the minimum and maximum delay for message delivery, and a list of processes.

## Reading_Config Function:

This function reads a configuration file, parses it, and returns a Config struct. The configuration file should list the min and max delay on the first line, followed by a list of processes, each on its own line. Each process line should contain the ID, IP, and port, separated by spaces.

## unicast_send and unicast_send_with_delay Functions:

These are helper functions for sending messages. The former sends a message immediately, while the latter sends a message after a delay.

## unicast_receive Function:

This function listens for incoming messages on a network connection.

## startProcess Function:

This function is the heart of the process simulation. It opens a network connection for the process and establishes connections to other processes. It also starts a goroutine to handle user input and another to listen for incoming messages.

## getOtherID Function:

This function retrieves the ID of the other process from a network connection.
handleUserInput Function: This function handles user input from the console. Users can input commands in the form of send [destinationID] [message] to send a message to a specific process.

## main Function:

This is the entry point of the program. It reads the config file, starts a goroutine for each process, and then waits forever.

## In terms of the flow of the code,

The main function starts the program by reading the configuration file and starting a separate goroutine for each process. Each process sets up its network connection and connects to other processes, then starts listening for user input and incoming messages. When a user input command is received, a message is sent to the specified process with a delay. When an incoming message is received, it is printed to the console along with the process id, system time & the message.
