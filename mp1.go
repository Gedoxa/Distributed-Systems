package main

//import necessary packages.
import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Process struct represents a single process in the system.
// It has an ID, IP address, and a port.
type Process struct {
	ID   int    // Unique identifier for the process
	IP   string // IP address of the machine where the process is running
	Port string // Port on which the process is listening for connections
}

// Config struct represents the configuration of the system.
// It includes the minimum and maximum delay for sending messages,
// and a list of all processes in the system.
type Config struct {
	MinDelay  int       // Minimum delay for sending messages
	MaxDelay  int       // Maximum delay for sending messages
	Processes []Process // List of all processes in the system
}

// ParseConfig function reads a configuration file and returns a Config struct.
// The configuration file should have the following format:
// - The first line contains two integers, representing the minimum and maximum delay.
// - Each subsequent line represents a process, with the format: ID IP Port.
func ParseConfig(filename string) (*Config, error) {
	// Open the configuration file.
	file, err := os.Open(filename)
	if err != nil {
		return nil, err // Return an error if the file cannot be opened.
	}
	defer file.Close() // Ensure the file is closed when the function returns. As defer delays the activity until the function returns some value

	// Create a scanner to read the file line by line.
	scanner := bufio.NewScanner(file)
	scanner.Scan()                                     // Read the first line of the file.
	minMaxDelays := strings.Split(scanner.Text(), " ") // Split the first line into two parts.
	minDelay, _ := strconv.Atoi(minMaxDelays[0])       // Convert the first part to an integer.
	maxDelay, _ := strconv.Atoi(minMaxDelays[1])       // Convert the second part to an integer.

	// Create a new Config struct and set the minimum and maximum delay.
	config := &Config{
		MinDelay: minDelay,
		MaxDelay: maxDelay,
	}
	// Read the rest of the file line by line.
	for scanner.Scan() {
		processInfo := strings.Split(scanner.Text(), " ") // Split each line after every space, into three parts.
		processID, _ := strconv.Atoi(processInfo[0])      // Convert the first part to an integer.
		// Create a new Process struct and add it to the list of processes.
		process := Process{
			ID:   processID,
			IP:   processInfo[1],
			Port: processInfo[2],
		}
		config.Processes = append(config.Processes, process)
	}
	// Check for errors that occurred while reading the file.
	if err := scanner.Err(); err != nil {
		return nil, err // Return an error if there was a problem reading the file.
	}
	// Return the Config struct.
	return config, nil
}

// unicast_send function sends a message to a process through a network connection.
// The message is prefixed with the sender's process ID.
func unicast_send(conn net.Conn, processID int, message string) {
	// Write the process ID and the message to the connection.
	fmt.Fprintln(conn, strconv.Itoa(processID)+" "+message)
}

// unicast_send_with_delay function sends a message to a process with a delay.
// The delay is a random duration between the minimum and maximum delay specified in the configuration.
func unicast_send_with_delay(conn net.Conn, processID int, message string, delay time.Duration) {
	// Start a new goroutine to send the message after the delay.
	go func() {
		time.Sleep(delay)
		unicast_send(conn, processID, message)
	}()
}

// unicast_receive function listens for incoming messages from a process.
// When a message is received, it prints the message, the sender's process ID, and the current system time.
func unicast_receive(conn net.Conn, processID int) {
	scanner := bufio.NewScanner(conn)
	// Create a scanner to read the connection line by line.
	for scanner.Scan() { // Read each line.
		message := scanner.Text()                           // Get the message from the line.
		senderID := strings.Split(message, " ")[0]          // the sender ID is the first part of the message
		message = strings.TrimPrefix(message, senderID+" ") // Remove the sender ID from the message.
		//Print the received message, the sender's process ID, and the current system time.
		fmt.Println("Received message:", message, "from process", senderID, ", system time is:", time.Now())
	}
}

