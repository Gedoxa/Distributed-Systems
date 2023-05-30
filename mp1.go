package main

//import necessary packages.
import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
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

// UnicastMessage is the struct for passing messages between processes
// it includes the source id and it's corresponding messages
type UnicastMessage struct {
	SourceID int    //Source ID or Sender ID
	Message  string // Message from the sender
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
func unicast_send(encoder *gob.Encoder, sourceID int, message string) {
	//creating a new instance of UnicastMessage Struct
	msg := UnicastMessage{SourceID: sourceID, Message: message}
	//Encoding the msg object
	err := encoder.Encode(msg)
	if err != nil {
		log.Fatal(err)
	}
}

// unicast_send_with_delay function sends a message to a process with a delay.
// The delay is a random duration between the minimum and maximum delay specified in the configuration.
func unicast_send_with_delay(encoder *gob.Encoder, processID int, message string, delay time.Duration) {
	// Start a new goroutine to send the message after the delay.
	go func() {
		time.Sleep(delay)
		unicast_send(encoder, processID, message)
	}()
}

// unicast_receive function listens for incoming messages from a process.
func unicast_receive(decoder *gob.Decoder) {
	for {
		// Create a new UnicastMessage struct to store the incoming message
		msg := UnicastMessage{}
		//  decoding the incoming message
		err := decoder.Decode(&msg)

		if err != nil {
			log.Fatal(err)
		}
		// Print the received message, the sender's process ID, and the current time
		fmt.Printf("Received message: %s from process %d, system time is: %s\n", msg.Message, msg.SourceID, time.Now().Format(time.RFC3339))
	}
}

// startProcess function starts a process.
func startProcess(process Process, config *Config) {
	// initialize a wait group to sync multiple goroutines
	var wg sync.WaitGroup
	// Create a map to store gob.Encoder objects for each connection
	connMap := make(map[int]*gob.Encoder)

	// Server side
	go func() {
		// Start listening for incoming connections
		ln, _ := net.Listen("tcp", ":"+process.Port)
		for {
			// Accept an incoming connection
			conn, _ := ln.Accept()
			// Create a new gob.Decoder for the connection
			decoder := gob.NewDecoder(conn)
			// Increment the wait group counter
			wg.Add(1)
			// Start a new goroutine
			go func() {
				unicast_receive(decoder)
				// Decrement the counter when the goroutine completes
				wg.Done()
			}()
		}
	}()

	// Client side
	for _, otherProcess := range config.Processes {
		if otherProcess.ID != process.ID {
			var conn net.Conn
			var err error
			retries := 5
			// Try to establish the connection
			for i := 0; i < retries; i++ {
				// Dial the other process
				conn, err = net.Dial("tcp", otherProcess.IP+":"+otherProcess.Port)
				if err == nil { // If the connection is successful, break the loop
					break
				}
				// If the connection is not successful, wait for a period and retry
				time.Sleep(time.Second * time.Duration(i+1))
			}
			// If the connection is still not successful after all retries, log the error
			if err != nil {
				log.Fatal(err)
			}
			// Close the connection when the function returns
			defer conn.Close()
			// Create a new gob.Encoder for the connection and store it in the map
			connMap[otherProcess.ID] = gob.NewEncoder(conn)
		}
	}

	// Start a goroutine to handle user input
	go handleUserInput(process, connMap, config.MinDelay, config.MaxDelay)
	// Wait for all goroutines to complete
	wg.Wait()
}

// handleUserInput function listens for user input.
func handleUserInput(process Process, connections map[int]*gob.Encoder, minDelay int, maxDelay int) {
	scanner := bufio.NewScanner(os.Stdin)
	// Continuously read from input
	for scanner.Scan() {
		// Split the input into words
		command := strings.Split(scanner.Text(), " ")
		if command[0] == "send" && len(command) > 1 {
			// convert the second word to an integer
			destinationID, err := strconv.Atoi(command[1])
			if err == nil {
				// Check if there is a connection to the destination process
				if encoder, ok := connections[destinationID]; ok {
					message := strings.Join(command[2:], " ")
					// Calculate a random delay within the specified range
					delay := time.Duration(minDelay+rand.Intn(maxDelay-minDelay)) * time.Millisecond
					// Send the message to the destination process after the delay
					unicast_send_with_delay(encoder, process.ID, message, delay)
					fmt.Printf("Sent message: %s to process %d, system time is: %s\n", message, destinationID, time.Now().Format(time.RFC3339))
				} else {
					fmt.Printf("Invalid destination process ID: %d\n", destinationID)
				}
			} else {
				fmt.Println("Invalid command format. Use: send [destinationID] [message]")
			}
		} else {
			fmt.Println("Invalid command format. Use: send [destinationID] [message]")
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
