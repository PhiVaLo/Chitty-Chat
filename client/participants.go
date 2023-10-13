package main

import (
	"LogicalTime/proto"
	"bufio"
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
)

var global_id int = 0

type Client struct {
	id         int
	portNumber int
	timestamp  int
}

var (
	clientPort = flag.Int("cPort", 0, "client port number")
	serverPort = flag.Int("sPort", 0, "server port number (should match the port used for the server)")
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Each client has a unique id
	global_id++

	// Create a client
	client := &Client{
		id:         global_id,
		portNumber: *clientPort,
		timestamp:  0,
	}
	//serverConnection, _ := connectToServer(client)

	// Wait for the client (user) to send message
	go sendMessage(client)
	//go receiveMessage(client, serverConnection)

	// Keep the client running
	for {
	}
}

func connectToServer(client *Client) (proto.PublishClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	} else {
		//SHOULD BE SENT TO SERVER TO BROADCAST TO THE REST
		log.Printf("Participant %d joined the Chitty-chat server at port %d\n", client.id, *serverPort)

		//Update timestamp since participant joined
		client.timestamp++
	}

	return proto.NewPublishClient(conn), nil
}

func sendMessage(client *Client) {
	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)

	serverConnection, _ := connectToServer(client)

	for scanner.Scan() {
		input := scanner.Text()

		if len(input) > 128 {
			log.Printf("Message is too big (maximum length 128 characters)")
			continue //Starts the for loop from scratch
		}

		/*
			// To ensure that a string is a valid message with a maximum length of 128 characters
			if utf8.RuneCountInString(input) > 128 {
				log.Printf("Message is too big (maximum length 128 characters)")
				continue //Starts the for loop from scratch
			}
		*/

		//Sends message timestamp++
		client.timestamp++
		log.Printf("Participant sends the message: %s at the time %d \n", input, client.timestamp)

		// Ask the server to publish the message
		broadcastReturnMessage, err := serverConnection.AskForPublish(context.Background(), &proto.PublishMessage{
			ClientId:  int64(client.id),
			Timestamp: int64(client.timestamp),
			Message:   input,
		})

		if err != nil {
			log.Printf(err.Error())
		}

		//Client receives a message, so the timestamp is updated
		if client.timestamp < int(broadcastReturnMessage.Timestamp) {
			client.timestamp = int(broadcastReturnMessage.Timestamp)
		}
		client.timestamp++
	}
}

/*
func receiveMessage(client *Client, serverConnection *connectToServer) {
	//PRINT (Client received the message: *** at lamport timestamp)
}

func leaveChat() {
	//log.Printf("Participant %s left chitty-chat", client.id)
	//Broadcast to all clients that this person left
}

func waitForTimeRequest(client *Client) {
	// Connect to the server
	serverConnection, _ := connectToServer()

	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		log.Printf("Client asked for time with input: %s\n", input)

		// Ask the server for the time
		timeReturnMessage, err := serverConnection.AskForTime(context.Background(), &proto.AskForTimeMessage{
			ClientId: int64(client.id),
		})

		if err != nil {
			log.Printf(err.Error())
		} else {
			log.Printf("Server %s says the time is %s\n", timeReturnMessage.ServerName, timeReturnMessage.Time)
		}
	}
}
*/