// startProcess function starts a process.
// It opens a network connection for each other process in the system,
// and starts a goroutine to listen for incoming messages from each other process.
// It also starts a goroutine to handle user input.
func startProcess(process Process, config *Config) {
	// Start listening for incoming network connections.
	ln, err := net.Listen("tcp", process.IP+":"+process.Port)
	if err != nil {
		log.Fatal(err) // Log an error and exit if there's a problem starting the listener.
	}
	defer ln.Close() // Ensure the listener is closed when the function returns. Defer waits until the completion of function.

	// Create a map to store the network connections to other processes.
	connections := make(map[int]net.Conn)
	for _, otherProcess := range config.Processes {
		if otherProcess.ID < process.ID {
			// Connect to each other process that has a lower ID.
			conn, err := net.Dial("tcp", otherProcess.IP+":"+otherProcess.Port)
			if err != nil {
				log.Fatal(err) // Log an error and exit if there's a problem connecting.
			}
			connections[otherProcess.ID] = conn       // Store the connection in the map.
			fmt.Fprintln(conn, process.ID)            // Send the process ID to the other process.
			go unicast_receive(conn, otherProcess.ID) // Start a goroutine to receive messages from the other process.
		}
	}
	// Start a goroutine to handle user input.
	go handleUserInput(process, connections, config.MinDelay, config.MaxDelay)
	// Loop indefinitely, accepting incoming connections.
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err) // Log an error and exit if there's a problem accepting a connection.
		}
		// Get the ID of the other process from the connection.
		otherID, err := getOtherID(conn)
		if err != nil {
			log.Fatal(err) // Log an error and exit if there's a problem getting the other process's ID.
		}
		// Store the connection in the map.
		connections[otherID] = conn
		go unicast_receive(conn, otherID)
	}
}

// getOtherID function reads the first line from a network connection,
// which should contain the ID of the other process.
func getOtherID(conn net.Conn) (int, error) {
	// Create a scanner to read the connection line by line.
	scanner := bufio.NewScanner(conn)
	scanner.Scan()              // read the first line
	firstLine := scanner.Text() // Get the first line.

	// Convert the first line to an integer.
	otherID, err := strconv.Atoi(firstLine)
	if err != nil {
		return 0, err // Return an error if there's a problem converting the line to an integer.
	}
	// Return the other process's ID.
	return otherID, nil
}

// handleUserInput function listens for user input.
// When the user enters a command, it parses the command and performs the appropriate action.
// The command format is: send [destinationID] [message]
func handleUserInput(process Process, connections map[int]net.Conn, minDelay int, maxDelay int) {
	// Create a scanner to read user input line by line.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() { // Read each line.
		command := strings.Split(scanner.Text(), " ") // Split the line into words.
		if len(command) < 2 {
			// Print an error message if the command is not in proper format.
			fmt.Println("Invalid command. Please enter a command in the format: send [destinationID] [message]")
			continue
		}
		//Handle the "send" command.
		if command[0] == "send" {
			// Convert the destination ID to an integer.
			destinationID, _ := strconv.Atoi(command[1])

			// Join the rest of the command into a message, even if it's an empty string.
			message := ""
			if len(command) > 2 {
				message = strings.Join(command[2:], " ")
			}

			// Check if the connection to the destination process exists
			conn, ok := connections[destinationID]
			if !ok {
				// Print an error message if the connection does not exist.
				fmt.Printf("No connection to process %d. Please check the configuration.\n", destinationID)
				continue
			}
			// Calculate a random delay for sending the message.
			delay := time.Duration(minDelay+rand.Intn(maxDelay-minDelay)) * time.Millisecond
			// Send the message with the delay.
			unicast_send_with_delay(conn, process.ID, message, delay)
			// Print a message indicating that the message was sent.
			fmt.Println("Sent message:", message, "to process:", destinationID, ", system time is:", time.Now())
		}
	}
}

// main function parses the configuration file and starts a goroutine for each process.
// Then it waits indefinitely.
func main() {
	// Parse the config file
	config, err := ParseConfig("config.txt")
	if err != nil {
		log.Fatal(err) // Log an error and exit if there's a problem parsing the configuration file.
	}

	// Start a goroutine for each process
	for _, process := range config.Processes {
		go startProcess(process, config)
	}

	// Wait indefinitely
	select {}
}
